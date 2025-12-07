package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/YogPandya12/kdiff.git/pkg/diff"
	"github.com/YogPandya12/kdiff.git/pkg/kube"
	"github.com/YogPandya12/kdiff.git/pkg/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	targetNamespace string
	allNamespaces   bool
	targetResources []string
	labelSelector   string
)

var clusterCmd = &cobra.Command{
	Use:   "cluster <context1> <context2>",
	Short: "Compare resources between two clusters/contexts",
	Long:  `Compare core resources (Deployments, Services, ConfigMaps, Secrets) between two Kubernetes contexts.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		context1 := args[0]
		context2 := args[1]

		fmt.Printf("Comparing context '%s' vs '%s'...\n", context1, context2)

		// Initialize clients for both contexts
		client1, err := kube.NewClient(kubeconfig, context1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating client for context '%s': %v\n", context1, err)
			os.Exit(1)
		}

		client2, err := kube.NewClient(kubeconfig, context2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating client for context '%s': %v\n", context2, err)
			os.Exit(1)
		}

		// Initialize resource clients
		rc1, err := kube.NewResourceClient(client1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating resource client for '%s': %v\n", context1, err)
			os.Exit(1)
		}

		rc2, err := kube.NewResourceClient(client2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating resource client for '%s': %v\n", context2, err)
			os.Exit(1)
		}

		// Determine namespace
		namespace := targetNamespace
		if allNamespaces {
			namespace = "" // Empty string means all namespaces
			fmt.Println("Fetching resources from ALL namespaces...")
		} else {
			if namespace == "" {
				namespace = "default"
			}
			fmt.Printf("Fetching resources from namespace '%s'...\n", namespace)
		}

		if len(targetResources) > 0 {
			fmt.Printf("Filtering for resources: %v\n", targetResources)
		}
		if labelSelector != "" {
			fmt.Printf("Filtering by label: %s\n", labelSelector)
		}

		resources1, err := rc1.GetCoreResources(namespace, targetResources, labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching resources from '%s': %v\n", context1, err)
			os.Exit(1)
		}

		resources2, err := rc2.GetCoreResources(namespace, targetResources, labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching resources from '%s': %v\n", context2, err)
			os.Exit(1)
		}

		compareResources(context1, context2, resources1, resources2)
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.Flags().StringVarP(&targetNamespace, "namespace", "n", "default", "Target namespace")
	clusterCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces")
	clusterCmd.Flags().StringSliceVarP(&targetResources, "resource", "r", []string{}, "List of resources to compare (e.g., Deployment,Service)")
	clusterCmd.Flags().StringVarP(&labelSelector, "label", "l", "", "Selector (label query) to filter on")
}

func compareResources(ctx1, ctx2 string, res1, res2 map[string][]unstructured.Unstructured) {
	kinds := []string{"Deployment", "Service", "ConfigMap", "Secret"}

	for _, kind := range kinds {
		list1 := res1[kind]
		list2 := res2[kind]

		map1 := toMap(list1)
		map2 := toMap(list2)

		names := make(map[string]bool)
		for k := range map1 {
			names[k] = true
		}
		for k := range map2 {
			names[k] = true
		}

		sortedNames := make([]string, 0, len(names))
		for k := range names {
			sortedNames = append(sortedNames, k)
		}
		sort.Strings(sortedNames)

		for _, name := range sortedNames {
			obj1, exists1 := map1[name]
			obj2, exists2 := map2[name]

			if exists1 && exists2 {
				// Compare
				yaml1, _ := parser.YAMLToCanonicalString(obj1.Object)
				yaml2, _ := parser.YAMLToCanonicalString(obj2.Object)

				diffs := diff.LineByLine(yaml1, yaml2)
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
					fmt.Printf("Mismatch: %s/%s (+%d, -%d)\n", kind, name, additions, deletions)

					// Print detailed diff
					for _, d := range diffs {
						if d.Type == "added" {
							fmt.Printf("  [%s] + %s\n", ctx2, d.Line)
						} else if d.Type == "removed" {
							fmt.Printf("  [%s] + %s\n", ctx1, d.Line)
						}
					}
					fmt.Println() // Add a newline for separation
				} else {
					// fmt.Printf("Match: %s/%s\n", kind, name)
				}

			} else if exists1 {
				fmt.Printf("Only in %s: %s/%s\n", ctx1, kind, name)
			} else {
				fmt.Printf("Only in %s: %s/%s\n", ctx2, kind, name)
			}
		}
	}
}

func toMap(list []unstructured.Unstructured) map[string]unstructured.Unstructured {
	m := make(map[string]unstructured.Unstructured)
	for _, item := range list {
		m[item.GetName()] = item
	}
	return m
}
