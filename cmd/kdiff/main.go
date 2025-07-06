package main

import (
    "fmt"
    "os"
    "github.com/spf13/cobra" 
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

IGNORE_WHEN_COPYING_START
Use code with caution.
IGNORE_WHEN_COPYING_END

comparison of their Kubernetes resource definitions. This helps in identifying
differences between configurations easily.`,
    Args: cobra.ExactArgs(2),
    Run: func(cmd *cobra.Command, args []string) {
        file1 := args[0]
        file2 := args[1]
        fmt.Printf("Comparing '%s' with '%s' (logic to be implemented)\n", file1, file2)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(compareCmd)
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
