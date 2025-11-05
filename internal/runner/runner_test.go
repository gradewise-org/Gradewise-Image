package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunTests_NoScriptFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source directory but no script
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	results, err := RunTests(tmpDir)
	if err != nil {
		t.Fatalf("RunTests should not return error when no script found: %v", err)
	}

	if results == nil {
		t.Fatal("Expected results to be non-nil")
	}

	if results.Score != 0 {
		t.Errorf("Expected score 0, got %f", results.Score)
	}

	if results.Output == "" {
		t.Error("Expected error message in output")
	}
}

func TestRunTests_WithResultsJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source directory with a script
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a simple script that generates results.json
	scriptPath := filepath.Join(sourceDir, "run_autograder.sh")
	scriptContent := `#!/bin/bash
mkdir -p "$RESULTS_DIR"
cat > "$RESULTS_DIR/results.json" <<EOF
{
  "score": 85.5,
  "max_score": 100.0,
  "output": "All tests passed",
  "visibility": "visible",
  "stdout_visibility": "visible",
  "tests": [
    {
      "name": "Test 1",
      "score": 85.5,
      "max_score": 100.0,
      "output": "Passed",
      "visibility": "visible"
    }
  ],
  "execution_time": 1.5
}
EOF
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	results, err := RunTests(tmpDir)
	if err != nil {
		t.Fatalf("RunTests failed: %v", err)
	}

	if results == nil {
		t.Fatal("Expected results to be non-nil")
	}

	if results.Score != 85.5 {
		t.Errorf("Expected score 85.5, got %f", results.Score)
	}

	if results.MaxScore != 100.0 {
		t.Errorf("Expected max_score 100.0, got %f", results.MaxScore)
	}

	if len(results.Tests) != 1 {
		t.Errorf("Expected 1 test, got %d", len(results.Tests))
	}

	if results.Tests[0].Name != "Test 1" {
		t.Errorf("Expected test name 'Test 1', got %s", results.Tests[0].Name)
	}
}

func TestRunTests_WithPythonScript(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a Python script
	scriptPath := filepath.Join(sourceDir, "run_autograder.py")
	scriptContent := `#!/usr/bin/env python3
import json
import os
import sys

results_dir = os.environ.get('RESULTS_DIR', '/tmp')
os.makedirs(results_dir, exist_ok=True)

results = {
    "score": 90.0,
    "max_score": 100.0,
    "output": "Python test passed",
    "visibility": "visible",
    "stdout_visibility": "visible",
    "tests": [],
    "execution_time": 0.5
}

with open(os.path.join(results_dir, 'results.json'), 'w') as f:
    json.dump(results, f)
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	results, err := RunTests(tmpDir)
	if err != nil {
		t.Fatalf("RunTests failed: %v", err)
	}

	if results.Score != 90.0 {
		t.Errorf("Expected score 90.0, got %f", results.Score)
	}
}

func TestRunTests_ScriptOutputOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create a script that doesn't generate results.json
	scriptPath := filepath.Join(sourceDir, "run_autograder.sh")
	scriptContent := `#!/bin/bash
echo "This is test output"
echo "Score: 75/100"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	results, err := RunTests(tmpDir)
	if err != nil {
		t.Fatalf("RunTests failed: %v", err)
	}

	if results.Output == "" {
		t.Error("Expected output to contain script output")
	}

	if !contains(results.Output, "This is test output") {
		t.Errorf("Expected output to contain 'This is test output', got: %s", results.Output)
	}
}

func TestRunTests_EnvVariables(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceDir := filepath.Join(tmpDir, "source")
	submissionDir := filepath.Join(tmpDir, "submission")
	resultsDir := filepath.Join(tmpDir, "results")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create script that writes env vars to results
	scriptPath := filepath.Join(sourceDir, "run_autograder.sh")
	scriptContent := `#!/bin/bash
mkdir -p "$RESULTS_DIR"
cat > "$RESULTS_DIR/results.json" <<EOF
{
  "score": 100.0,
  "max_score": 100.0,
  "output": "GRADER_DIR=$GRADER_DIR\nSUBMISSION_DIR=$SUBMISSION_DIR\nRESULTS_DIR=$RESULTS_DIR",
  "visibility": "visible",
  "stdout_visibility": "visible",
  "tests": [],
  "execution_time": 0.1
}
EOF
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	results, err := RunTests(tmpDir)
	if err != nil {
		t.Fatalf("RunTests failed: %v", err)
	}

	if !contains(results.Output, sourceDir) {
		t.Errorf("Expected GRADER_DIR to be %s, output: %s", sourceDir, results.Output)
	}
	if !contains(results.Output, submissionDir) {
		t.Errorf("Expected SUBMISSION_DIR to be %s, output: %s", submissionDir, results.Output)
	}
	if !contains(results.Output, resultsDir) {
		t.Errorf("Expected RESULTS_DIR to be %s, output: %s", resultsDir, results.Output)
	}
}

func TestRunTests_ExecutableScript(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create executable script (no extension)
	scriptPath := filepath.Join(sourceDir, "run_autograder")
	scriptContent := `#!/bin/bash
mkdir -p "$RESULTS_DIR"
echo '{"score": 80.0, "max_score": 100.0, "output": "Executable test", "visibility": "visible", "stdout_visibility": "visible", "tests": [], "execution_time": 0.2}' > "$RESULTS_DIR/results.json"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	results, err := RunTests(tmpDir)
	if err != nil {
		t.Fatalf("RunTests failed: %v", err)
	}

	if results.Score != 80.0 {
		t.Errorf("Expected score 80.0, got %f", results.Score)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
