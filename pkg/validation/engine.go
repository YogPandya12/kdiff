package validation

import (
    "fmt"
    "strings"

    "github.com/YogPandya12/kdiff.git/pkg/parser"
    "github.com/go-openapi/spec"
    "github.com/go-openapi/strfmt"
    "github.com/go-openapi/validate"

    k8sspec "k8s.io/kube-openapi/pkg/validation/spec"
    validationErrors "github.com/go-openapi/errors"
)

type ValidationResult struct {
    Message string
    Path    string
    Severity string 
}

type Engine struct {
    schemaLoader SchemaLoader 
    loadedSchemas map[string]*k8sspec.Swagger 
    strict        bool 
}

func NewEngine(loader SchemaLoader, strict bool) *Engine {
    return &Engine{
        schemaLoader: loader,
        loadedSchemas: make(map[string]*k8sspec.Swagger),
        strict:        strict,
    }
}

// ValidateK8sObject validates a single Kubernetes object (map[string]interface{}) against its corresponding OpenAPI schema.
func (e *Engine) ValidateK8sObject(obj map[string]interface{}) ([]ValidationResult, error) {
    if obj == nil || len(obj) == 0 {
        return nil, fmt.Errorf("cannot validate empty Kubernetes object")
    }

    apiVersion, kind, err := parser.GetGVKFromObject(obj)
    if err != nil {
        return nil, fmt.Errorf("failed to get apiVersion and kind from object: %w", err)
    }

    // --- Step 1: Load/Get the Schema for the GVK ---
    group := ""
    version := apiVersion

    if strings.Contains(apiVersion, "/") {
        parts := strings.SplitN(apiVersion, "/", 2)
        group = parts[0]
        version = parts[1]
    } else if apiVersion == "v1" {
        group = "core"
    }

    // Construct the canonical schema definition name.
    schemaDefName := fmt.Sprintf("io.k8s.api.%s.%s.%s", group, version, kind)

    // Load the full OpenAPI schema for this API version/group if not already loaded.
    loadedSwagger, ok := e.loadedSchemas[apiVersion]
    if !ok {
        var loadErr error
        loadedSwagger, loadErr = e.schemaLoader.LoadSchema(apiVersion)
        if loadErr != nil {
            fmt.Printf("Failed to load schema for specific API version '%s'. Trying generic 'v2' schema: %v\n", apiVersion, loadErr)
            loadedSwagger, loadErr = e.schemaLoader.LoadSchema("v2")
            if loadErr != nil {
                 return nil, fmt.Errorf("failed to load schema for apiVersion '%s' or generic 'v2': %w", apiVersion, loadErr)
            }
        }
        e.loadedSchemas[apiVersion] = loadedSwagger
    }

    // --- Step 2: Find the specific schema definition ---
    schema, found := loadedSwagger.Definitions[schemaDefName]
    if !found {
        fmt.Printf("Schema definition '%s' not found directly. Attempting heuristic search...\n", schemaDefName)
        foundSchema := false
        for defName, def := range loadedSwagger.Definitions {
            if strings.Contains(defName, kind) && strings.Contains(defName, version) {
                if strings.Contains(defName, group) || group == "core" { 
                    schema = def
                    foundSchema = true
                    fmt.Printf("Heuristically found schema for %s/%s at definition: %s\n", apiVersion, kind, defName)
                    break
                }
            }
        }
        if !foundSchema {
            return nil, fmt.Errorf("schema definition not found for GVK %s/%s in loaded schema. Tried searching for '%s'. Ensure the schema for this API version is comprehensive.", apiVersion, kind, schemaDefName)
        }
    }

    // --- Step 3: Perform Validation ---
    goApiSchema := convertK8sSchemaToGoApi(schema)
    fullGoApiSwagger := convertK8sSwaggerToGoApi(loadedSwagger)

    validator := validate.NewSchemaValidator(goApiSchema, fullGoApiSwagger, "", strfmt.Default)

    validationResult := validator.Validate(obj) 

    var results []ValidationResult

    if validationResult != nil {
        for _, err := range validationResult.Errors {
            path := ""
            if valErr, ok := err.(*validationErrors.Validation); ok {
                path = valErr.Name
            }

            results = append(results, ValidationResult{
                Message:  err.Error(),
                Path:     path,
                Severity: "error",
            })
        }
    }
    return results, nil
}

func convertK8sSwaggerToGoApi(k8sSwagger *k8sspec.Swagger) *spec.Swagger {
	if k8sSwagger == nil {
		return nil
	}
	
	var goInfo *spec.Info
	if k8sSwagger.Info != nil {
		goInfo = &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       k8sSwagger.Info.Title,
				Description: k8sSwagger.Info.Description,
				Version:     k8sSwagger.Info.Version,
			},
		}
	}
	
	goSwagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger:     k8sSwagger.Swagger,
			Info:        goInfo,
			Paths:       nil, 
			Definitions: make(map[string]spec.Schema, len(k8sSwagger.Definitions)),
		},
	}
	
	for defName, k8sDef := range k8sSwagger.Definitions {
		goSwagger.Definitions[defName] = *convertK8sSchemaToGoApi(k8sDef)
	}
	
	return goSwagger
}

func convertK8sSchemaToGoApi(k8sSchema k8sspec.Schema) *spec.Schema {
    var goApiType spec.StringOrArray
    if len(k8sSchema.Type) > 0 {
        goApiType = make(spec.StringOrArray, len(k8sSchema.Type))
        copy(goApiType, k8sSchema.Type)
    }

    var properties map[string]spec.Schema
    if k8sSchema.Properties != nil {
        properties = make(map[string]spec.Schema)
        for key, k8sProp := range k8sSchema.Properties {
            properties[key] = *convertK8sSchemaToGoApi(k8sProp)
        }
    }

    var items *spec.SchemaOrArray
    if k8sSchema.Items != nil {
        if k8sSchema.Items.Schema != nil {
            items = &spec.SchemaOrArray{
                Schema: convertK8sSchemaToGoApi(*k8sSchema.Items.Schema),
            }
        }
    }

    // Handle Ref
    var ref spec.Ref
    if k8sSchema.Ref.String() != "" {
        ref = spec.MustCreateRef(k8sSchema.Ref.String())
    }

    return &spec.Schema{
        SchemaProps: spec.SchemaProps{
            Type:        goApiType,
            Format:      k8sSchema.Format,
            Title:       k8sSchema.Title,
            Description: k8sSchema.Description,
            Default:     k8sSchema.Default,
            Maximum:     k8sSchema.Maximum,
            Minimum:     k8sSchema.Minimum,
            MaxLength:   k8sSchema.MaxLength,
            MinLength:   k8sSchema.MinLength,
            Pattern:     k8sSchema.Pattern,
            MaxItems:    k8sSchema.MaxItems,
            MinItems:    k8sSchema.MinItems,
            UniqueItems: k8sSchema.UniqueItems,
            MultipleOf:  k8sSchema.MultipleOf,
            Enum:        k8sSchema.Enum,
            Required:    k8sSchema.Required,
            Properties:  properties,
            Items:       items,
            Ref:         ref,
        },
    }
}

func getErrorPath(err error) string {
    errStr := err.Error()
    if strings.Contains(errStr, "validation failed") {
        parts := strings.Split(errStr, ":")
        if len(parts) > 1 {
            return strings.TrimSpace(parts[0])
        }
    }
    return ""
}
