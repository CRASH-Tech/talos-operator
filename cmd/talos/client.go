package talos

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"os"

	kubernetes "github.com/CRASH-Tech/talos-operator/cmd/kubernetes"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/client"
	talosCLient "github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	client *talosCLient.Client
}

func NewClient(ctx context.Context, endpoint string, machineConfig kubernetes.MachineConfig) *Client {
	tCtx := clientconfig.Context{
		CA:  machineConfig.MachineSecrets.CA,
		Crt: machineConfig.MachineSecrets.Crt,
		Key: machineConfig.MachineSecrets.Key,
	}

	cOpts := client.WithConfigContext(&tCtx)
	eOpts := client.WithEndpoints(endpoint)

	cert, err := base64.StdEncoding.DecodeString(machineConfig.MachineSecrets.Crt)
	if err != nil {
		panic(err)
	}
	key, err := base64.StdEncoding.DecodeString(machineConfig.MachineSecrets.Key)
	if err != nil {
		panic(err)
	}

	log.Info(string(cert))
	log.Info(string(key))

	xCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Info("yyy")
		panic(err)
	}

	log.Error(xCert)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{xCert},
		ClientAuth:         tls.NoClientCert,
	}

	tOpts := client.WithTLSConfig(tlsConfig)

	c, err := client.New(ctx, tOpts, cOpts, eOpts)
	if err != nil {
		panic(err)
	}

	r := machine.BootstrapRequest{
		RecoverEtcd:          false,
		RecoverSkipHashCheck: true,
	}

	err = c.Bootstrap(ctx, &r)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
	/////////

	// req := machine.ApplyConfigurationRequest{
	// 	//Mode: machine.ApplyConfigurationRequest_NO_REBOOT,
	// 	Data: []byte(machineConfig.MachineConfig),
	// }

	// x, err := c.ApplyConfiguration(ctx, &req)
	// if err != nil {
	// 	log.Info("yyyy")
	// 	panic(err)
	// }
	// log.Info(x)

	/////////////////
	client := Client{
		client: c,
	}

	return &client
}

func (client *Client) Bootstrap(endpoint string, machineConfig kubernetes.MachineConfig) {
	log.Infof("Bootstrap: %s", endpoint)

}

func (client *Client) ApplyConfiguration(endpoint string, machineConfig kubernetes.MachineConfig) {
	log.Infof("Apply config: %s", endpoint)

}

// func (client *Client) ApplyConfiguration(conf string) {
// 	log.Info("loool")
// 	tCtx := clientconfig.Context{
// 		CA:  machineConfig.MachineSecrets.CA,
// 		Crt: machineConfig.MachineSecrets.Crt,
// 		Key: machineConfig.MachineSecrets.Key,
// 		//Nodes:     []string{"10.171.120.151", "127.0.0.1"},
// 		//Endpoints: []string{"10.171.120.151", "127.0.0.1"},
// 		//Cluster:   "k-test",
// 	}

// 	cOpts := client.WithConfigContext(&tCtx)
// 	eOpts := client.WithEndpoints(endpoint)

// 	cert, err := base64.StdEncoding.DecodeString(machineConfig.MachineSecrets.Crt)
// 	if err != nil {
// 		panic(err)
// 	}
// 	key, err := base64.StdEncoding.DecodeString(machineConfig.MachineSecrets.Key)
// 	if err != nil {
// 		panic(err)
// 	}

// 	log.Info(string(cert))
// 	log.Info(string(key))

// 	xCert, err := tls.X509KeyPair(cert, key)
// 	if err != nil {
// 		log.Info("yyy")
// 		panic(err)
// 	}

// 	log.Error(xCert)

// 	tlsConfig := &tls.Config{
// 		InsecureSkipVerify: true,
// 		Certificates:       []tls.Certificate{xCert},
// 		ClientAuth:         tls.NoClientCert,
// 	}

// 	tOpts := client.WithTLSConfig(tlsConfig)

// 	c, err := client.New(ctx, tOpts, cOpts, eOpts)
// 	if err != nil {
// 		panic(err)
// 	}

// 	r := machine.BootstrapRequest{
// 		RecoverEtcd:          false,
// 		RecoverSkipHashCheck: true,
// 	}

// 	err = c.Bootstrap(ctx, &r)
// 	if err != nil {
// 		panic(err)
// 	}
// 	os.Exit(0)
// 	/////////

// 	// req := machine.ApplyConfigurationRequest{
// 	// 	//Mode: machine.ApplyConfigurationRequest_NO_REBOOT,
// 	// 	Data: []byte(machineConfig.MachineConfig),
// 	// }

// 	// x, err := c.ApplyConfiguration(ctx, &req)
// 	// if err != nil {
// 	// 	log.Info("yyyy")
// 	// 	panic(err)
// 	// }
// 	// log.Info(x)

// 	/////////////////
// 	client := Client{
// 		client: c,
// 	}

// 	return &client
// }
