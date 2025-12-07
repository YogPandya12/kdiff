package tests

import (
	"errors"
	"testing"

	"github.com/YogPandya12/kdiff.git/pkg/kube"
)

// Mock error to simulate auth failure
var errUnauthorized = errors.New("Unauthorized")

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"Unauthorized", errors.New("Unauthorized"), true},
		{"Forbidden", errors.New("Forbidden"), true},
		{"Auth Required", errors.New("authentication required"), true},
		{"Invalid Config", errors.New("invalid configuration"), true},
		{"Other Error", errors.New("connection refused"), false},
		{"Nil Error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kube.IsAuthError(tt.err); got != tt.expected {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.expected)
			}
		})
	}
}
