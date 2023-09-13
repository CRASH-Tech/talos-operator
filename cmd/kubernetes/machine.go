package kubernetes

import (
	"encoding/json"
	"time"

	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MachineConfig struct {
	Name           string
	MachineConfig  string
	MachineSecrets MachineSecrets
	TalosConfig    string
	KubeConfig     string
}

type MachineSecrets struct {
	CA  string `yaml:"ca"`
	Crt string `yaml:"crt"`
	Key string `yaml:"key"`
}

type Machine struct {
	client     *Client
	resourceId schema.GroupVersionResource
}

func (machine *Machine) New(m v1alpha1.Machine) (v1alpha1.Machine, error) {
	m.APIVersion = "talos.xfix.org/v1alpha1"
	m.Kind = "Machine"
	m.Spec.Allocated = false
	m.Metadata.CreationTimestamp = time.Now().Format("2006-01-02T15:04:05Z")

	item, err := machine.client.dynamicCreate(machine.resourceId, &m)
	if err != nil {
		return v1alpha1.Machine{}, err
	}

	var result v1alpha1.Machine
	err = json.Unmarshal(item, &result)
	if err != nil {
		return v1alpha1.Machine{}, err
	}

	return result, nil
}

// func (qemu *Qemu) Get(name string) (v1alpha1.Qemu, error) {
// 	item, err := qemu.client.dynamicGet(qemu.resourceId, name)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	var result v1alpha1.Qemu
// 	err = json.Unmarshal(item, &result)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	return result, nil
// }

// func (qemu *Qemu) GetAll() ([]v1alpha1.Qemu, error) {
// 	items, err := qemu.client.dynamicGetAll(qemu.resourceId)
// 	if err != nil {
// 		panic(err)
// 	}

// 	var result []v1alpha1.Qemu
// 	for _, item := range items {
// 		var q v1alpha1.Qemu
// 		err = json.Unmarshal(item, &q)
// 		if err != nil {
// 			return nil, err
// 		}

// 		result = append(result, q)
// 	}

// 	return result, nil
// }

// func (qemu *Qemu) Patch(q v1alpha1.Qemu) (v1alpha1.Qemu, error) {
// 	jsonData, err := json.Marshal(q)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	resp, err := qemu.client.dynamicPatch(qemu.resourceId, q.Metadata.Name, jsonData)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	var result v1alpha1.Qemu
// 	err = json.Unmarshal(resp, &result)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	return result, nil
// }

// func (qemu *Qemu) UpdateStatus(q v1alpha1.Qemu) (v1alpha1.Qemu, error) {
// 	jsonData, err := json.Marshal(q)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	resp, err := qemu.client.dynamicUpdateStatus(qemu.resourceId, q.Metadata.Name, jsonData)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	var result v1alpha1.Qemu
// 	err = json.Unmarshal(resp, &result)
// 	if err != nil {
// 		return v1alpha1.Qemu{}, err
// 	}

// 	return result, nil
// }
