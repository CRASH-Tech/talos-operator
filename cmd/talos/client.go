package talos

import (
	"context"
	"crypto/tls"
	"encoding/base64"

	kubernetes "github.com/CRASH-Tech/talos-operator/cmd/kubernetes"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
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

	req := machine.ApplyConfigurationRequest{
		Mode: mode,
		Data: []byte(machineConfig.MachineConfig),
	}

	resp, err := client.ApplyConfiguration(ctx, &req)
	if err != nil {
		return machine.ApplyConfigurationResponse{}, err
	}

	return *resp, nil
}

func Reset(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig) error {
	client, err := newClient(ctx, endpoint, machineConfig)
	if err != nil {
		return err
	}

	err = client.Reset(ctx, true, false)
	if err != nil {
		return err
	}

	return nil
}
