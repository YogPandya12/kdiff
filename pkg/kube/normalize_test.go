package kube

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Remove status field",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
				"status": map[string]interface{}{
					"phase": "Running",
				},
			},
			expected: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			},
		},
		{
			name: "Remove metadata runtime fields",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":              "test-service",
					"uid":               "12345",
					"resourceVersion":   "999",
					"creationTimestamp": "2023-01-01T00:00:00Z",
					"generation":        int64(1),
					"managedFields":     []interface{}{"some", "fields"},
					"selfLink":          "/api/v1/namespaces/default/services/test-service",
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
			},
			expected: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "test-service",
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
			},
		},
		{
			name: "Remove specific annotations",
			input: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "test-deploy",
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/last-applied-configuration": "{}",
						"deployment.kubernetes.io/revision":                "1",
						"my-annotation":                                    "keep-me",
					},
				},
			},
			expected: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "test-deploy",
					"annotations": map[string]interface{}{
						"my-annotation": "keep-me",
					},
				},
			},
		},
		{
			name: "Nested structure preservation",
			input: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "test-cm",
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
			expected: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "test-cm",
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{Object: tt.input}
			err := Normalize(obj)
			if err != nil {
				t.Errorf("Normalize() error = %v", err)
				return
			}

			// Check status
			if _, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "status"); found {
				t.Errorf("status field should have been removed")
			}

			// Check metadata fields
			metaFields := []string{"uid", "resourceVersion", "creationTimestamp", "generation", "managedFields", "selfLink"}
			for _, field := range metaFields {
				if _, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "metadata", field); found {
					t.Errorf("metadata.%s should have been removed", field)
				}
			}

			// Check annotations
			if anns, found, _ := unstructured.NestedStringMap(obj.Object, "metadata", "annotations"); found {
				if _, ok := anns["kubectl.kubernetes.io/last-applied-configuration"]; ok {
					t.Errorf("last-applied-configuration annotation should have been removed")
				}
				if _, ok := anns["deployment.kubernetes.io/revision"]; ok {
					t.Errorf("deployment revision annotation should have been removed")
				}
			}
		})
	}
}
