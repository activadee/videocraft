#!/bin/bash

# Test script to validate GitHub workflow structure
# Tests for Issue #50: Remove Security Scan and Docker Tests

set -e

WORKFLOW_FILE=".github/workflows/test.yml"
EXIT_CODE=0

echo "🔴 RED PHASE: Testing current workflow structure (should FAIL after refactoring)"
echo "=============================================================================="

# Test 1: Security job should exist (will fail after removal)
echo "Test 1: Checking if security job exists..."
if grep -q "security:" "$WORKFLOW_FILE"; then
    echo "✅ PASS (RED): Security job found (lines 128-149)"
else
    echo "❌ FAIL (RED): Security job not found - this test should pass initially"
    EXIT_CODE=1
fi

# Test 2: Docker job should exist (will fail after removal)
echo "Test 2: Checking if docker job exists..."
if grep -q "docker:" "$WORKFLOW_FILE"; then
    echo "✅ PASS (RED): Docker job found (lines 212-234)"
else
    echo "❌ FAIL (RED): Docker job not found - this test should pass initially"
    EXIT_CODE=1
fi

# Test 3: Security tools installation should exist (will fail after removal)
echo "Test 3: Checking for security tools installation..."
if grep -q "govulncheck\|gosec" "$WORKFLOW_FILE"; then
    echo "✅ PASS (RED): Security tools found"
else
    echo "❌ FAIL (RED): Security tools not found"
    EXIT_CODE=1
fi

# Test 4: Docker Buildx setup should exist (will fail after removal)
echo "Test 4: Checking for Docker Buildx setup..."
if grep -q "docker/setup-buildx-action" "$WORKFLOW_FILE"; then
    echo "✅ PASS (RED): Docker Buildx setup found"
else
    echo "❌ FAIL (RED): Docker Buildx setup not found"
    EXIT_CODE=1
fi

# Test 5: Essential jobs should remain (should always pass)
echo "Test 5: Checking essential jobs remain..."
ESSENTIAL_JOBS=("lint:" "test:" "integration:" "coverage:" "benchmark:")
for job in "${ESSENTIAL_JOBS[@]}"; do
    if grep -q "$job" "$WORKFLOW_FILE"; then
        echo "✅ PASS: Essential job '$job' found"
    else
        echo "❌ FAIL: Essential job '$job' missing"
        EXIT_CODE=1
    fi
done

# Test 6: Job count validation
echo "Test 6: Counting total jobs..."
ACTUAL_JOBS=$(grep "^  [a-z-]*:$" "$WORKFLOW_FILE" | grep -v "push:" | grep -v "pull_request:" | wc -l)
echo "Current job count: $ACTUAL_JOBS"
if [ "$ACTUAL_JOBS" -eq 7 ]; then
    echo "✅ PASS (RED): Found expected 7 jobs (including security and docker)"
elif [ "$ACTUAL_JOBS" -eq 5 ]; then
    echo "❌ FAIL (RED): Found 5 jobs - security and docker have been removed"
    EXIT_CODE=1
else
    echo "❌ FAIL: Unexpected job count: $ACTUAL_JOBS"
    EXIT_CODE=1
fi

echo ""
echo "🧪 Test Summary:"
echo "- These tests validate the CURRENT state (before refactoring)"
echo "- After refactoring, Tests 1, 2, 3, 4, and 6 should FAIL (as expected)"
echo "- Test 5 should always PASS (essential jobs must remain)"
echo ""

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ ALL RED PHASE TESTS PASSED - Workflow structure validated"
    echo "Ready to proceed with GREEN PHASE (implementation)"
else
    echo "❌ SOME RED PHASE TESTS FAILED - Check workflow structure"
fi

exit $EXIT_CODE