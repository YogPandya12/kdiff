package validation

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "k8s.io/kube-openapi/pkg/validation/spec"
)

type SchemaLoader interface {
    LoadSchema(apiVersion string) (*spec.Swagger, error)
}

type DefaultSchemaLoader struct {
    kubeAPIServerURL string
    cacheDir         string
    httpClient       *http.Client
}

func NewDefaultSchemaLoader(apiServerURL string) (*DefaultSchemaLoader, error) {
    cacheDir, err := os.UserCacheDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get user cache directory: %w", err)
    }
    kdiffCacheDir := filepath.Join(cacheDir, "kdiff", "schemas")
    if err := os.MkdirAll(kdiffCacheDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create kdiff schema cache directory: %w", err)
    }

    client := &http.Client{Timeout: 10 * time.Second}

    return &DefaultSchemaLoader{
        kubeAPIServerURL: apiServerURL,
        cacheDir:         kdiffCacheDir,
        httpClient:       client,
    }, nil
}

func (l *DefaultSchemaLoader) LoadSchema(apiVersion string) (*spec.Swagger, error) {
    cacheFilePath := filepath.Join(l.cacheDir, fmt.Sprintf("swagger-%s.json", apiVersion))

    schema, err := l.loadSchemaFromCache(cacheFilePath)
    if err == nil {
        fmt.Printf("Loaded schema for %s from cache.\n", apiVersion)
        return schema, nil
    }
    fmt.Printf("Cache miss or error for %s schema: %v. Fetching from API server...\n", apiVersion, err)

    schema, err = l.fetchSchemaFromAPIServer(apiVersion)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch schema for %s from API server: %w", apiVersion, err)
    }

    if err := l.saveSchemaToCache(schema, cacheFilePath); err != nil {
        fmt.Printf("Warning: Failed to save schema to cache for %s: %v\n", apiVersion, err)
    }

    return schema, nil
}

func (l *DefaultSchemaLoader) fetchSchemaFromAPIServer(apiVersion string) (*spec.Swagger, error) {
    url := fmt.Sprintf("%s/openapi/v2", l.kubeAPIServerURL)
    if apiVersion != "" {
        // This is a simplification. Real Kubernetes OpenAPI endpoints are more nuanced.
        // Let's stick to the generic /openapi/v2 for now for a simpler start.
    }

    fmt.Printf("Attempting to fetch schema from: %s\n", url)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Accept", "application/json;as=Swagger")

    resp, err := l.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make HTTP request to %s: %w", url, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to fetch schema from %s: HTTP status %d - %s", url, resp.StatusCode, string(bodyBytes))
    }

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    var swagger spec.Swagger
    if err := json.Unmarshal(bodyBytes, &swagger); err != nil {
        return nil, fmt.Errorf("failed to unmarshal OpenAPI schema: %w", err)
    }
    return &swagger, nil
}

func (l *DefaultSchemaLoader) loadSchemaFromCache(filePath string) (*spec.Swagger, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read schema from cache file %s: %w", filePath, err)
    }

    var swagger spec.Swagger
    if err := json.Unmarshal(data, &swagger); err != nil {
        _ = os.Remove(filePath)
        return nil, fmt.Errorf("failed to unmarshal schema from cache file %s (corrupted?): %w", filePath, err)
    }
    return &swagger, nil
}

func (l *DefaultSchemaLoader) saveSchemaToCache(schema *spec.Swagger, filePath string) error {
    data, err := json.MarshalIndent(schema, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal schema for caching: %w", err)
    }

    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write schema to cache file %w", err)
    }
    return nil
}

type DummySchemaLoader struct{}

func (d *DummySchemaLoader) LoadSchema(apiVersion string) (*spec.Swagger, error) {
    fmt.Printf("Using dummy schema loader for API version: %s\n", apiVersion)
    dummySchema := &spec.Swagger{
        SwaggerProps: spec.SwaggerProps{
            Swagger:     "2.0",
            Info:        &spec.Info{InfoProps: spec.InfoProps{Title: "Dummy Kubernetes API", Version: "v1.0"}},
            Paths:       &spec.Paths{},
            Definitions: map[string]spec.Schema{
                "io.k8s.api.apps.v1.Deployment": {
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
                                        "replicas": {SchemaProps: spec.SchemaProps{Type: []string{"integer"}, Format: "int32"}},
                                    },
                                },
                            },
                        },
                        Required: []string{"apiVersion", "kind", "metadata", "spec"},
                    },
                },
                "io.k8s.api.core.v1.Pod": { // Add Pod for testing
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
                                },
                            },
                        },
                        Required: []string{"apiVersion", "kind", "metadata", "spec"},
                    },
                },
            },
        },
    }
    return dummySchema, nil
}