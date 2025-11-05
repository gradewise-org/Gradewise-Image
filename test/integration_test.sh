#!/bin/bash
# Integration test for Gradewise-Image harness
# This can be run locally or in CI

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR=$(mktemp -d)
TEST_PORT=18080
TEST_SERVER_PID=""

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    if [ -n "$TEST_SERVER_PID" ]; then
        kill $TEST_SERVER_PID 2>/dev/null || true
    fi
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}Starting integration tests...${NC}"

# Build the harness binary
echo -e "${YELLOW}Building harness binary...${NC}"
cd "$PROJECT_ROOT"
go build -o "$TEST_DIR/harness" ./cmd/harness

# Create test autograder
echo -e "${YELLOW}Creating test autograder...${NC}"
AUTOGRADER_DIR="$TEST_DIR/autograder"
mkdir -p "$AUTOGRADER_DIR"

cat > "$AUTOGRADER_DIR/run_autograder.sh" <<'EOF'
#!/bin/bash
set -e

mkdir -p "$RESULTS_DIR"

# Create a simple test that checks for submission.txt
score=0
max_score=100
output=""
tests=()

if [ -f "$SUBMISSION_DIR/submission.txt" ]; then
    content=$(cat "$SUBMISSION_DIR/submission.txt")
    if [ "$content" = "Hello, World!" ]; then
        score=100
        output="✓ Submission is correct!"
        tests+=('{"name": "Correct submission", "score": 100, "max_score": 100, "output": "Passed", "visibility": "visible"}')
    else
        score=50
        output="✗ Submission content is incorrect. Expected 'Hello, World!', got '$content'"
        tests+=('{"name": "Correct submission", "score": 50, "max_score": 100, "output": "Failed", "visibility": "visible"}')
    fi
else
    output="✗ submission.txt not found"
    tests+=('{"name": "File exists", "score": 0, "max_score": 100, "output": "File not found", "visibility": "visible"}')
fi

# Create results.json
cat > "$RESULTS_DIR/results.json" <<JSON
{
  "score": $score,
  "max_score": $max_score,
  "output": "$output",
  "visibility": "visible",
  "stdout_visibility": "visible",
  "tests": [$(IFS=','; echo "${tests[*]}")],
  "execution_time": 0.5
}
JSON

echo "Autograder completed. Score: $score/$max_score"
EOF

chmod +x "$AUTOGRADER_DIR/run_autograder.sh"

# Create test submission
echo -e "${YELLOW}Creating test submission...${NC}"
SUBMISSION_DIR="$TEST_DIR/submission"
mkdir -p "$SUBMISSION_DIR"
echo "Hello, World!" > "$SUBMISSION_DIR/submission.txt"

# Create tar.gz archive of autograder
echo -e "${YELLOW}Creating autograder archive...${NC}"
cd "$AUTOGRADER_DIR"
tar -czf "$TEST_DIR/autograder.tar.gz" .
cd "$PROJECT_ROOT"

# Start HTTP server to serve autograder and receive results
echo -e "${YELLOW}Starting test HTTP server on port $TEST_PORT...${NC}"
cat > "$TEST_DIR/server.py" <<'EOF'
#!/usr/bin/env python3
import http.server
import socketserver
import json
import os
import sys

PORT = int(sys.argv[1])
AUTOGRADER_PATH = sys.argv[2]
RESULTS_RECEIVED = []

class TestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/autograder.tar.gz':
            self.send_response(200)
            self.send_header('Content-Type', 'application/gzip')
            self.end_headers()
            with open(AUTOGRADER_PATH, 'rb') as f:
                self.wfile.write(f.read())
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == '/submit':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            results = json.loads(post_data.decode('utf-8'))
            RESULTS_RECEIVED.append(results)
            
            self.send_response(201)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        pass  # Suppress log messages

with socketserver.TCPServer(("", PORT), TestHandler) as httpd:
    httpd.serve_forever()
EOF

chmod +x "$TEST_DIR/server.py"
python3 "$TEST_DIR/server.py" "$TEST_PORT" "$TEST_DIR/autograder.tar.gz" &
TEST_SERVER_PID=$!
sleep 2  # Wait for server to start

# Test 1: Update harness
echo -e "${YELLOW}Test 1: Testing harness update...${NC}"
WORK_DIR="$TEST_DIR/work"
mkdir -p "$WORK_DIR"

export ASSIGNMENT_TITLE="Integration Test Assignment"
export AUTHENTICATION_TOKEN="test-token"
export SUBMISSION_URL="http://localhost:$TEST_PORT/autograder.tar.gz"
export SUBMIT_RESULTS_URL="http://localhost:$TEST_PORT/submit"
export DEVEL="true"

"$TEST_DIR/harness" -update-only -workdir "$WORK_DIR"

if [ ! -f "$WORK_DIR/source/run_autograder.sh" ]; then
    echo -e "${RED}✗ Test 1 FAILED: Autograder script not found after update${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Test 1 PASSED: Harness update successful${NC}"

# Test 2: Copy submission and run tests
echo -e "${YELLOW}Test 2: Testing test execution...${NC}"
cp -r "$SUBMISSION_DIR" "$WORK_DIR/submission"
mkdir -p "$WORK_DIR/results"

"$TEST_DIR/harness" -run-only -workdir "$WORK_DIR" -results "$WORK_DIR/results/results.json"

if [ ! -f "$WORK_DIR/results/results.json" ]; then
    echo -e "${RED}✗ Test 2 FAILED: Results file not created${NC}"
    exit 1
fi

# Check results
RESULTS=$(cat "$WORK_DIR/results/results.json")
if ! echo "$RESULTS" | grep -q '"score": 100'; then
    echo -e "${RED}✗ Test 2 FAILED: Expected score 100${NC}"
    echo "Results: $RESULTS"
    exit 1
fi
echo -e "${GREEN}✓ Test 2 PASSED: Test execution successful${NC}"

# Test 3: Full integration test (update + run + submit)
echo -e "${YELLOW}Test 3: Testing full integration (update + run + submit)...${NC}"
WORK_DIR2="$TEST_DIR/work2"
mkdir -p "$WORK_DIR2"
cp -r "$SUBMISSION_DIR" "$WORK_DIR2/submission"

# Clear any previous results
rm -f "$TEST_DIR/server_results.json"

# Run full harness (update + run + submit)
"$TEST_DIR/harness" -workdir "$WORK_DIR2" -results "$WORK_DIR2/results/results.json"

# Verify results were created
if [ ! -f "$WORK_DIR2/results/results.json" ]; then
    echo -e "${RED}✗ Test 3 FAILED: Results file not created${NC}"
    exit 1
fi

# Verify results content
RESULTS=$(cat "$WORK_DIR2/results/results.json")
if ! echo "$RESULTS" | grep -q '"score": 100'; then
    echo -e "${RED}✗ Test 3 FAILED: Expected score 100${NC}"
    echo "Results: $RESULTS"
    exit 1
fi

echo -e "${GREEN}✓ Test 3 PASSED: Full integration test completed${NC}"

# Test 4: Test with missing submission
echo -e "${YELLOW}Test 4: Testing with missing submission...${NC}"
WORK_DIR3="$TEST_DIR/work3"
mkdir -p "$WORK_DIR3"
mkdir -p "$WORK_DIR3/submission"  # Empty submission directory

"$TEST_DIR/harness" -run-only -workdir "$WORK_DIR3" -results "$WORK_DIR3/results/results.json"

if [ ! -f "$WORK_DIR3/results/results.json" ]; then
    echo -e "${RED}✗ Test 4 FAILED: Results file not created${NC}"
    exit 1
fi

RESULTS=$(cat "$WORK_DIR3/results/results.json")
if ! echo "$RESULTS" | grep -q '"score": 0'; then
    echo -e "${RED}✗ Test 4 FAILED: Expected score 0 for missing submission${NC}"
    echo "Results: $RESULTS"
    exit 1
fi
echo -e "${GREEN}✓ Test 4 PASSED: Missing submission handled correctly${NC}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All integration tests PASSED!${NC}"
echo -e "${GREEN}========================================${NC}"
