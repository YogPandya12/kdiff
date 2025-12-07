package validation

import (
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

// MockSchemaLoader for testing
type MockSchemaLoader struct {
	schemas map[string]*spec.Swagger
}

func (m *MockSchemaLoader) LoadSchema(apiVersion string) (*spec.Swagger, error) {
	if schema, ok := m.schemas[apiVersion]; ok {
		return schema, nil
	}
	return nil, nil // Return nil if not found (simulating error or missing schema)
}

func TestValidateK8sObject(t *testing.T) {
	// Define a simple schema for testing
	podSchema := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Definitions: map[string]spec.Schema{
				"io.k8s.api.core.v1.Pod": {
					SchemaProps: spec.SchemaProps{
						Type: []string{"object"},
						Properties: map[string]spec.Schema{
							"apiVersion": {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
							"kind":       {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
							"metadata":   {SchemaProps: spec.SchemaProps{Type: []string{"object"}}},
							"spec": {
								SchemaProps: spec.SchemaProps{
									Type: []string{"object"},
									Properties: map[string]spec.Schema{
										"containers": {SchemaProps: spec.SchemaProps{Type: []string{"array"}}},
									},
									Required: []string{"containers"},
								},
							},
						},
						Required: []string{"apiVersion", "kind", "metadata", "spec"},
					},
				},
			},
		},
	}

	mockLoader := &MockSchemaLoader{
		schemas: map[string]*spec.Swagger{
			"v1": podSchema,
		},
	}

	engine := NewEngine(mockLoader, true)

	tests := []struct {
		name    string
		obj     map[string]interface{}
		wantErr bool
	}{
		{
			name: "Valid Pod",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{"name": "nginx"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Pod (Missing Spec)",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			},
			wantErr: true,
		},
		{
			name: "Unknown Kind",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Unknown",
			},
			wantErr: true, // Should fail because schema for Unknown is not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.ValidateK8sObject(tt.obj)
			if err != nil {
				// If we expect an error (like schema loading failure), this is fine
				if !tt.wantErr {
					t.Errorf("ValidateK8sObject() unexpected error = %v", err)
				}
				return
			}

			// If we got results, check if we have validation errors
			hasValidationErrors := len(results) > 0
			if hasValidationErrors != tt.wantErr {
				t.Errorf("ValidateK8sObject() hasErrors = %v, wantErr %v. Results: %v", hasValidationErrors, tt.wantErr, results)
			}
		})
	}
}
