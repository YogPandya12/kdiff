package diff

import (
    "fmt"
    "strings"
)

type LineDiffResult struct {
    Type  string 
    Line  string 
    Index int    
}

func LineByLine(text1, text2 string) []LineDiffResult {
    lines1 := strings.Split(text1, "\n")
    lines2 := strings.Split(text2, "\n")

    var results []LineDiffResult

    i, j := 0, 0
    for i < len(lines1) || j < len(lines2) {
        if i < len(lines1) && j < len(lines2) {
            if lines1[i] == lines2[j] {
                results = append(results, LineDiffResult{Type: "common", Line: lines1[i], Index: i + 1})
                i++
                j++
            } else {
                found1 := false
                for k := j; k < len(lines2); k++ {
                    if lines1[i] == lines2[k] {
                        for l := j; l < k; l++ {
                            results = append(results, LineDiffResult{Type: "added", Line: lines2[l], Index: l + 1})
                        }
                        j = k 
                        found1 = true
                        break
                    }
                }

                found2 := false
                for k := i; k < len(lines1); k++ {
                    if lines2[j] == lines1[k] {
                        for l := i; l < k; l++ {
                            results = append(results, LineDiffResult{Type: "removed", Line: lines1[l], Index: l + 1})
                        }
                        i = k 
                        found2 = true
                        break
                    }
                }

                if !found1 && !found2 {
                    results = append(results, LineDiffResult{Type: "removed", Line: lines1[i], Index: i + 1})
                    results = append(results, LineDiffResult{Type: "added", Line: lines2[j], Index: j + 1})
                    i++
                    j++
                } else if found1 && !found2 {
                    results = append(results, LineDiffResult{Type: "removed", Line: lines1[i], Index: i + 1})
                    i++
                } else if !found1 && found2 {
                    results = append(results, LineDiffResult{Type: "added", Line: lines2[j], Index: j + 1})
                    j++
                }
            }
        } else if i < len(lines1) {
            results = append(results, LineDiffResult{Type: "removed", Line: lines1[i], Index: i + 1})
            i++
        } else if j < len(lines2) {
            results = append(results, LineDiffResult{Type: "added", Line: lines2[j], Index: j + 1})
            j++
        }
    }

    return results
}

func FormatLineDiffResult(result LineDiffResult, useColor bool) string {
    prefix := ""
    line := result.Line

    if useColor {
        switch result.Type {
        case "added":
            prefix = "\033[32m+ \033[0m"
        case "removed":
            prefix = "\033[31m- \033[0m" 
        case "common":
            prefix = "  "
        }
    } else {
        switch result.Type {
        case "added":
            prefix = "+ "
        case "removed":
            prefix = "- "
        case "common":
            prefix = "  "
        }
    }
    return fmt.Sprintf("%s%s", prefix, line)
}