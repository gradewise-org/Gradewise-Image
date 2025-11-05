# Test Suite

This directory contains the test suite for Gradewise-Image.

## Test Structure

- **Unit Tests**: Located in `internal/*/*_test.go`
- **Integration Tests**: Located in `test/integration_test.sh`

## Running Tests

### Unit Tests

Run all unit tests:
```bash
make test-unit
# or
go test ./...
```

Run with coverage:
```bash
make test-coverage
# or
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Run integration tests locally:
```bash
make test-integration
# or
./test/integration_test.sh
```

### All Tests

Run both unit and integration tests:
```bash
make test-all
```

## Integration Test Details

The integration test (`integration_test.sh`) performs the following:

1. **Test 1: Harness Update**
   - Builds the harness binary
   - Creates a test autograder archive
   - Starts an HTTP server to serve the archive
   - Tests downloading and extracting the autograder

2. **Test 2: Test Execution**
   - Copies a test submission
   - Runs the autograder tests
   - Verifies results.json is created with correct scores

3. **Test 3: Full Integration**
   - Tests the complete flow: update → run → submit
   - Verifies all components work together

4. **Test 4: Error Handling**
   - Tests behavior with missing submission files
   - Verifies graceful error handling

## Requirements

- Go 1.21 or later
- Python 3 (for integration test HTTP server)
- Bash shell
- Standard Unix utilities (tar, gzip, etc.)

## CI/CD

Tests run automatically in GitHub Actions on:
- Push to main/master/mvp branches
- Pull requests to main/master/mvp branches

See `.github/workflows/test.yml` for the CI configuration.

