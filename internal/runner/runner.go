package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Test represents a single test case result
type Test struct {
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	MaxScore   float64 `json:"max_score"`
	Output     string  `json:"output"`
	Visibility string  `json:"visibility"` // "visible", "hidden", "after_due_date", "after_published"
}

// Results represents the complete grading results
type Results struct {
	Score            float64   `json:"score"`
	MaxScore         float64   `json:"max_score"`
	Output           string    `json:"output"`
	Visibility       string    `json:"visibility"`
	StdoutVisibility string    `json:"stdout_visibility"`
	Tests            []Test    `json:"tests"`
	ExecutionTime    float64   `json:"execution_time"`
}

// RunTests executes the autograder tests and returns results
func RunTests(workDir string) (*Results, error) {
	startTime := time.Now()

	// Look for common autograder entry points
	possibleScripts := []string{
		"run_autograder",
		"run_autograder.sh",
		"run_autograder.py",
		"autograder.sh",
		"autograder.py",
	}

	sourceDir := filepath.Join(workDir, "source")
	submissionDir := filepath.Join(workDir, "submission")

	// Ensure submission directory exists
	if err := os.MkdirAll(submissionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create submission directory: %w", err)
	}

	var scriptPath string
	for _, script := range possibleScripts {
		path := filepath.Join(sourceDir, script)
		if _, err := os.Stat(path); err == nil {
			scriptPath = path
			break
		}
	}

	if scriptPath == "" {
		// Try to find any executable script
		entries, err := os.ReadDir(sourceDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					path := filepath.Join(sourceDir, entry.Name())
					info, err := os.Stat(path)
					if err == nil && (info.Mode()&0111 != 0 || strings.HasSuffix(path, ".sh") || strings.HasSuffix(path, ".py")) {
						scriptPath = path
						break
					}
				}
			}
		}
	}

	if scriptPath == "" {
		return &Results{
			Score:      0,
			MaxScore:   0,
			Output:     "No autograder script found. Expected one of: run_autograder, run_autograder.sh, run_autograder.py, autograder.sh, autograder.py",
			Visibility: "visible",
		}, nil
	}

	// Determine how to run the script
	var cmd *exec.Cmd
	if strings.HasSuffix(scriptPath, ".py") {
		cmd = exec.Command("python3", scriptPath)
	} else if strings.HasSuffix(scriptPath, ".sh") || filepath.Ext(scriptPath) == "" {
		// Try as bash script or executable
		info, err := os.Stat(scriptPath)
		if err == nil && info.Mode()&0111 != 0 {
			// Executable, run directly
			cmd = exec.Command(scriptPath)
		} else {
			// Not executable, run with bash
			cmd = exec.Command("/bin/bash", scriptPath)
		}
	} else {
		// Default to bash
		cmd = exec.Command("/bin/bash", scriptPath)
	}

	cmd.Dir = sourceDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GRADER_DIR="+sourceDir)
	cmd.Env = append(cmd.Env, "SUBMISSION_DIR="+submissionDir)
	cmd.Env = append(cmd.Env, "RESULTS_DIR="+filepath.Join(workDir, "results"))

	output, err := cmd.CombinedOutput()
	executionTime := time.Since(startTime).Seconds()

	// Try to read results.json if it exists
	resultsPath := filepath.Join(workDir, "results", "results.json")
	results := &Results{
		Score:            0,
		MaxScore:         0,
		Output:           string(output),
		Visibility:       "visible",
		StdoutVisibility: "visible",
		Tests:            []Test{},
		ExecutionTime:    executionTime,
	}

	if data, err := os.ReadFile(resultsPath); err == nil {
		if err := json.Unmarshal(data, results); err == nil {
			// Successfully parsed results.json
			return results, nil
		}
	}

	// If no results.json, try to parse output or create a basic result
	if err != nil {
		results.Output = fmt.Sprintf("Error running autograder: %v\n\nOutput:\n%s", err, output)
		return results, nil
	}

	// If script ran successfully but no results.json, create a basic result
	if len(output) > 0 {
		results.Output = string(output)
	}

	return results, nil
}
