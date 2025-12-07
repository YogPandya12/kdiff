package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	client := &http.Client{Timeout: 30 * time.Second}

	return &DefaultSchemaLoader{
		kubeAPIServerURL: apiServerURL,
		cacheDir:         kdiffCacheDir,
		httpClient:       client,
	}, nil
}

func (l *DefaultSchemaLoader) LoadSchema(apiVersion string) (*spec.Swagger, error) {
	// Normalize apiVersion (e.g., "v1" -> "v1", "apps/v1" -> "v1.28.0-standalone")
	// For simplicity in MVP, we might just use a fixed version or try to detect.
	// But here we will try to cache by the exact string provided first.
	safeVersion := strings.ReplaceAll(apiVersion, "/", "_")
	cacheFilePath := filepath.Join(l.cacheDir, fmt.Sprintf("swagger-%s.json", safeVersion))

	schema, err := l.loadSchemaFromCache(cacheFilePath)
	if err == nil {
		// fmt.Printf("Loaded schema for %s from cache.\n", apiVersion)
		return schema, nil
	}
	// fmt.Printf("Cache miss for %s schema. Fetching...\n", apiVersion)

	// Strategy 1: Fetch from API Server if URL is provided
	if l.kubeAPIServerURL != "" {
		schema, err = l.fetchSchemaFromAPIServer(apiVersion)
		if err == nil {
			if saveErr := l.saveSchemaToCache(schema, cacheFilePath); saveErr != nil {
				fmt.Printf("Warning: Failed to save schema to cache: %v\n", saveErr)
			}
			return schema, nil
		}
		fmt.Printf("Failed to fetch from API server: %v. Trying fallback...\n", err)
	}

	// Strategy 2: Fallback to GitHub (Kubernetes upstream)
	// We default to a recent stable version if we can't determine the exact version needed.
	// Ideally, we should allow the user to specify the K8s version via flag.
	// For now, let's default to v1.29.0 for the fallback.
	k8sVersion := "v1.29.0" 
	schema, err = l.fetchSchemaFromGitHub(k8sVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema from GitHub fallback: %w", err)
	}

	if saveErr := l.saveSchemaToCache(schema, cacheFilePath); saveErr != nil {
		fmt.Printf("Warning: Failed to save schema to cache: %v\n", saveErr)
	}

	return schema, nil
}

func (l *DefaultSchemaLoader) fetchSchemaFromAPIServer(apiVersion string) (*spec.Swagger, error) {
	url := fmt.Sprintf("%s/openapi/v2", l.kubeAPIServerURL)
	// Note: Real K8s API server might require auth headers. 
	// This simple implementation assumes a proxy or open access for now, 
	// or that the user provided a URL that includes auth (unlikely).
	// For MVP, this is a placeholder for "connected" mode.
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	return l.parseSwagger(resp.Body)
}

func (l *DefaultSchemaLoader) fetchSchemaFromGitHub(k8sVersion string) (*spec.Swagger, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/%s/api/openapi-spec/swagger.json", k8sVersion)
	// fmt.Printf("Fetching schema from GitHub: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d from GitHub", resp.StatusCode)
	}

	return l.parseSwagger(resp.Body)
}

func (l *DefaultSchemaLoader) parseSwagger(r io.Reader) (*spec.Swagger, error) {
	bodyBytes, err := io.ReadAll(r)
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
		return nil, err
	}

	var swagger spec.Swagger
	if err := json.Unmarshal(data, &swagger); err != nil {
		_ = os.Remove(filePath) // Corrupted cache
		return nil, err
	}
	return &swagger, nil
}

func (l *DefaultSchemaLoader) saveSchemaToCache(schema *spec.Swagger, filePath string) error {
	data, err := json.Marshal(schema) // Save compact to save space, or Indent for readability
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}