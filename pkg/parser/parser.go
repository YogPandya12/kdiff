package parser

import (
    // "bytes"
    "fmt"
    "os"

    "gopkg.in/yaml.v3" 
)

func ParseYAMLFile(filePath string) (map[string]interface{}, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("file not found: %s", filePath)
        }
        return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
    }

    var result map[string]interface{}
    if err := yaml.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to parse YAML from %s: %w", filePath, err)
    }

    return result, nil
}

func YAMLToCanonicalString(data map[string]interface{}) (string, error) {
    canonicalBytes, err := yaml.Marshal(data)
    if err != nil {
        return "", fmt.Errorf("failed to marshal YAML to canonical string: %w", err)
    }
    return string(canonicalBytes), nil
}