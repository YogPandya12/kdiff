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
    "github.com/YogPandya12/kdiff.git/pkg/parser"
	"github.com/olekukonko/tablewriter"
)

var (
    compareOutputFormat string
    compareNoColor      bool
    compareOutputFile   string
)

var rootCmd = &cobra.Command{
    Use:   "kdiff",
    Short: "kdiff is a powerful CLI for comparing and validating Kubernetes configurations.",
    Long: `A robust command-line tool designed to simplify Kubernetes configuration management.
kdiff provides features like real-time validation, visualization of configuration differences,
and drift detection across various environments.`,
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



func init() {
    rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(compareCmd)

	// flags for compare command
	compareCmd.Flags().StringVarP(&compareOutputFormat, "output", "o", "default", "Output format (default, json, table)")
    compareCmd.Flags().BoolVar(&compareNoColor, "no-color", false, "Disable colorized output")
    compareCmd.Flags().StringVarP(&compareOutputFile, "output-file", "f", "", "Write output to a file instead of stdout")
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