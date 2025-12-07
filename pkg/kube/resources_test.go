package kube

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestGetCoreResources(t *testing.T) {
	// Create a fake dynamic client with standard scheme
	client := fake.NewSimpleDynamicClient(scheme.Scheme)

	// Create a ResourceClient with the fake client
	rc := &ResourceClient{
		DynamicClient: client,
	}

	// We can't easily pre-populate the fake client with resources because 
	// NewSimpleDynamicClient takes runtime.Object, but we are working with Unstructured.
	// However, the fake client should return empty lists, which is enough to test that
	// GetCoreResources runs without error and returns the expected map keys.

	resources, err := rc.GetCoreResources("default")
	if err != nil {
		t.Fatalf("GetCoreResources() error = %v", err)
	}

	expectedKinds := []string{"Deployment", "Service", "ConfigMap", "Secret"}
	for _, kind := range expectedKinds {
		if _, ok := resources[kind]; !ok {
			t.Errorf("GetCoreResources() missing kind %s", kind)
		}
	}
}

func TestListResources(t *testing.T) {
	client := fake.NewSimpleDynamicClient(scheme.Scheme)
	rc := &ResourceClient{DynamicClient: client}

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	list, err := rc.ListResources(gvr, "default")
	if err != nil {
		t.Errorf("ListResources() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListResources() expected empty list, got %d items", len(list))
	}
}
