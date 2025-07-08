package parser

import (
    // "bytes"
    "fmt"
    "os"

    "gopkg.in/yaml.v3" 
)

func GetGVKFromObject(obj map[string]interface{}) (apiVersion string, kind string, err error) {
    if obj == nil {
        return "", "", fmt.Errorf("cannot get GVK from nil object")
    }

    apiV, ok := obj["apiVersion"].(string)
    if !ok || apiV == "" {
        return "", "", fmt.Errorf("object missing 'apiVersion' field or it's not a string")
    }
    apiVersion = apiV

    k, ok := obj["kind"].(string)
    if !ok || k == "" {
        return "", "", fmt.Errorf("object missing 'kind' field or it's not a string")
    }
    kind = k

    return apiVersion, kind, nil
}

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