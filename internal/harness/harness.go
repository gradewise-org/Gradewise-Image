package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gradewise-image/internal/runner"
)

// Config holds configuration for the harness
type Config struct {
	AssignmentTitle  string
	AuthToken        string
	BasicAuth        string
	SubmissionURL    string
	SubmitResultsURL string
	WorkDir          string
	ResultsPath      string
	Devel            bool
}

// Harness manages the autograder harness lifecycle
type Harness struct {
	config *Config
	client *http.Client
}

// New creates a new harness instance
func New(config *Config) *Harness {
	return &Harness{
		config: config,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Update downloads and updates the autograder code
func (h *Harness) Update() error {
	if h.config.SubmissionURL == "" {
		return fmt.Errorf("SUBMISSION_URL not set")
	}

	log.Printf("Downloading autograder from %s", h.config.SubmissionURL)

	// Create request
	req, err := http.NewRequest("GET", h.config.SubmissionURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication
	if h.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.config.AuthToken)
	} else if h.config.BasicAuth != "" {
		req.Header.Set("Authorization", "Basic "+h.config.BasicAuth)
	}

	// Make request
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Ensure work directory exists
	if err := os.MkdirAll(h.config.WorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Extract archive (assuming tar.gz format)
	archivePath := filepath.Join(h.config.WorkDir, "autograder.tar.gz")
	out, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return fmt.Errorf("failed to write archive: %w", err)
	}
	out.Close()

	// Extract archive
	extractDir := filepath.Join(h.config.WorkDir, "source")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	// Run tar to extract
	// First try with --strip-components=1 (for archives with a top-level directory)
	// If that fails or extracts nothing, try without it (for archives with files at root)
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", extractDir, "--strip-components=1")
	if err := cmd.Run(); err != nil {
		// If strip-components failed, try without it
		cmd = exec.Command("tar", "-xzf", archivePath, "-C", extractDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
		return nil
	}
	
	// Check if any files were extracted (in case --strip-components=1 silently succeeds but extracts nothing)
	entries, _ := os.ReadDir(extractDir)
	if len(entries) == 0 {
		// If nothing was extracted, try without strip-components
		cmd = exec.Command("tar", "-xzf", archivePath, "-C", extractDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
	}
	

	// Clean up archive
	os.Remove(archivePath)

	log.Println("Autograder updated successfully")
	return nil
}

// SubmitResults submits the grading results to the server
func (h *Harness) SubmitResults(results *runner.Results) error {
	if h.config.SubmitResultsURL == "" {
		return fmt.Errorf("SUBMIT_RESULTS_URL not set")
	}

	// Prepare submission data
	submissionData := map[string]interface{}{
		"score":          results.Score,
		"max_score":      results.MaxScore,
		"output":         results.Output,
		"visibility":     results.Visibility,
		"stdout_visibility": results.StdoutVisibility,
		"tests":          results.Tests,
		"execution_time": results.ExecutionTime,
	}

	jsonData, err := json.Marshal(submissionData)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", h.config.SubmitResultsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set authentication
	if h.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.config.AuthToken)
	} else if h.config.BasicAuth != "" {
		req.Header.Set("Authorization", "Basic "+h.config.BasicAuth)
	}

	// Make request
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submission failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Println("Results submitted successfully")
	return nil
}
