package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
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
		processV1aplha1(kClient)

		time.Sleep(5 * time.Second)
	}
}

func readConfig(path string) (common.Config, error) {
	config := common.Config{}
	//config.Clusters = make(map[string]proxmox.ClusterApiConfig)

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
	host := strings.Split(r.Host, ":")[0]
	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		params[k] = strings.Join(v, ",")
	}

	log.Debug("register query received: ", host, params)

	err := CreateNewMachine(host, params)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusAlreadyReported)
		w.Write([]byte("Error!"))
	}
}

func CreateNewMachine(host string, params map[string]string) error {
	machine := v1alpha1.Machine{}

	machine.Metadata.Name = host
	machine.Spec.Host = host
	machine.Spec.Bootstrap = false

	for k, v := range params {
		p := v1alpha1.MachineParams{
			Key:   k,
			Value: v,
		}
		machine.Spec.Params = append(machine.Spec.Params, p)
	}

	result, err := kClient.V1alpha1().Machine().Create(machine)
	if err != nil {
		return err
	}

	log.Debug("registered new machine: ", result)

	return nil
}

func processV1aplha1(kClient *kubernetes.Client) {
	log.Info("Refreshing v1alpha1...")
	ctx := context.Background()

	ns := "talos-operator"

	machines, err := kClient.V1alpha1().Machine().GetAll()
	if err != nil {
		log.Error(err)
	}

	for _, machine := range machines {
		if machine.Spec.Protected {
			log.Infof("Machine %s is protected. Ignore it.", machine.Metadata.Name)
			continue
		}

		machineConfig, err := kClient.GetMachineConfig(machine.Spec.Config, ns)
		if err != nil {
			log.Error(err)
			continue
		}

		newHashB := md5.Sum([]byte(machineConfig.MachineConfig))
		newHash := hex.EncodeToString(newHashB[:])

		if newHash != machine.Status.ConfigHash {
			var mode talosMachine.ApplyConfigurationRequest_Mode
			if machine.Status.ConfigHash == "" {
				mode = talosMachine.ApplyConfigurationRequest_AUTO
			} else {
				mode = talosMachine.ApplyConfigurationRequest_NO_REBOOT
			}
			log.Infof("Apply new config to %s: %s mode: %s", machine.Metadata.Name, machine.Spec.Host, mode)
			_, err := talos.ApplyConfiguration(ctx, machine.Spec.Host, machineConfig, mode)
			if err != nil {
				log.Error(err)
				continue
			}

			machine.Status.ConfigHash = newHash
			_, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
			if err != nil {
				log.Error(err)
				continue
			}
		}

		if machine.Status.ConfigHash != "" && machine.Spec.Bootstrap && !machine.Status.Bootstrapped {
			log.Infof("Bootstrap %s:%s", machine.Metadata.Name, machine.Spec.Host)
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
	}

}

// machines, err := kClient.V1alpha1().Machine().GetAll()
// if err != nil {
// 	panic(err)
// }
// log.Debug(machines)

// machine, err := kClient.V1alpha1().Machine().Get("k-test-m1")
// if err != nil {
// 	panic(err)
// }

// machine.Status.Bootstrapped = false

// _, err = kClient.V1alpha1().Machine().UpdateStatus(machine)
// if err != nil {
// 	panic(err)
// }

// machineConfigs, err := kClient.GetMachineConfigs("talos-operator")
// if err != nil {
// 	log.Error(err)
// }

// for _, machineConfig := range machineConfigs {
// 	ctx := context.Background()
// 	//talos.ApplyConfiguration(ctx, "10.171.120.151", machineConfig)
// 	talos.Bootstrap(ctx, "10.171.120.151", machineConfig)
// 	os.Exit(0)
// }
