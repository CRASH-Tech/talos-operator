package talos

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"time"

	kubernetes "github.com/CRASH-Tech/talos-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"google.golang.org/protobuf/types/known/durationpb"
)

func newClient(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig) (*client.Client, error) {
	tCtx := clientconfig.Context{
		CA:  machineConfig.MachineSecrets.CA,
		Crt: machineConfig.MachineSecrets.Crt,
		Key: machineConfig.MachineSecrets.Key,
	}

	configContext := client.WithConfigContext(&tCtx)
	configEndpoints := client.WithEndpoints(endpoint)

	cert, err := base64.StdEncoding.DecodeString(machineConfig.MachineSecrets.Crt)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(machineConfig.MachineSecrets.Key)
	if err != nil {
		return nil, err
	}

	xCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{xCert},
		ClientAuth:         tls.NoClientCert,
	}

	configTls := client.WithTLSConfig(tlsConfig)

	//configTimeout := client.WithGRPCDialOptions(grpc.WithTimeout(time.Second * time.Duration(timeout)))

	client, err := client.New(ctx, configTls, configContext, configEndpoints)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func Bootstrap(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig) error {
	client, err := newClient(ctx, endpoint, machineConfig)
	if err != nil {
		return err
	}

	defer client.Close()

	req := machine.BootstrapRequest{
		RecoverEtcd:          false,
		RecoverSkipHashCheck: true,
	}

	err = client.Bootstrap(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}

func ApplyConfiguration(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig, mode machine.ApplyConfigurationRequest_Mode) (machine.ApplyConfigurationResponse, error) {
	client, err := newClient(ctx, endpoint, machineConfig)
	if err != nil {
		return machine.ApplyConfigurationResponse{}, err
	}

	defer client.Close()

	req := machine.ApplyConfigurationRequest{
		Mode:           mode,
		Data:           []byte(machineConfig.MachineConfig),
		TryModeTimeout: durationpb.New(time.Minute * 5),
	}

	resp, err := client.ApplyConfiguration(ctx, &req)
	if err != nil {
		return machine.ApplyConfigurationResponse{}, err
	}

	return *resp, nil
}

func Check(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig) bool {
	client, err := newClient(ctx, endpoint, machineConfig)
	if err != nil {
		return false
	}

	defer client.Close()

	services, err := client.ServiceList(ctx)
	if err != nil {
		return false
	}

	for _, msg := range services.Messages {
		for _, service := range msg.Services {
			if service.Id == "kubelet" {
				return service.Health.Healthy
			}
		}
	}

	return true
}

func Reset(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig) error {
	client, err := newClient(ctx, endpoint, machineConfig)
	if err != nil {
		return err
	}

	defer client.Close()

	err = client.Reset(ctx, true, false)
	if err != nil {
		return err
	}

	return nil
}

func ServicesStatus(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig, machineStatus v1alpha1.MachineStatus) (v1alpha1.MachineStatus, error) {
	client, err := newClient(ctx, endpoint, machineConfig)
	if err != nil {
		return machineStatus, err
	}

	defer client.Close()

	services, err := client.ServiceList(ctx)
	if err != nil {
		return machineStatus, err
	}

	for _, msg := range services.Messages {
		for _, service := range msg.Services {
			var health string

			if service.Health.Healthy {
				health = "Healthy"
			} else {
				health = "Unhealthy"
			}
			status := fmt.Sprintf("%s/%s", service.State, health)
			switch service.Id {
			case "etcd":
				machineStatus.Etcd = status
			case "apid":
				machineStatus.Apid = status
			case "kubelet":
				machineStatus.Kubelet = status
			case "containerd":
				machineStatus.Containerd = status
			case "cri":
				machineStatus.Cri = status
			case "machined":
				machineStatus.Machined = status
			default:
			}
		}
	}

	return machineStatus, nil
}
