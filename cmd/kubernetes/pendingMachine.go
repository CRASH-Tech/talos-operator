package kubernetes

import (
	"encoding/json"
	"time"

	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type PendingMachine struct {
	client     *Client
	resourceId schema.GroupVersionResource
}

func (machine *PendingMachine) Create(m v1alpha1.PendingMachine) (v1alpha1.PendingMachine, error) {
	m.APIVersion = "talos.xfix.org/v1alpha1"
	m.Kind = "PendingMachine"
	m.Metadata.CreationTimestamp = time.Now().Format("2006-01-02T15:04:05Z")

	item, err := machine.client.dynamicCreate(machine.resourceId, &m)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	var result v1alpha1.PendingMachine
	err = json.Unmarshal(item, &result)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	return result, nil
}

func (machine *PendingMachine) Get(name string) (v1alpha1.PendingMachine, error) {
	item, err := machine.client.dynamicGet(machine.resourceId, name)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	var result v1alpha1.PendingMachine
	err = json.Unmarshal(item, &result)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	return result, nil
}

func (machine *PendingMachine) GetAll() ([]v1alpha1.PendingMachine, error) {
	items, err := machine.client.dynamicGetAll(machine.resourceId)
	if err != nil {
		panic(err)
	}

	var result []v1alpha1.PendingMachine
	for _, item := range items {
		var q v1alpha1.PendingMachine
		err = json.Unmarshal(item, &q)
		if err != nil {
			return nil, err
		}

		result = append(result, q)
	}

	return result, nil
}

func (machine *PendingMachine) Patch(m v1alpha1.Machine) (v1alpha1.PendingMachine, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	resp, err := machine.client.dynamicPatch(machine.resourceId, m.Metadata.Name, jsonData)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	var result v1alpha1.PendingMachine
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return v1alpha1.PendingMachine{}, err
	}

	return result, nil
}
