package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ResourceClient struct {
	DynamicClient dynamic.Interface
}

// NewResourceClient creates a new ResourceClient
func NewResourceClient(client *Client) (*ResourceClient, error) {
	restConfig, err := client.Config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get rest config: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &ResourceClient{
		DynamicClient: dynClient,
	}, nil
}

// ListResources fetches resources of a specific GVR (GroupVersionResource)
func (rc *ResourceClient) ListResources(gvr schema.GroupVersionResource, namespace string, opts metav1.ListOptions) ([]unstructured.Unstructured, error) {
	var list *unstructured.UnstructuredList
	var err error

	if namespace == "" {
		list, err = rc.DynamicClient.Resource(gvr).List(context.TODO(), opts)
	} else {
		list, err = rc.DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), opts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list resources for %v: %w", gvr, err)
	}

	return list.Items, nil
}

// GetCoreResources fetches common resources (Deployments, Services, ConfigMaps, Secrets)
func (rc *ResourceClient) GetCoreResources(namespace string, kinds []string, labelSelector string) (map[string][]unstructured.Unstructured, error) {
	resources := make(map[string][]unstructured.Unstructured)

	// Define GVRs for core resources
	allGvrs := map[string]schema.GroupVersionResource{
		"Deployment": {Group: "apps", Version: "v1", Resource: "deployments"},
		"Service":    {Group: "", Version: "v1", Resource: "services"},
		"ConfigMap":  {Group: "", Version: "v1", Resource: "configmaps"},
		"Secret":     {Group: "", Version: "v1", Resource: "secrets"},
	}

	// Filter GVRs based on requested kinds
	targetGvrs := make(map[string]schema.GroupVersionResource)
	if len(kinds) == 0 {
		targetGvrs = allGvrs
	} else {
		for _, k := range kinds {
			if gvr, ok := allGvrs[k]; ok {
				targetGvrs[k] = gvr
			}
		}
	}

	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	for kind, gvr := range targetGvrs {
		items, err := rc.ListResources(gvr, namespace, opts)
		if err != nil {
			return nil, err
		}
		resources[kind] = items
	}

	return resources, nil
}

// GetResource fetches a specific resource by name
func (rc *ResourceClient) GetResource(gvr schema.GroupVersionResource, namespace string, name string) (*unstructured.Unstructured, error) {
	return rc.DynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
