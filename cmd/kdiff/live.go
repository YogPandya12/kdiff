package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/YogPandya12/kdiff.git/pkg/diff"
	"github.com/YogPandya12/kdiff.git/pkg/kube"
	"github.com/YogPandya12/kdiff.git/pkg/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var liveCmd = &cobra.Command{
	Use:   "live <file_path>",
	Short: "Compare a local YAML file against a live cluster resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// 1. Parse local file
		localMap, err := parser.ParseYAMLFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing file '%s': %v\n", filePath, err)
			os.Exit(1)
		}
		localObj := &unstructured.Unstructured{Object: localMap}

		gvk := localObj.GroupVersionKind()
		name := localObj.GetName()
		namespace := localObj.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}

		fmt.Printf("Comparing local file '%s' against live resource %s/%s in namespace '%s'...\n", filePath, gvk.Kind, name, namespace)

		// 2. Initialize Kubernetes client
		client, err := kube.NewClient(kubeconfig, "") // Use current context
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating kubernetes client: %v\n", err)
			os.Exit(1)
		}

		rc, err := kube.NewResourceClient(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating resource client: %v\n", err)
			os.Exit(1)
		}

		// 3. Map GVK to GVR (Simplified mapping for core resources)
		// TODO: Implement dynamic discovery later
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
			// Try to guess plural
			resource := fmt.Sprintf("%ss", toLower(gvk.Kind))
			gvr = schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: resource}
		}

		// 4. Fetch live resource
		liveObj, err := rc.GetResource(gvr, namespace, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching live resource: %v\n", err)
			os.Exit(1)
		}

		// 5. Normalize live resource
		if err := kube.Normalize(liveObj); err != nil {
			fmt.Fprintf(os.Stderr, "Error normalizing live resource: %v\n", err)
			os.Exit(1)
		}

		// 6. Convert to YAML strings
		localYAML, _ := parser.YAMLToCanonicalString(localObj.Object)
		liveYAML, _ := parser.YAMLToCanonicalString(liveObj.Object)

		// 7. Compare
		diffs := diff.LineByLine(localYAML, liveYAML)
		
		hasDiff := false
		for _, d := range diffs {
			if d.Type != "common" {
				hasDiff = true
				break
			}
		}

		if hasDiff {
			additions := 0
			deletions := 0
			for _, d := range diffs {
				if d.Type == "added" {
					additions++
				} else if d.Type == "removed" {
					deletions++
				}
			}
			fmt.Printf("Mismatch: %s/%s (+%d, -%d)\n", gvk.Kind, name, additions, deletions)
			for _, d := range diffs {
				if d.Type == "added" {
					fmt.Printf("  + %s\n", d.Line)
				} else if d.Type == "removed" {
					fmt.Printf("  - %s\n", d.Line)
				}
			}
		} else {
			fmt.Printf("Match: %s/%s\n", gvk.Kind, name)
		}
	},
}

func toLower(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	if r[0] >= 'A' && r[0] <= 'Z' {
		r[0] += 32
	}
	return string(r)
}
