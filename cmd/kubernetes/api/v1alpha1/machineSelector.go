package v1alpha1

import "github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api"

type MachineSelector struct {
	APIVersion string                     `json:"apiVersion"`
	Kind       string                     `json:"kind"`
	Metadata   api.CustomResourceMetadata `json:"metadata"`
	Spec       MachineSelectorSpec        `json:"spec"`
}

type MachineSelectorSpec struct {
	Config    string          `json:"config"`
	Bootstrap bool            `json:"bootstrap"`
	Params    []MachineParams `json:"params"`
}
