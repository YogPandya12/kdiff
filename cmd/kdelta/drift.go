package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/YogPandya12/kdelta.git/pkg/diff"
	"github.com/YogPandya12/kdelta.git/pkg/history"
	"github.com/YogPandya12/kdelta.git/pkg/kube"
	"github.com/YogPandya12/kdelta.git/pkg/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var driftCmd = &cobra.Command{
	Use:   "drift <directory>",
	Short: "Check for configuration drift against a directory of desired state files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := args[0]

		// Initialize Kubernetes client
		client, err := kube.NewClient(kubeconfig, "") 
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating kubernetes client: %v\n", err)
			os.Exit(1)
		}

		rc, err := kube.NewResourceClient(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating resource client: %v\n", err)
			os.Exit(1)
		}

		err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
				return nil
			}

			fmt.Printf("Checking %s...\n", path)

			// 1. Parse local file
			localMap, err := parser.ParseYAMLFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error parsing file: %v\n", err)
				return nil // Continue to next file
			}
			localObj := &unstructured.Unstructured{Object: localMap}

			gvk := localObj.GroupVersionKind()
			name := localObj.GetName()
			namespace := localObj.GetNamespace()
			if namespace == "" {
				namespace = "default"
			}

			// 2. Map GVK to GVR (Simplified)
			var gvr schema.GroupVersionResource
			switch gvk.Kind {
			case "Deployment":
				gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
			case "Service":
				gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
			case "ConfigMap":
				gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
			case "Secret":
				gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
			default:
				resource := fmt.Sprintf("%ss", toLower(gvk.Kind))
				gvr = schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: resource}
			}

			// 3. Fetch live resource
			liveObj, err := rc.GetResource(gvr, namespace, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error fetching live resource %s/%s: %v\n", gvk.Kind, name, err)
				return nil
			}

			// 4. Normalize live resource
			if err := kube.Normalize(liveObj); err != nil {
				fmt.Fprintf(os.Stderr, "  Error normalizing live resource: %v\n", err)
				return nil
			}

			// 5. Compare
			localYAML, _ := parser.YAMLToCanonicalString(localObj.Object)
			liveYAML, _ := parser.YAMLToCanonicalString(liveObj.Object)

			diffs := diff.LineByLine(localYAML, liveYAML)

			hasDiff := false
			for _, d := range diffs {
				if d.Type != "common" {
					hasDiff = true
					break
				}
			}

			if hasDiff {
				fmt.Printf("  DRIFT DETECTED: %s/%s\n", gvk.Kind, name)
				
				// Construct diff string for log and output
				var diffBuilder strings.Builder
				additions := 0
				deletions := 0
				for _, d := range diffs {
					if d.Type == "added" {
						additions++
						diffBuilder.WriteString(fmt.Sprintf("+ %s\n", d.Line))
						fmt.Printf("    + %s\n", d.Line)
					} else if d.Type == "removed" {
						deletions++
						diffBuilder.WriteString(fmt.Sprintf("- %s\n", d.Line))
						fmt.Printf("    - %s\n", d.Line)
					}
				}

				// Log drift
				resourceID := fmt.Sprintf("%s/%s/%s", namespace, gvk.Kind, name)
				if err := history.LogDrift(resourceID, diffBuilder.String()); err != nil {
					fmt.Fprintf(os.Stderr, "  Error logging drift: %v\n", err)
				}
			} else {
				fmt.Printf("  Synced: %s/%s\n", gvk.Kind, name)
			}

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
			os.Exit(1)
		}
	},
}
