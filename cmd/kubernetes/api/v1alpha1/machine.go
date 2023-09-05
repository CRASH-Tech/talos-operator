package v1alpha1

import "github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api"

const (
	STATUS_POWER_ON      = "ON"
	STATUS_POWER_OFF     = "OFF"
	STATUS_POWER_UNKNOWN = "UNKNOWN"
)

type Machine struct {
	APIVersion string                     `json:"apiVersion"`
	Kind       string                     `json:"kind"`
	Metadata   api.CustomResourceMetadata `json:"metadata"`
	//api.CustomResource
	Spec   MachineSpec   `json:"spec"`
	Status MachineStatus `json:"status"`
}

type MachineSpec struct {
	Host   string      `json:"host"`
	Params interface{} `json:"params"`
}

type MachineStatus struct {
	Status string `json:"status"`
	Power  string `json:"power"`
}
