#!/bin/bash
# Example autograder script for Gradewise-Image

set -e

echo "Running autograder..."

# Create results directory
mkdir -p $RESULTS_DIR

# Example: Check if submission exists
if [ ! -d "$SUBMISSION_DIR" ]; then
    echo "Error: Submission directory not found"
    exit 1
fi

# Example: Run a simple test
score=0
max_score=100
output=""

# Check if a file exists
if [ -f "$SUBMISSION_DIR/submission.txt" ]; then
    score=$((score + 50))
    output="${output}✓ Found submission.txt (+50 points)\n"
else
    output="${output}✗ Missing submission.txt (-50 points)\n"
fi

# Check file content
if [ -f "$SUBMISSION_DIR/submission.txt" ]; then
    if grep -q "Hello World" "$SUBMISSION_DIR/submission.txt"; then
        score=$((score + 50))
        output="${output}✓ File contains 'Hello World' (+50 points)\n"
    else
        output="${output}✗ File does not contain 'Hello World' (-50 points)\n"
    fi
fi

# Create results JSON
cat > "$RESULTS_DIR/results.json" <<EOF
{
  "score": $score,
  "max_score": $max_score,
  "output": "$(echo -e "$output" | sed 's/"/\\"/g' | tr '\n' '\\n')",
  "visibility": "visible",
  "stdout_visibility": "visible",
  "tests": [
    {
      "name": "File exists",
      "score": $([ -f "$SUBMISSION_DIR/submission.txt" ] && echo "50" || echo "0"),
      "max_score": 50,
      "output": "$([ -f "$SUBMISSION_DIR/submission.txt" ] && echo "File found" || echo "File not found")",
      "visibility": "visible"
    },
    {
      "name": "File content",
      "score": $([ -f "$SUBMISSION_DIR/submission.txt" ] && grep -q "Hello World" "$SUBMISSION_DIR/submission.txt" && echo "50" || echo "0"),
      "max_score": 50,
      "output": "$([ -f "$SUBMISSION_DIR/submission.txt" ] && grep -q "Hello World" "$SUBMISSION_DIR/submission.txt" && echo "Content correct" || echo "Content incorrect")",
      "visibility": "visible"
    }
  ],
  "execution_time": 0.5
}
EOF

echo "Autograder completed. Score: $score/$max_score"

