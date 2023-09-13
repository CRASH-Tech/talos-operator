package talos

import (
	"context"

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
		//Nodes:     []string{"10.171.120.151", "127.0.0.1"},
		//Endpoints: []string{"10.171.120.151", "127.0.0.1"},
		//Cluster:   "k-test",
	}

	cOpts := client.WithConfigContext(&tCtx)
	eOpts := client.WithEndpoints(endpoint)

	c, err := client.New(ctx, cOpts, eOpts)
	if err != nil {
		panic(err)
	}

	req := machine.ApplyConfigurationRequest{
		Mode: machine.ApplyConfigurationRequest_NO_REBOOT,
		Data: []byte(machineConfig.MachineConfig),
	}
	// op := grpc.WithInsecure()
	// lol := grpc.CallOption

	x, err := c.ApplyConfiguration(ctx, &req)
	if err != nil {
		log.Info("yyyy")
		panic(err)
	}
	log.Info(x)

	// err = c.Bootstrap(ctx, &req)
	// if err != nil {
	// 	log.Info("yyyy")
	// 	panic(err)
	// }

	//c.
	client := Client{
		client: c,
	}

	return &client
}

func (client *Client) ApplyConfiguration(conf string) {
	log.Info("loool")

}
