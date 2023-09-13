package talos

import (
	"context"

	"github.com/siderolabs/talos/pkg/machinery/client"
	talosCLient "github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	client *talosCLient.Client
}

func NewClient(ctx context.Context) *Client {
	//opts := client.Options{}

	//conf := clientconfig.Config{}
	//conf.

	tCtx := clientconfig.Context{
		CA:        "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJQakNCOGFBREFnRUNBaEFZeEhKUDdlSlJRN2lqZUFSdWhLTlpNQVVHQXl0bGNEQVFNUTR3REFZRFZRUUsKRXdWMFlXeHZjekFlRncweU16QTVNVEl5TVRBeU1qSmFGdzB6TXpBNU1Ea3lNVEF5TWpKYU1CQXhEakFNQmdOVgpCQW9UQlhSaGJHOXpNQ293QlFZREsyVndBeUVBNE1CTW81d2JVZjcwYnRjRm84R3Y0b01QelFicXZsOUQ4Tk8vCkRSOHZ6SkdqWVRCZk1BNEdBMVVkRHdFQi93UUVBd0lDaERBZEJnTlZIU1VFRmpBVUJnZ3JCZ0VGQlFjREFRWUkKS3dZQkJRVUhBd0l3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFkQmdOVkhRNEVGZ1FVVkhBVTFRaTNQM2x1Q2JqZgp4SG5QaEpiKzBhWXdCUVlESzJWd0EwRUFsdk1GOUs3VUVSeDZlNTRJbUk2UVRaem83Mzc5Zzdnd0VsajBqTWpDClJoL2NmT2M4YzhUTkpSM2RDWU5oS1ZDUy9FYzV4U1JXRWVDR3J1NTBxbG81RGc9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==",
		Crt:       "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJLVENCM0tBREFnRUNBaEVBLzZ3cmtxODVPQnVTRzBhb0s0WXNFekFGQmdNclpYQXdFREVPTUF3R0ExVUUKQ2hNRmRHRnNiM013SGhjTk1qTXdPVEV5TWpFd09ESTVXaGNOTXpNd09UQTVNakV3T0RJNVdqQVRNUkV3RHdZRApWUVFLRXdodmN6cGhaRzFwYmpBcU1BVUdBeXRsY0FNaEFLbDlEYjlybmEycXcyWnV5RDl5ZHYzYTJ2TFFYWTg3CnRYUm1IblVtTmxJaW8wZ3dSakFPQmdOVkhROEJBZjhFQkFNQ0I0QXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0h3WURWUjBqQkJnd0ZvQVVWSEFVMVFpM1AzbHVDYmpmeEhuUGhKYiswYVl3QlFZREsyVndBMEVBUFUrMApLeWRGYmhNK3F6YUtXdStRTXNWSnd2T2YrbU1QQmpSOGovSXVJdmI1QmdTRzdlQjhlUGtRbjNYTW1rbEtwUVZECkZRVWlRRklUOXJjak42TTJBQT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K",
		Key:       "LS0tLS1CRUdJTiBFRDI1NTE5IFBSSVZBVEUgS0VZLS0tLS0KTUM0Q0FRQXdCUVlESzJWd0JDSUVJQmtKR3JmT3FubWtzcExlRUdvNUdXTnhtS3V2QUtQVFgwYUxEY0pLMTFlUQotLS0tLUVORCBFRDI1NTE5IFBSSVZBVEUgS0VZLS0tLS0K",
		Nodes:     []string{"10.171.120.151", "127.0.0.1"},
		Endpoints: []string{"10.171.120.151", "127.0.0.1"},
		Cluster:   "k-test",
	}

	opts := client.WithConfigContext(&tCtx)
	opts1 := client.WithEndpoints("10.171.120.151")

	c, err := client.New(ctx, opts, opts1)
	if err != nil {
		log.Info("sssssssssss")
		panic(err)
	}

	//c.
	client := Client{
		client: c,
	}

	return &client
}

func (client *Client) ApplyConfiguration(conf string) {
	log.Info("loool")

}
