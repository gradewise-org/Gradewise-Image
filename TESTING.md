# Testing Guide

This document describes the testing strategy for Gradewise-Image.

## Test Types

### Unit Tests

Unit tests are located alongside the code they test in `internal/*/*_test.go`. They test individual components in isolation.

**Coverage:**
- `internal/harness/harness_test.go`: Tests harness update and submission functionality
- `internal/runner/runner_test.go`: Tests test execution and result collection

**Running:**
```bash
go test ./...
go test -v ./...  # Verbose output
go test -cover ./...  # With coverage
```

### Integration Tests

Integration tests verify that all components work together correctly. They are located in `test/integration_test.sh`.

**What they test:**
1. Harness update (downloading and extracting autograder)
2. Test execution (running autograder scripts)
3. Full integration (update → run → submit)
4. Error handling (missing files, etc.)

**Running:**
```bash
./test/integration_test.sh
# or
make test-integration
```

**Requirements:**
- Go installed and in PATH
- Python 3 installed
- Network access (for HTTP server in tests)

## CI/CD Testing

Tests run automatically in GitHub Actions on every push and pull request.

**Workflow:** `.github/workflows/test.yml`

**Jobs:**
1. **unit-tests**: Runs all unit tests with coverage
2. **integration-tests**: Runs full integration test suite
3. **build**: Builds binaries and verifies they work
4. **docker-build**: Builds and tests Docker image

## Writing Tests

### Unit Test Guidelines

1. Use table-driven tests for multiple test cases
2. Mock external dependencies (HTTP servers, file system)
3. Use `t.Helper()` for test helper functions
4. Clean up temporary files and resources

Example:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name string
        input string
        expected string
    }{
        {"test1", "input1", "expected1"},
        {"test2", "input2", "expected2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Integration Test Guidelines

1. Use temporary directories for test files
2. Clean up resources in `trap` handlers
3. Use unique ports for test servers
4. Verify all expected outcomes

## Test Coverage

Current coverage targets:
- Aim for >80% code coverage
- Focus on critical paths (harness update, test execution, result submission)

View coverage:
```bash
make test-coverage
open coverage.html
```

## Troubleshooting

### Tests fail locally but pass in CI

- Check Go version (should be 1.21+)
- Ensure all dependencies are downloaded: `go mod download`
- Check for port conflicts (integration tests use port 18080)

### Integration tests hang

- Check if Python server is running: `ps aux | grep python`
- Kill any existing test servers: `pkill -f server.py`
- Check firewall settings

### Docker build tests fail

- Ensure Docker is running
- Check disk space
- Try clearing Docker cache: `docker system prune`

