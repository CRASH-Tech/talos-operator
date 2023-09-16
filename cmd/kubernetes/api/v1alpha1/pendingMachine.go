package v1alpha1

import "github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api"

type PendingMachine struct {
	APIVersion string                     `json:"apiVersion"`
	Kind       string                     `json:"kind"`
	Metadata   api.CustomResourceMetadata `json:"metadata"`
	Spec       PendingMachineSpec         `json:"spec"`
}

type PendingMachineSpec struct {
	Host   string          `json:"host"`
	Params []MachineParams `json:"params"`
}
