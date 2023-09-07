package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/CRASH-Tech/talos-operator/cmd/common"
	kubernetes "github.com/CRASH-Tech/talos-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/dynamic"
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
		log.Printf("Using in-cluster configuration")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	config.DynamicClient = dynamic.NewForConfigOrDie(restConfig)
}

func main() {
	log.Infof("Starting talos-operator %s", version)

	ctx := context.Background()
	kClient = kubernetes.NewClient(ctx, *config.DynamicClient)

	err := listen()
	if err != nil {
		log.Panic(err)
	}

	//pClient := proxmox.NewClient(config.Clusters)

	// for {
	// 	processV1aplha1(kClient)

	// 	time.Sleep(5 * time.Second)
	// }
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

func listen() error {
	http.HandleFunc("/register", registerHandler)
	err := http.ListenAndServe(config.Listen, nil)
	if err != nil {
		return err
	}

	return nil
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
	machine.Spec.Allocated = false

	for k, v := range params {
		p := v1alpha1.MachineParams{
			Key:   k,
			Value: v,
		}
		machine.Spec.Params = append(machine.Spec.Params, p)
	}

	result, err := kClient.V1alpha1().Machine().New(machine)
	if err != nil {
		return err
	}

	log.Debug("registered new machine: ", result)

	return nil

}


// func processV1aplha1(kClient *kubernetes.Client) {
// 	log.Info("Refreshing v1alpha1...")
// 	qemus, err := kClient.V1alpha1().Qemu().GetAll()
// 	if err != nil {
// 		log.Error(err)
// 		return
// 	}

// 	for _, qemu := range qemus {
// 		switch qemu.Status.Status {
// 		case v1alpha1.STATUS_QEMU_EMPTY:
// 			if qemu.Status.Status == v1alpha1.STATUS_QEMU_EMPTY && qemu.Metadata.DeletionTimestamp != "" {
// 				qemu.RemoveFinalizers()
// 				_, err = kClient.V1alpha1().Qemu().Patch(qemu)
// 				if err != nil {
// 					log.Errorf("cannot patch qemu cr %s: %s", qemu.Metadata.Name, err)

// 					continue
// 				}

// 				continue
// 			}

// 			if qemu.Spec.Clone != "" {
// 				qemu.Status.Status = v1alpha1.STATUS_QEMU_CLONING
// 				qemu, err = updateQemuStatus(kClient, qemu)
// 				if err != nil {
// 					return
// 				}

// 				continue
// 			}

// 			qemu, err := getQemuPlace(pClient, qemu)
// 			if err != nil {
// 				log.Errorf("cannot get qemu place %s: %s", qemu.Metadata.Name, err)

// 				continue
// 			}

// 			if qemu.Status.Status == v1alpha1.STATUS_QEMU_OUT_OF_SYNC {
// 				qemu, err = updateQemuStatus(kClient, qemu)
// 				if err != nil {
// 					return
// 				}

// 				continue
// 			}

// 			qemu, err = createNewQemu(pClient, qemu)
// 			if err != nil {
// 				log.Errorf("cannot create qemu %s: %s", qemu.Metadata.Name, err)
// 				if qemu.Status.Status == v1alpha1.STATUS_QEMU_EMPTY {
// 					qemu = cleanQemuPlaceStatus(qemu)
// 				}

// 				qemu, err = updateQemuStatus(kClient, qemu)
// 				if err != nil {
// 					return
// 				}

// 				continue
// 			}

// 			qemu.Status.Status = v1alpha1.STATUS_QEMU_SYNCED
// 			qemu, err = updateQemuStatus(kClient, qemu)
// 			if err != nil {
// 				return
// 			}

// 			// Need by proxmox api delay
// 			time.Sleep(time.Second * 10)

// 			continue

// 			default:
// 				qemu, err = checkQemuSyncStatus(pClient, qemu)
// 				if err != nil {
// 					log.Errorf("cannot get qemu sync status %s: %s", qemu.Metadata.Name, err)
// 					qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
// 					qemu, err = updateQemuStatus(kClient, qemu)
// 					if err != nil {
// 						return
// 					}

// 					continue
// 				}

// 				qemu, err = updateQemuStatus(kClient, qemu)
// 				if err != nil {
// 					return
// 				}

// 				continue
// 			}
// 		default:
// 			log.Warnf("unknown qemu state: %s %s", qemu.Metadata.Name, qemu.Status.Status)

// 			continue

// 	}

// }
