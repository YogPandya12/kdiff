package kube

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Normalize removes runtime-specific fields from a Kubernetes resource
// to allow for accurate comparison between local and live states.
func Normalize(obj *unstructured.Unstructured) error {
	unstructured.RemoveNestedField(obj.Object, "status")

	// Remove metadata fields that are runtime-specific
	metadata, found, err := unstructured.NestedMap(obj.Object, "metadata")
	if err != nil || !found {
		return nil
	}

	// List of metadata fields to remove
	fieldsToRemove := []string{
		"managedFields",
		"creationTimestamp",
		"resourceVersion",
		"uid",
		"generation",
		"selfLink",
	}

	for _, field := range fieldsToRemove {
		delete(metadata, field)
	}

	// Remove specific annotations if necessary
	if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
		delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
		delete(annotations, "deployment.kubernetes.io/revision")
		
		if len(annotations) == 0 {
			delete(metadata, "annotations")
		}
	}

	// Set updated metadata back to object
	obj.SetUnstructuredContent(obj.Object)
	if err := unstructured.SetNestedMap(obj.Object, metadata, "metadata"); err != nil {
		return err
	}

	return nil
}
