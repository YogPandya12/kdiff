package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ResourceClient handles fetching resources from Kubernetes
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
func (rc *ResourceClient) ListResources(gvr schema.GroupVersionResource, namespace string) ([]unstructured.Unstructured, error) {
	var list *unstructured.UnstructuredList
	var err error

	if namespace == "" {
		list, err = rc.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	} else {
		list, err = rc.DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list resources for %v: %w", gvr, err)
	}

	return list.Items, nil
}

// GetCoreResources fetches common resources (Deployments, Services, ConfigMaps, Secrets)
func (rc *ResourceClient) GetCoreResources(namespace string) (map[string][]unstructured.Unstructured, error) {
	resources := make(map[string][]unstructured.Unstructured)

	// Define GVRs for core resources
	gvrs := map[string]schema.GroupVersionResource{
		"Deployment": {Group: "apps", Version: "v1", Resource: "deployments"},
		"Service":    {Group: "", Version: "v1", Resource: "services"},
		"ConfigMap":  {Group: "", Version: "v1", Resource: "configmaps"},
		"Secret":     {Group: "", Version: "v1", Resource: "secrets"},
	}

	for kind, gvr := range gvrs {
		items, err := rc.ListResources(gvr, namespace)
		if err != nil {
			// Log error but continue? Or fail? For now, let's fail to be explicit.
			return nil, err
		}
		resources[kind] = items
	}

	return resources, nil
}
