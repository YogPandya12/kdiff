package main

import (
    // "bytes"
    "io"
    "fmt"
    "os"
	"encoding/json"
	"strings"
	
    "github.com/spf13/cobra" 
	"github.com/YogPandya12/kdiff.git/pkg/diff"
	"github.com/YogPandya12/kdiff.git/pkg/kube"
	"github.com/YogPandya12/kdiff.git/pkg/parser"
    "github.com/YogPandya12/kdiff.git/pkg/validation"
	"github.com/olekukonko/tablewriter"
)

var (
    compareOutputFormat string
    compareNoColor      bool
    compareOutputFile   string
    validateStrict bool
)

var (
	kubeconfig string
	context    string
	kubeClient *kube.Client
)

var rootCmd = &cobra.Command{
	Use:   "kdiff",
	Short: "Kubernetes Diff Tool",
	Long: `kdiff is a tool for comparing Kubernetes configurations.
It allows you to compare local YAML files, validate them against schemas,
and check for differences between local files and live cluster resources.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize Kubernetes client
		// We don't error out here if connection fails, as some commands (like local diff)
		// might not need it. Individual commands should check if kubeClient is nil or valid.
		client, err := kube.NewClient(kubeconfig, context)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize Kubernetes client: %v\n", err)
		} else {
			kubeClient = client
		}
	},
    Run: func(cmd *cobra.Command, args []string) {
        cmd.Help()
    },
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version number of kdiff",
    Long:  `All software has versions. This is kdiff's.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("kdiff v0.1.0-alpha") 
    },
}
var compareCmd = &cobra.Command{
    Use:   "compare <file1> <file2>",
    Short: "Compares two Kubernetes configuration files",
    Long: `The compare command takes two file paths as arguments and performs a detailed

comparison of their Kubernetes resource definitions. This helps in identifying
differences between configurations easily.`,
    Args: cobra.ExactArgs(2), 
    Run: func(cmd *cobra.Command, args []string) {
        file1Path := args[0]
        file2Path := args[1]
        
        if compareOutputFile != "" {
            compareNoColor = true
        }

        var targetWriter io.Writer = os.Stdout 
        if compareOutputFile != "" {
            file, err := os.Create(compareOutputFile)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", compareOutputFile, err)
                os.Exit(1)
            }
            defer file.Close()
            targetWriter = file
        }

        // --- Step 1: Parse YAML files ---
        yaml1, err := parser.ParseYAMLFile(file1Path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file1Path, err)
            os.Exit(1)
        }

        yaml2, err := parser.ParseYAMLFile(file2Path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file2Path, err)
            os.Exit(1)
        }

        // --- Step 2: Canonicalize YAML content ---
        text1, err := parser.YAMLToCanonicalString(yaml1)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file1Path, err)
            os.Exit(1)
        }
        text1 = strings.TrimSuffix(text1, "\n")

        text2, err := parser.YAMLToCanonicalString(yaml2)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file2Path, err)
            os.Exit(1)
        }
        text2 = strings.TrimSuffix(text2, "\n")

        // --- Step 3: Perform Line-by-Line Diff ---
        diffResults := diff.LineByLine(text1, text2)

        // --- Step 4: Check if Differences Exist ---
        hasDifferences := false
        for _, res := range diffResults {
            if res.Type != "common" {
                hasDifferences = true
                break
            }
        }

        // --- Step 5: Output Formatting ---
        switch compareOutputFormat {
        case "json":
            type DiffEntry struct {
                Type       string `json:"type"`
                LineNumber int    `json:"lineNumber"`
                Content    string `json:"content"`
            }

            var jsonOutput []DiffEntry
            for _, res := range diffResults {
                content := strings.ReplaceAll(res.Line, "\n", "\\n")
                jsonOutput = append(jsonOutput, DiffEntry{
                    Type:       res.Type,
                    LineNumber: res.Index,
                    Content:    content,
                })
            }

            jsonBytes, err := json.MarshalIndent(jsonOutput, "", "  ")
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error marshaling JSON output: %v\n", err)
                os.Exit(1)
            }

            fmt.Fprintln(targetWriter, string(jsonBytes))

        case "table":
            writer := tablewriter.NewWriter(targetWriter)
            writer.SetHeader([]string{"Type", "Line No.", "Content"})
            writer.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
            writer.SetCenterSeparator("|")
            writer.SetColumnSeparator("|")
            writer.SetRowSeparator("-")
            writer.SetHeaderLine(true)

            writer.SetAutoWrapText(false)
            writer.SetAutoFormatHeaders(false)

            for _, res := range diffResults {
                lineContent := strings.TrimSpace(strings.ReplaceAll(res.Line, "\n", " "))
                lineContent = strings.ReplaceAll(lineContent, "\r", "")
        
                if !compareNoColor {
                    var rowColor []tablewriter.Colors
                    switch res.Type {
                    case "added":
                        rowColor = []tablewriter.Colors{
                            {tablewriter.FgGreenColor},
                            {tablewriter.FgGreenColor},
                            {tablewriter.FgGreenColor},
                        }
                    case "removed":
                        rowColor = []tablewriter.Colors{
                            {tablewriter.FgRedColor},
                            {tablewriter.FgRedColor},
                            {tablewriter.FgRedColor},
                        }
                    default:
                        rowColor = []tablewriter.Colors{}
                    }
        
                    writer.Rich([]string{
                        strings.ToUpper(res.Type),
                        fmt.Sprintf("%d", res.Index),
                        lineContent,
                    }, rowColor)
                } else {
                    writer.Append([]string{
                        strings.ToUpper(res.Type),
                        fmt.Sprintf("%d", res.Index),
                        lineContent,
                    })
                }
            }
            writer.Render()
        case "default":
            fallthrough
        default:
            for _, res := range diffResults {
                fmt.Fprintln(targetWriter, diff.FormatLineDiffResult(res, !compareNoColor))
            }
        }

        // --- Step 6: Summary ---
        if !hasDifferences && compareOutputFormat == "default" && compareOutputFile == "" {
            fmt.Fprintln(targetWriter, "\nNo significant differences found between the files.")
        }
    },
}

var validateCmd = &cobra.Command{
    Use:   "validate <file1> [file2...]",
    Short: "Validates Kubernetes configuration files against OpenAPI schemas",
    Long: `The validate command checks Kubernetes manifests for structural correctness
and adherence to OpenAPI schema definitions for the specified Kubernetes API version.`,
    Args: cobra.MinimumNArgs(1), 
    Run: func(cmd *cobra.Command, args []string) {
        // --- Schema Loader Selection ---
        // Initialize the real schema loader (fallback to GitHub if no API server)
        // For MVP, we pass empty string for apiServerURL to force fallback or use default
        schemaLoader, err := validation.NewDefaultSchemaLoader("")
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error initializing schema loader: %v\n", err)
            os.Exit(1)
        }

        validatorEngine := validation.NewEngine(schemaLoader, validateStrict) 

        allErrors := false

        for _, filePath := range args {
            fmt.Printf("\nValidating file: %s\n", filePath)

            obj, err := parser.ParseYAMLFile(filePath)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filePath, err)
                allErrors = true
                continue
            }

            validationResults, err := validatorEngine.ValidateK8sObject(obj)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error during validation of %s: %v\n", filePath, err)
                allErrors = true
                continue
            }

            if len(validationResults) == 0 {
                fmt.Printf("✅ %s is valid against its Kubernetes schema.\n", filePath)
            } else {
                allErrors = true
                fmt.Printf("❌ %s has validation issues:\n", filePath)
                for _, res := range validationResults {
                    prefix := ""
                    switch res.Severity {
                    case "error":
                        prefix = "\033[31mError:\033[0m"
                    case "warning":
                        prefix = "\033[33mWarning:\033[0m" 
                    }
                    fmt.Printf("  %s %s (Path: %s)\n", prefix, res.Message, res.Path)
                }
            }
        }

        if allErrors {
            os.Exit(1) 
        }
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(compareCmd)
    rootCmd.AddCommand(validateCmd)
    rootCmd.AddCommand(liveCmd)
    rootCmd.AddCommand(driftCmd)

	// flags for compare command
	compareCmd.Flags().StringVarP(&compareOutputFormat, "output", "o", "default", "Output format (default, json, table)")
    compareCmd.Flags().BoolVar(&compareNoColor, "no-color", false, "Disable colorized output")
    compareCmd.Flags().StringVarP(&compareOutputFile, "output-file", "f", "", "Write output to a file instead of stdout")

    // flags for validate command 
    validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Enable strict validation (e.g., disallow unknown fields)")
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err) 
        os.Exit(1)                  
    }
}

func main() {
    Execute() 
}