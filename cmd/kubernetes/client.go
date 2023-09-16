package kubernetes

import (
	"context"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	ctx        context.Context
	dynamic    dynamic.DynamicClient
	kubernetes kubernetes.Clientset
}

type V1alpha1 struct {
	client *Client
}

func NewClient(ctx context.Context, dynamic dynamic.DynamicClient, clientSet kubernetes.Clientset) *Client {
	client := Client{
		ctx:        ctx,
		dynamic:    dynamic,
		kubernetes: clientSet,
	}

	return &client
}

func (client *Client) dynamicCreate(resourceId schema.GroupVersionResource, obj interface{}) ([]byte, error) {
	uns, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
	if err != nil {
		return nil, err
	}

	uObj := unstructured.Unstructured{}

	uObj.Object = uns

	item, err := client.dynamic.Resource(resourceId).Create(client.ctx, &uObj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (client *Client) dynamicGet(resourceId schema.GroupVersionResource, name string) ([]byte, error) {

	item, err := client.dynamic.Resource(resourceId).Get(client.ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (client *Client) dynamicGetAll(resourceId schema.GroupVersionResource) ([][]byte, error) {

	items, err := client.dynamic.Resource(resourceId).List(client.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result [][]byte
	for _, item := range items.Items {
		jsonData, err := item.MarshalJSON()
		if err != nil {
			return nil, err
		}
		result = append(result, jsonData)
	}

	return result, nil
}

func (client *Client) dynamicPatch(resourceId schema.GroupVersionResource, name string, patch []byte) ([]byte, error) {

	item, err := client.dynamic.Resource(resourceId).Patch(client.ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (client *Client) dynamicUpdateStatus(resourceId schema.GroupVersionResource, name string, patch []byte) ([]byte, error) {
	var data unstructured.Unstructured
	err := data.UnmarshalJSON(patch)
	if err != nil {
		return nil, err
	}

	result, err := client.dynamic.Resource(resourceId).UpdateStatus(client.ctx, &data, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := result.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (client *Client) V1alpha1() *V1alpha1 {
	result := V1alpha1{
		client: client,
	}

	return &result
}

func (v1alpha1 *V1alpha1) Machine() *Machine {
	machine := Machine{
		client: v1alpha1.client,
		resourceId: schema.GroupVersionResource{
			Group:    "talos.xfix.org",
			Version:  "v1alpha1",
			Resource: "machine",
		},
	}

	return &machine
}

func (v1alpha1 *V1alpha1) PendingMachine() *PendingMachine {
	pMachine := PendingMachine{
		client: v1alpha1.client,
		resourceId: schema.GroupVersionResource{
			Group:    "talos.xfix.org",
			Version:  "v1alpha1",
			Resource: "pendingmachine",
		},
	}

	return &pMachine
}

func (v1alpha1 *V1alpha1) MachineSelector() *MachineSelector {
	selector := MachineSelector{
		client: v1alpha1.client,
		resourceId: schema.GroupVersionResource{
			Group:    "talos.xfix.org",
			Version:  "v1alpha1",
			Resource: "machineselector",
		},
	}

	return &selector
}

func (client *Client) GetMachineConfig(name, ns string) (MachineConfig, error) {
	secret, err := client.kubernetes.CoreV1().Secrets(ns).Get(client.ctx, name, metav1.GetOptions{})
	if err != nil {
		return MachineConfig{}, err
	}

	var machineSecrets MachineSecrets
	err = yaml.Unmarshal(secret.Data["machinesecrets"], &machineSecrets)
	if err != nil {
		return MachineConfig{}, err
	}

	result := MachineConfig{
		Name:           secret.Name,
		MachineConfig:  string(secret.Data["machineconfig"]),
		TalosConfig:    string(secret.Data["talosconfig"]),
		KubeConfig:     string(secret.Data["kubeconfig"]),
		MachineSecrets: machineSecrets,
	}

	return result, nil
}

func (client *Client) GetMachineConfigs(ns string) (map[string]MachineConfig, error) {
	result := make(map[string]MachineConfig)

	listOptions := metav1.ListOptions{
		LabelSelector: "talos/secret-type=machineconfig",
	}

	data, err := client.kubernetes.CoreV1().Secrets(ns).List(client.ctx, listOptions)
	if err != nil {
		return result, err
	}

	for _, secret := range data.Items {
		var machineSecrets MachineSecrets
		err = yaml.Unmarshal(secret.Data["machinesecrets"], &machineSecrets)
		if err != nil {
			return result, err
		}

		result[secret.Name] = MachineConfig{
			Name:           secret.Name,
			MachineConfig:  string(secret.Data["machineconfig"]),
			TalosConfig:    string(secret.Data["talosconfig"]),
			KubeConfig:     string(secret.Data["kubeconfig"]),
			MachineSecrets: machineSecrets,
		}
	}

	return result, nil
}
