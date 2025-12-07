package history

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLogDrift(t *testing.T) {
	// Setup: Remove existing log file if any
	logFile := "drift.log"
	os.Remove(logFile)
	defer os.Remove(logFile) 

	// Test 1: Create new log file
	resource1 := "default/Deployment/nginx"
	diff1 := "+ replicas: 3"
	err := LogDrift(resource1, diff1)
	if err != nil {
		t.Fatalf("LogDrift failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("drift.log was not created")
	}

	// Read content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read drift.log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line in log, got %d", len(lines))
	}

	var entry1 DriftLogEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	if entry1.Resource != resource1 {
		t.Errorf("Expected resource %s, got %s", resource1, entry1.Resource)
	}
	if entry1.Diff != diff1 {
		t.Errorf("Expected diff %s, got %s", diff1, entry1.Diff)
	}

	// Test 2: Append to existing log file
	resource2 := "default/Service/nginx"
	diff2 := "+ type: LoadBalancer"
	err = LogDrift(resource2, diff2)
	if err != nil {
		t.Fatalf("LogDrift append failed: %v", err)
	}

	content, err = os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read drift.log: %v", err)
	}

	lines = strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in log, got %d", len(lines))
	}

	var entry2 DriftLogEntry
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("Failed to unmarshal second log entry: %v", err)
	}

	if entry2.Resource != resource2 {
		t.Errorf("Expected resource %s, got %s", resource2, entry2.Resource)
	}
}
