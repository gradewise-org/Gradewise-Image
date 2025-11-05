package harness

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gradewise-image/internal/runner"
)

func TestNew(t *testing.T) {
	config := &Config{
		AssignmentTitle: "Test Assignment",
		WorkDir:         "/tmp/test",
	}
	h := New(config)

	if h.config != config {
		t.Error("Expected config to be set")
	}
	if h.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestUpdate_MissingSubmissionURL(t *testing.T) {
	h := New(&Config{
		WorkDir: "/tmp/test",
	})

	err := h.Update()
	if err == nil {
		t.Error("Expected error when SUBMISSION_URL is missing")
	}
}

func TestUpdate_Success(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "harness_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test tar.gz archive
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	createTestArchive(t, archivePath, map[string]string{
		"run_autograder.sh": "#!/bin/bash\necho 'Hello World'",
		"test.txt":          "test content",
	})

	// Create HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Check authentication
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Bearer token, got %s", auth)
		}

		// Serve the archive
		data, err := os.ReadFile(archivePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(data)
	}))
	defer server.Close()

	// Create harness
	workDir := filepath.Join(tmpDir, "work")
	h := New(&Config{
		SubmissionURL: server.URL,
		AuthToken:     "test-token",
		WorkDir:       workDir,
	})

	// Run update
	err = h.Update()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify files were extracted
	sourceDir := filepath.Join(workDir, "source")
	scriptPath := filepath.Join(sourceDir, "run_autograder.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Debug: list what's actually in the directory
		entries, _ := os.ReadDir(sourceDir)
		t.Errorf("Expected run_autograder.sh to be extracted, but found: %v", entries)
	}

	testFilePath := filepath.Join(sourceDir, "test.txt")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Error("Expected test.txt to be extracted")
	}

	// Verify archive was cleaned up
	archivePath = filepath.Join(workDir, "autograder.tar.gz")
	if _, err := os.Stat(archivePath); err == nil {
		t.Error("Expected archive to be cleaned up")
	}
}

func TestUpdate_HTTPError(t *testing.T) {
	// Create HTTP test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "harness_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	h := New(&Config{
		SubmissionURL: server.URL,
		WorkDir:       tmpDir,
	})

	err = h.Update()
	if err == nil {
		t.Error("Expected error for HTTP 404")
	}
}

func TestUpdate_BasicAuth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "harness_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	createTestArchive(t, archivePath, map[string]string{
		"test.txt": "test",
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" {
			t.Errorf("Expected Basic auth, got %s", auth)
		}

		data, _ := os.ReadFile(archivePath)
		w.Write(data)
	}))
	defer server.Close()

	workDir := filepath.Join(tmpDir, "work")
	h := New(&Config{
		SubmissionURL: server.URL,
		BasicAuth:     "dXNlcm5hbWU6cGFzc3dvcmQ=",
		WorkDir:       workDir,
	})

	err = h.Update()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestSubmitResults_Success(t *testing.T) {
	results := &runner.Results{
		Score:            85.5,
		MaxScore:         100.0,
		Output:           "Test output",
		Visibility:       "visible",
		StdoutVisibility: "visible",
		Tests: []runner.Test{
			{
				Name:       "Test 1",
				Score:      85.5,
				MaxScore:   100.0,
				Output:     "Passed",
				Visibility: "visible",
			},
		},
		ExecutionTime: 2.5,
	}

	var receivedData map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type: application/json")
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Bearer token, got %s", auth)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &receivedData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	h := New(&Config{
		SubmitResultsURL: server.URL,
		AuthToken:        "test-token",
	})

	err := h.SubmitResults(results)
	if err != nil {
		t.Fatalf("SubmitResults failed: %v", err)
	}

	if receivedData["score"].(float64) != 85.5 {
		t.Errorf("Expected score 85.5, got %v", receivedData["score"])
	}
	if receivedData["max_score"].(float64) != 100.0 {
		t.Errorf("Expected max_score 100.0, got %v", receivedData["max_score"])
	}
}

func TestSubmitResults_MissingURL(t *testing.T) {
	h := New(&Config{})
	results := &runner.Results{}

	err := h.SubmitResults(results)
	if err == nil {
		t.Error("Expected error when SUBMIT_RESULTS_URL is missing")
	}
}

func TestSubmitResults_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	h := New(&Config{
		SubmitResultsURL: server.URL,
	})
	results := &runner.Results{}

	err := h.SubmitResults(results)
	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

// Helper function to create a test tar.gz archive
func createTestArchive(t *testing.T, path string, files map[string]string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for filename, content := range files {
		hdr := &tar.Header{
			Name: filename,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			t.Fatalf("Failed to write header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write content: %v", err)
		}
	}
}
