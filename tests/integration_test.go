package tests

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIIntegration(t *testing.T) {
	// Path to main.go
	mainPath := "../cmd/kdiff/main.go"
	dataPath := "data"

	tests := []struct {
		name           string
		args           []string
		expectedOutput string // Substring to check for
		expectError    bool
	}{
		// --- Validate Command Tests ---
		{
			name:           "Validate Valid Deployment",
			args:           []string{"validate", filepath.Join(dataPath, "valid_deployment.yaml")},
			expectedOutput: "is valid against its Kubernetes schema",
			expectError:    false,
		},
		{
			name:           "Validate Invalid Syntax",
			args:           []string{"validate", filepath.Join(dataPath, "invalid_syntax.yaml")},
			expectedOutput: "Error parsing",
			expectError:    true,
		},
		{
			name:           "Validate Invalid Type Replicas",
			args:           []string{"validate", filepath.Join(dataPath, "invalid_type_replicas.yaml")},
			expectedOutput: "validation issues",
			expectError:    true,
		},
		{
			name:           "Validate Missing Kind",
			args:           []string{"validate", filepath.Join(dataPath, "invalid_missing_kind.yaml")},
			expectedOutput: "missing 'kind' field",
			expectError:    true,
		},

		// --- Compare Command Tests ---
		{
			name:           "Compare Identical Files",
			args:           []string{"compare", filepath.Join(dataPath, "diff_base.yaml"), filepath.Join(dataPath, "diff_base.yaml")},
			expectedOutput: "No significant differences found",
			expectError:    false,
		},
		{
			name:           "Compare Different Files",
			args:           []string{"compare", filepath.Join(dataPath, "diff_base.yaml"), filepath.Join(dataPath, "diff_modified.yaml")},
			expectedOutput: "value1-modified", // Check for specific diff content
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdArgs := append([]string{"run", mainPath}, tt.args...)
			cmd := exec.Command("go", cmdArgs...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				if err == nil && !strings.Contains(outputStr, "Error") && !strings.Contains(outputStr, "issues") {
					t.Errorf("Expected error but got none. Output: %s", outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, outputStr)
				}
			}

			if tt.expectedOutput != "" && !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, but got: %s", tt.expectedOutput, outputStr)
			}
		})
	}
}
