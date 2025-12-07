package parser

import (
	"testing"
)

func TestParseYAMLFile(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantErr   bool
	}{
		{
			name:     "Valid Deployment",
			filePath: "../../tests/data/valid_deployment.yaml",
			wantErr:  false,
		},
		{
			name:     "Invalid Syntax",
			filePath: "../../tests/data/invalid_syntax.yaml",
			wantErr:  true,
		},
		{
			name:     "File Not Found",
			filePath: "non_existent.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseYAMLFile(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYAMLFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetGVKFromObject(t *testing.T) {
	tests := []struct {
		name           string
		obj            map[string]interface{}
		wantApiVersion string
		wantKind       string
		wantErr        bool
	}{
		{
			name: "Valid Object",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
			},
			wantApiVersion: "v1",
			wantKind:       "Pod",
			wantErr:        false,
		},
		{
			name: "Missing apiVersion",
			obj: map[string]interface{}{
				"kind": "Pod",
			},
			wantErr: true,
		},
		{
			name: "Missing kind",
			obj: map[string]interface{}{
				"apiVersion": "v1",
			},
			wantErr: true,
		},
		{
			name:    "Nil Object",
			obj:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiVersion, kind, err := GetGVKFromObject(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGVKFromObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if apiVersion != tt.wantApiVersion {
					t.Errorf("GetGVKFromObject() apiVersion = %v, want %v", apiVersion, tt.wantApiVersion)
				}
				if kind != tt.wantKind {
					t.Errorf("GetGVKFromObject() kind = %v, want %v", kind, tt.wantKind)
				}
			}
		})
	}
}
