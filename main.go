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
	"time"

	"github.com/CRASH-Tech/talos-operator/cmd/common"
	kubernetes "github.com/CRASH-Tech/talos-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	"github.com/CRASH-Tech/talos-operator/cmd/talos"
	talosMachine "github.com/siderolabs/talos/pkg/machinery/api/machine"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version = "0.0.1"
	config  common.Config
	kClient *kubernetes.Client
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
}

func main() {
	log.Infof("Starting talos-operator %s", version)

	ctx := context.Background()
	kClient = kubernetes.NewClient(ctx, *config.DynamicClient, *config.KubernetesClient)

	listen()

	for {
		processV1aplha1MachineSelectors(kClient)
		//processV1aplha1Machines(kClient)

		time.Sleep(5 * time.Second)
	}
}

func readConfig(path string) (common.Config, error) {
	config := common.Config{}

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return common.Config{}, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return common.Config{}, err
	}

	return config, err
}

func listen() {
	go func() {
		http.HandleFunc("/register", registerHandler)
		err := http.ListenAndServe(config.Listen, nil)
		if err != nil {
			log.Panic(err)
		}
	}()
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.RemoteAddr, ":")[0]
	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		params[k] = strings.Join(v, ",")
	}

	log.Infof("Register query received: %s %s", host, params)

	pMachine := v1alpha1.PendingMachine{}

	pMachine.Metadata.Name = host
	pMachine.Spec.Host = host
	pMachine.Spec.Params = append(pMachine.Spec.Params, v1alpha1.PendingMachineParams{Key: "host", Value: host})
	for k, v := range params {
		if v != "" {
			p := v1alpha1.PendingMachineParams{
				Key:   k,
				Value: v,
			}
			pMachine.Spec.Params = append(pMachine.Spec.Params, p)
		}
	}

	//CHECK IS ALREADY EXISTS
	machines, err := kClient.V1alpha1().Machine().GetAll()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error!"))

		return
	}
	for _, machine := range machines {
		if machine.Spec.Host == pMachine.Spec.Host {
			log.Errorf("Received register query with already exists host: %s", pMachine.Spec.Host)
			w.WriteHeader(http.StatusAlreadyReported)
			w.Write([]byte("Already exists!"))

			return
		}
	}

	result, err := kClient.V1alpha1().PendingMachine().Create(pMachine)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	log.Infof("Registered new machine: %s", result)
}

func processV1aplha1MachineSelectors(kClient *kubernetes.Client) {
	log.Info("Processing v1alpha1 machine selectors...")
	//ctx := context.Background()

	selectors, err := kClient.V1alpha1().MachineSelector().GetAll()
	if err != nil {
		log.Error(err)

		return
	}

	pMachines, err := kClient.V1alpha1().PendingMachine().GetAll()
	if err != nil {
		log.Error(err)

		return
	}

	for _, selector := range selectors {
		for _, sParams := range selector.Spec.Params {
			
		}
	}

}

func processV1aplha1Machines(kClient *kubernetes.Client) {
	log.Info("Processing v1alpha1 machines...")
	ctx := context.Background()

	ns := "talos-operator"

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
		machineConfig, err := kClient.GetMachineConfig(machine.Spec.Config, ns)
		if err != nil {
			log.Error(err)
			continue
		}

		///
		// machine.Status.LastApplyFail = false
		// _, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
		// continue
		///

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
					_, err = kClient.V1alpha1().Machine().Patch(machine)
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
								machine.Status.LastApplyFail = true
								_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
								if err != nil {
									log.Error(err)
									continue
								}
								continue
							}

							log.Infof("Check is machine alive %s(%s)", machine.Metadata.Name, machine.Spec.Host)
							time.Sleep(time.Second * 60)
							if !talos.Check(ctx, machine.Spec.Host, machineConfig) {
								log.Warnf("Machine %s(%s) health check fail", machine.Metadata.Name, machine.Spec.Host)
								machine.Status.LastApplyFail = true
								_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
								if err != nil {
									log.Error(err)
									continue
								}
								continue
							}
						}

						//APPLY
						log.Infof("Apply new config to %s(%s) mode: %s", machine.Metadata.Name, machine.Spec.Host, mode)
						_, err := talos.ApplyConfiguration(ctx, machine.Spec.Host, machineConfig, mode)
						if err != nil {
							log.Error(err)
							machine.Status.LastApplyFail = true
							_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
							if err != nil {
								log.Error(err)
								continue
							}
							continue
						}

						machine.Status.ConfigHash = newHash
						_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
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
					_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
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
		servicesStatus, err := talos.ServicesStatus(ctx, machine.Spec.Host, machineConfig, machine.Status)
		if err != nil {
			log.Error(err)
		}

		machine.Status = servicesStatus
		_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
		if err != nil {
			log.Error(err)
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
