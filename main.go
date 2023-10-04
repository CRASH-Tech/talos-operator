package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/CRASH-Tech/talos-operator/cmd/common"
	kubernetes "github.com/CRASH-Tech/talos-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api"
	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	"github.com/CRASH-Tech/talos-operator/cmd/talos"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	talosMachine "github.com/siderolabs/talos/pkg/machinery/api/machine"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version       = "0.0.3"
	config        common.Config
	kClient       *kubernetes.Client
	namespace     string
	hostname      string
	mutex         sync.Mutex
	leaseLockName = "talos-operator"

	machineStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "machine_status",
			Help: "The talos machine service status",
		},
		[]string{
			"host",
			"config",
			"service",
		},
	)
)

func init() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "config file path. Default: config.yaml")
	c, err := readConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	config = c

	switch config.Log.Format {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	switch config.Log.Level {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	var restConfig *rest.Config
	if path, isSet := os.LookupEnv("KUBECONFIG"); isSet {
		log.Printf("Using configuration from '%s'", path)
		restConfig, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("Using in-cluster configuration")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	config.DynamicClient = dynamic.NewForConfigOrDie(restConfig)
	config.KubernetesClient = k8s.NewForConfigOrDie(restConfig)

	ns, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Panic(err)
	}

	namespace = string(ns)
	hostname = os.Getenv("HOSTNAME")

	prometheus.MustRegister(machineStatus)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kClient = kubernetes.NewClient(ctx, *config.DynamicClient, *config.KubernetesClient)

	mutex.Lock()
	setLeaderLabel(false)

	lock := getNewLock(leaseLockName, hostname, namespace)
	runLeaderElection(lock, ctx, hostname)
}

func worker() {
	log.Infof("Starting talos-operator %s", version)

	listen()

	for {
		processV1aplha1Machines(kClient)
		metrics()
		time.Sleep(5 * time.Second)
	}
}

func listen() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/register", registerHandler)
		err := http.ListenAndServe(config.Listen, nil)
		if err != nil {
			log.Panic(err)
		}
	}()
}

func healthConverter(s string) float64 {
	switch s {
	case "Running/Healthy":
		return 1
	case "Running/Unhealthy":
		return 0
	default:
		return -1
	}
}

func metrics() {
	machines, err := kClient.V1alpha1().Machine().GetAll()
	if err != nil {
		log.Error(err)

		return
	}

	for _, machine := range machines {
		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"etcd",
		).Set(healthConverter(machine.Status.Etcd))

		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"apid",
		).Set(healthConverter(machine.Status.Apid))

		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"containerd",
		).Set(healthConverter(machine.Status.Containerd))

		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"cri",
		).Set(healthConverter(machine.Status.Cri))

		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"kubelet",
		).Set(healthConverter(machine.Status.Kubelet))

		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"machined",
		).Set(healthConverter(machine.Status.Machined))

		var confifOK float64
		if machine.Status.LastApplyFail {
			confifOK = 0
		} else {
			confifOK = 1
		}
		machineStatus.WithLabelValues(
			machine.Spec.Host,
			machine.Spec.Config,
			"config",
		).Set(confifOK)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	host := strings.Split(r.RemoteAddr, ":")[0]
	params := make(map[string]string)
	params["host"] = host

	for k, v := range r.URL.Query() {
		params[k] = strings.Join(v, ",")
	}

	log.Infof("Register query received: %s %s", host, params)

	machines, err := kClient.V1alpha1().Machine().GetAll()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error!"))

		return
	}
	//CHECK IS ALREADY EXISTS
	for _, machine := range machines {
		if machine.Spec.Host == host {
			log.Errorf("Received register query with already exists host: %s", host)
			w.WriteHeader(http.StatusAlreadyReported)
			w.Write([]byte("Already exists!"))

			return
		}
	}

	selectors, err := kClient.V1alpha1().MachineSelector().GetAll()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error!"))

		return
	}

	for _, selector := range selectors {
		fail := false
		for _, sParam := range selector.Spec.Params {
			match, err := regexMatch(sParam.Value, params[sParam.Key])
			if err != nil {
				log.Error(err)
				fail = true
			}

			if !match {
				fail = true
			}
		}

		if fail {
			continue
		}

		ps := []v1alpha1.MachineParams{}
		for k, v := range params {
			if v != "" {
				p := v1alpha1.MachineParams{
					Key:   k,
					Value: v,
				}
				ps = append(ps, p)
			}
		}

		ownerReference := api.CustomResourceOwnerReference{
			ApiVersion:         selector.APIVersion,
			Kind:               selector.Kind,
			Name:               selector.Metadata.Name,
			Uid:                selector.Metadata.Uid,
			BlockOwnerDeletion: true,
		}

		//FIND SELECTOR MACHINES
		var selectorMachines []v1alpha1.Machine
		for _, machine := range machines {
			for _, or := range machine.Metadata.OwnerReferences {
				if selector.Metadata.Uid == or.Uid {
					selectorMachines = append(selectorMachines, machine)
				}
			}
		}

		var bootstrap bool
		if selector.Spec.Bootstrap && len(selectorMachines) == 0 {
			bootstrap = true
		} else {
			bootstrap = params["bootstrap"] == "true"
		}

		machine := v1alpha1.Machine{
			Metadata: api.CustomResourceMetadata{
				Name:            host,
				Finalizers:      []string{"resources-finalizer.talos-operator.xfix.org"},
				OwnerReferences: []api.CustomResourceOwnerReference{ownerReference},
			},
			Spec: v1alpha1.MachineSpec{
				Host:      host,
				Config:    selector.Spec.Config,
				Bootstrap: bootstrap,
				Params:    ps,
			},
		}

		machineConfig, err := kClient.GetMachineConfig(selector.Spec.Config, namespace)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal error!"))

			return
		}

		machine, err = kClient.V1alpha1().Machine().Create(machine)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal error!"))

			return
		}

		log.Infof("Send machineconfig %s for: %s", selector.Spec.Config, host)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(machineConfig.MachineConfig))

		return

	}

	log.Warnf("No selector found for: %s", host)
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("No selector found!"))

}

func processV1aplha1Machines(kClient *kubernetes.Client) {
	log.Info("Processing v1alpha1 machines...")
	ctx := context.Background()

	machines, err := kClient.V1alpha1().Machine().GetAll()
	if err != nil {
		log.Error(err)

		return
	}

	var roMode bool
	for _, machine := range machines {
		//CHECK FOR FAILLED HOSTS
		if machine.Status.LastApplyFail {
			log.Warn("One or more machines have failed apply config. Working in readonly mode")
			roMode = true

			break
		}
		//CHECK FOR SAME HOST
		for _, m := range machines {
			if m.Metadata.Name != machine.Metadata.Name && m.Spec.Host == machine.Spec.Host {
				log.Warn("One or more machines have same host. Working in readonly mode")
				roMode = true
			}
		}
	}

	for _, machine := range machines {
		machineConfig, err := kClient.GetMachineConfig(machine.Spec.Config, namespace)
		if err != nil {
			log.Error(err)

			continue
		}

		if !roMode {
			//PROTECTION CHECK
			if !machine.Spec.Protected {
				//MACHINE DELETION
				if machine.Metadata.DeletionTimestamp != "" {
					log.Infof("Deleting machine %s(%s)", machine.Metadata.Name, machine.Spec.Host)
					err := talos.Reset(ctx, machine.Spec.Host, machineConfig)
					if err != nil {
						log.Error(err)

						continue
					}

					machine.Metadata.Finalizers = []string{}
					machine, err = kClient.V1alpha1().Machine().Patch(machine)
					if err != nil {
						log.Error(err)
					}

					continue
				}

				//APPLY CONFIG
				newHashB := md5.Sum([]byte(machineConfig.MachineConfig))
				newHash := hex.EncodeToString(newHashB[:])

				if newHash != machine.Status.ConfigHash {
					if !machine.Status.LastApplyFail {
						var mode talosMachine.ApplyConfigurationRequest_Mode
						if machine.Status.ConfigHash == "" {
							mode = talosMachine.ApplyConfigurationRequest_AUTO
						} else {
							mode = talosMachine.ApplyConfigurationRequest_NO_REBOOT
						}

						//TRY
						if mode == talosMachine.ApplyConfigurationRequest_NO_REBOOT {
							tryMode := talosMachine.ApplyConfigurationRequest_TRY
							log.Infof("Trying new config to %s(%s) mode: %s", machine.Metadata.Name, machine.Spec.Host, tryMode)
							_, err := talos.ApplyConfiguration(ctx, machine.Spec.Host, machineConfig, tryMode)
							if err != nil {
								log.Error(err)

								continue
							}

							log.Infof("Check is machine alive %s(%s)", machine.Metadata.Name, machine.Spec.Host)
							time.Sleep(time.Second * 60)
							if !talos.Check(ctx, machine.Spec.Host, machineConfig) {
								log.Warnf("Machine %s(%s) health check fail", machine.Metadata.Name, machine.Spec.Host)
								machine.Status.LastApplyFail = true
								machine, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
								if err != nil {
									log.Error(err)

									return
								}

								return
							}
						}

						//APPLY
						log.Infof("Apply new config to %s(%s) mode: %s", machine.Metadata.Name, machine.Spec.Host, mode)
						_, err := talos.ApplyConfiguration(ctx, machine.Spec.Host, machineConfig, mode)
						if err != nil {
							log.Error(err)

							continue
						}

						machine.Status.ConfigHash = newHash
						machine, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
						if err != nil {
							log.Error(err)

							continue
						}
					} else {
						log.Warnf("Last apply config to %s(%s) fail. Ignoring it", machine.Metadata.Name, machine.Spec.Host)

						continue
					}

				}

				//BOOTSTRAP
				if machine.Status.ConfigHash != "" && machine.Spec.Bootstrap && !machine.Status.Bootstrapped {
					log.Infof("Bootstrap %s(%s)", machine.Metadata.Name, machine.Spec.Host)
					err := talos.Bootstrap(ctx, machine.Spec.Host, machineConfig)
					if err != nil {
						log.Error(err)
						if !strings.Contains(err.Error(), "AlreadyExists") {

							continue
						}
					}

					machine.Status.Bootstrapped = true
					machine, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
					if err != nil {
						log.Error(err)

						continue
					}
				}
			} else {
				log.Warningf("Machine %s(%s) is protected. Ignore it.", machine.Metadata.Name, machine.Spec.Host)
			}
		}

		//SERVICES
		if machine.Status.ConfigHash != "" {
			servicesStatus, err := talos.ServicesStatus(ctx, machine.Spec.Host, machineConfig, machine.Status)
			if err != nil {
				log.Warnf("Cannot get services status %s: %s", machine.Spec.Host, err)

				continue
			}

			machine.Status = servicesStatus
			machine, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
			if err != nil {
				log.Error(err)

				continue
			}
		}

	}
}

func regexMatch(regex, value string) (bool, error) {
	match, err := regexp.MatchString(regex, value)
	if err != nil {
		return false, err
	}

	return match, nil
}
