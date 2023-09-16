package kubernetes

import (
	"encoding/json"

	"github.com/CRASH-Tech/talos-operator/cmd/kubernetes/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MachineSelector struct {
	client     *Client
	resourceId schema.GroupVersionResource
}

func (machine *MachineSelector) Get(name string) (v1alpha1.MachineSelector, error) {
	item, err := machine.client.dynamicGet(machine.resourceId, name)
	if err != nil {
		return v1alpha1.MachineSelector{}, err
	}

	var result v1alpha1.MachineSelector
	err = json.Unmarshal(item, &result)
	if err != nil {
		return v1alpha1.MachineSelector{}, err
	}

	return result, nil
}

func (machine *MachineSelector) GetAll() ([]v1alpha1.MachineSelector, error) {
	items, err := machine.client.dynamicGetAll(machine.resourceId)
	if err != nil {
		panic(err)
	}

	var result []v1alpha1.MachineSelector
	for _, item := range items {
		var q v1alpha1.MachineSelector
		err = json.Unmarshal(item, &q)
		if err != nil {
			return nil, err
		}

		result = append(result, q)
	}

	return result, nil
}
