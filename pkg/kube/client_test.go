package kube

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	// This test is tricky because it depends on the environment (kubeconfig).
	// We can test that it handles missing config gracefully or respects env vars.

	// Test case: Invalid path should return error
	_, err := NewClient("/non/existent/path", "")
	if err == nil {
		t.Error("Expected error for non-existent kubeconfig path, got nil")
	}

	// Test case: Env var override
	// Create a dummy file
	tmpFile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write minimal valid kubeconfig
	minimalConfig := `
apiVersion: v1
clusters:
- cluster:
    server: https://1.2.3.4
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: test-token
`
	if _, err := tmpFile.WriteString(minimalConfig); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	os.Setenv("KUBECONFIG", tmpFile.Name())
	defer os.Unsetenv("KUBECONFIG")

	client, err := NewClient("", "")
	if err != nil {
		t.Errorf("Expected success with valid env var, got error: %v", err)
	}
	if client == nil {
		t.Error("Expected client to be non-nil")
	}
}
