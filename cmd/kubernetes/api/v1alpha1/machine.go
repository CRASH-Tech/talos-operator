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
	Spec       MachineSpec                `json:"spec"`
	Status     MachineStatus              `json:"status"`
}

type MachineSpec struct {
	Host      string          `json:"host"`
	Bootstrap bool            `json:"bootstrap"`
	Config    string          `json:"config"`
	Protected bool            `json:"protected"`
	Params    []MachineParams `json:"params"`
}

type MachineParams struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MachineStatus struct {
	Bootstrapped bool   `json:"bootstrapped"`
	ConfigHash   string `json:"confighash"`
	ApplySuccess bool   `json:"applysuccess"`
	Etcd         string `json:"etcd"`
	Apid         string `json:"apid"`
	Kubelet      string `json:"kubelet"`
	Containerd   string `json:"containerd"`
	Cri          string `json:"cri"`
	Machined     string `json:"machined"`
}
