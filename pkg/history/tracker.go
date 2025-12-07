package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type DriftLogEntry struct {
	Timestamp string `json:"timestamp"`
	Resource  string `json:"resource"`
	Diff      string `json:"diff"`
}

// LogDrift appends a drift event to the drift.log file
func LogDrift(resource string, diff string) error {
	entry := DriftLogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Resource:  resource,
		Diff:      diff,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal drift log entry: %w", err)
	}

	f, err := os.OpenFile("drift.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open drift.log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write to drift.log: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline to drift.log: %w", err)
	}

	return nil
}
