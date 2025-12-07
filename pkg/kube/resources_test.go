package kube

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Test 1: No filters (default behavior)
	resources, err := rc.GetCoreResources("default", nil, "")
	if err != nil {
		t.Fatalf("GetCoreResources() error = %v", err)
	}

	expectedKinds := []string{"Deployment", "Service", "ConfigMap", "Secret"}
	for _, kind := range expectedKinds {
		if _, ok := resources[kind]; !ok {
			t.Errorf("GetCoreResources() missing kind %s", kind)
		}
	}

	// Test 2: Filter by Kind
	resources, err = rc.GetCoreResources("default", []string{"Deployment"}, "")
	if err != nil {
		t.Fatalf("GetCoreResources() error = %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("Expected 1 kind, got %d", len(resources))
	}
	if _, ok := resources["Deployment"]; !ok {
		t.Error("Expected Deployment kind")
	}
}

func TestListResources(t *testing.T) {
	client := fake.NewSimpleDynamicClient(scheme.Scheme)
	rc := &ResourceClient{DynamicClient: client}

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	list, err := rc.ListResources(gvr, "default", metav1.ListOptions{})
	if err != nil {
		t.Errorf("ListResources() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListResources() expected empty list, got %d items", len(list))
	}
}
