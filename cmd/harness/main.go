package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gradewise-image/internal/harness"
	"gradewise-image/internal/runner"
)

func main() {
	var (
		updateOnly  = flag.Bool("update-only", false, "Only update the harness, don't run tests")
		runOnly     = flag.Bool("run-only", false, "Only run tests, don't update")
		resultsPath = flag.String("results", "/autograder/results/results.json", "Path to results JSON file")
		workDir     = flag.String("workdir", "/autograder", "Working directory for autograder")
	)
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get environment variables
	assignmentTitle := os.Getenv("ASSIGNMENT_TITLE")
	authToken := os.Getenv("AUTHENTICATION_TOKEN")
	submissionURL := os.Getenv("SUBMISSION_URL")
	submitResultsURL := os.Getenv("SUBMIT_RESULTS_URL")
	basicAuth := os.Getenv("BASIC_AUTH")
	devel := os.Getenv("DEVEL") == "true"

	if assignmentTitle == "" {
		log.Println("Warning: ASSIGNMENT_TITLE not set")
	}
	if authToken == "" && basicAuth == "" {
		log.Println("Warning: No authentication credentials provided")
	}

	// Initialize harness
	h := harness.New(&harness.Config{
		AssignmentTitle:  assignmentTitle,
		AuthToken:        authToken,
		BasicAuth:        basicAuth,
		SubmissionURL:    submissionURL,
		SubmitResultsURL: submitResultsURL,
		WorkDir:          *workDir,
		ResultsPath:      *resultsPath,
		Devel:            devel,
	})

	// Update harness if needed
	if !*runOnly {
		log.Println("Updating harness...")
		if err := h.Update(); err != nil {
			log.Printf("Error updating harness: %v", err)
			if !*updateOnly {
				// Continue even if update fails
				log.Println("Continuing with existing harness...")
			} else {
				os.Exit(1)
			}
		}
		log.Println("Harness updated successfully")
	}

	if *updateOnly {
		log.Println("Update only mode, exiting")
		return
	}

	// Run tests
	log.Println("Running autograder tests...")
	results, err := runner.RunTests(*workDir)
	if err != nil {
		log.Printf("Error running tests: %v", err)
		results = &runner.Results{
			Score: 0,
			Output: fmt.Sprintf("Error running tests: %v", err),
			Visibility: "visible",
		}
	}

	// Save results
	if err := saveResults(*resultsPath, results); err != nil {
		log.Printf("Error saving results: %v", err)
		os.Exit(1)
	}

	log.Printf("Results saved to %s", *resultsPath)
	log.Printf("Score: %.2f/%.2f", results.Score, results.MaxScore)

	// Submit results if URL is provided
	if submitResultsURL != "" {
		log.Println("Submitting results...")
		if err := h.SubmitResults(results); err != nil {
			log.Printf("Error submitting results: %v", err)
			os.Exit(1)
		}
		log.Println("Results submitted successfully")
	} else {
		log.Println("SUBMIT_RESULTS_URL not set, skipping submission")
	}
}

func saveResults(path string, results *runner.Results) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// Write results JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	return nil
}
