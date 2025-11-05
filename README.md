# Gradewise-Image

A Go-based autograder harness system inspired by Gradescope, designed for containerized automated grading environments.

## Overview

Gradewise-Image provides a Docker-based autograder harness that:
- Downloads and updates autograder code from a remote URL
- Runs test suites and collects results
- Submits grading results back to a server
- Supports SSH-based access for debugging
- Uses Go for improved performance and reliability

## Architecture

The system consists of:
- **Harness**: Go-based harness that manages the autograder lifecycle
- **Runner**: Executes autograder scripts and collects results
- **Docker Image**: Fedora-based container with all necessary tools

## Project Structure

```
.
├── cmd/
│   ├── harness/          # Main harness entry point
│   └── sshd-setup/       # SSH daemon setup utility
├── internal/
│   ├── harness/          # Harness core functionality
│   └── runner/           # Test execution and result collection
├── Dockerfile            # Container image definition
├── sshd_config           # SSH daemon configuration
├── go.mod                # Go module definition
└── Makefile              # Build automation

```

## Environment Variables

The harness reads the following environment variables:

- `ASSIGNMENT_TITLE`: Title of the assignment being graded
- `AUTHENTICATION_TOKEN`: Bearer token for API authentication
- `BASIC_AUTH`: Basic auth credentials (alternative to token)
- `SUBMISSION_URL`: URL to download autograder code (tar.gz format)
- `SUBMIT_RESULTS_URL`: URL to submit grading results
- `DEVEL`: Set to "true" for development mode

## Building

### Prerequisites

- Go 1.21 or later
- Docker (for building the container image)

### Build the binaries

```bash
make build
```

### Build the Docker image

```bash
make docker-build
```

## Usage

### Running the harness

The harness can be run directly or via Docker:

```bash
# Direct execution
bin/harness

# Docker
docker run gradewise-image:latest
```

### Command-line flags

- `-update-only`: Only update the harness, don't run tests
- `-run-only`: Only run tests, don't update
- `-results`: Path to results JSON file (default: `/autograder/results/results.json`)
- `-workdir`: Working directory for autograder (default: `/autograder`)

## Autograder Format

The harness expects autograder code in a tar.gz archive containing one of:
- `run_autograder`
- `run_autograder.sh`
- `run_autograder.py`
- `autograder.sh`
- `autograder.py`

The autograder script should:
1. Read student submission from `$SUBMISSION_DIR`
2. Run tests and generate results
3. Write results to `$RESULTS_DIR/results.json` in the following format:

```json
{
  "score": 85.5,
  "max_score": 100.0,
  "output": "Test output here",
  "visibility": "visible",
  "stdout_visibility": "visible",
  "tests": [
    {
      "name": "Test 1",
      "score": 10.0,
      "max_score": 10.0,
      "output": "Test passed",
      "visibility": "visible"
    }
  ],
  "execution_time": 2.5
}
```

## Development

### Running tests

```bash
# Run all tests (unit + integration)
make test-all

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run tests with coverage
make test-coverage
```

See [test/README.md](test/README.md) for more details on the test suite.

### Formatting code

```bash
make fmt
```

### Installing dependencies

```bash
make deps
```

## Differences from Gradescope

- **Language**: Written in Go instead of Python for better performance
- **Architecture**: Modular design with separate packages for harness and runner
- **Extensibility**: Easier to extend with additional features
- **Error handling**: More robust error handling and logging

## License

See LICENSE file for details.
