#!/bin/bash

# Integration test for workflow refactoring
# Tests specific acceptance criteria from Issue #50

set -e

WORKFLOW_FILE=".github/workflows/test.yml"

echo "🧪 WORKFLOW REFACTORING INTEGRATION TESTS"
echo "========================================"

# Acceptance Criteria Test 1: Security scan job completely removed
test_security_removal() {
    echo "Test AC1: Security scan job completely removed"
    
    # Check job definition
    if grep -q "^  security:$" "$WORKFLOW_FILE"; then
        echo "❌ FAIL: Security job definition still exists"
        return 1
    fi
    
    # Check security tools installation
    if grep -q "govulncheck\|gosec" "$WORKFLOW_FILE"; then
        echo "❌ FAIL: Security tools installation still exists"
        return 1
    fi
    
    # Check make security command
    if grep -q "make security" "$WORKFLOW_FILE"; then
        echo "❌ FAIL: 'make security' command still exists"
        return 1
    fi
    
    echo "✅ PASS: Security scan job completely removed"
    return 0
}

# Acceptance Criteria Test 2: Docker test job completely removed
test_docker_removal() {
    echo "Test AC2: Docker test job completely removed"
    
    # Check job definition
    if grep -q "^  docker:$" "$WORKFLOW_FILE"; then
        echo "❌ FAIL: Docker job definition still exists"
        return 1
    fi
    
    # Check Docker Buildx setup
    if grep -q "docker/setup-buildx-action" "$WORKFLOW_FILE"; then
        echo "❌ FAIL: Docker Buildx setup still exists"
        return 1
    fi
    
    # Check Docker build commands
    if grep -q "docker build\|docker run\|docker-compose" "$WORKFLOW_FILE"; then
        echo "❌ FAIL: Docker commands still exist"
        return 1
    fi
    
    echo "✅ PASS: Docker test job completely removed"
    return 0
}

# Acceptance Criteria Test 3: Essential jobs retained
test_essential_jobs_retained() {
    echo "Test AC3: Essential jobs retained and functional"
    
    ESSENTIAL_JOBS=("lint" "test" "integration" "coverage" "benchmark")
    
    for job in "${ESSENTIAL_JOBS[@]}"; do
        if ! grep -q "^  $job:$" "$WORKFLOW_FILE"; then
            echo "❌ FAIL: Essential job '$job' missing"
            return 1
        fi
    done
    
    # Test job dependencies
    if ! grep -A 5 "integration:" "$WORKFLOW_FILE" | grep -q "needs: test"; then
        echo "❌ FAIL: Integration job missing 'needs: test' dependency"
        return 1
    fi
    
    if ! grep -A 5 "coverage:" "$WORKFLOW_FILE" | grep -q "needs: test"; then
        echo "❌ FAIL: Coverage job missing 'needs: test' dependency"
        return 1
    fi
    
    if ! grep -A 5 "benchmark:" "$WORKFLOW_FILE" | grep -q "needs: test"; then
        echo "❌ FAIL: Benchmark job missing 'needs: test' dependency"
        return 1
    fi
    
    echo "✅ PASS: All essential jobs retained with correct dependencies"
    return 0
}

# Performance validation test
test_workflow_performance() {
    echo "Test AC4: Workflow performance improvement"
    
    JOB_COUNT=$(grep "^  [a-z-]*:$" "$WORKFLOW_FILE" | grep -v "push:" | grep -v "pull_request:" | wc -l)
    
    if [ "$JOB_COUNT" -ne 5 ]; then
        echo "❌ FAIL: Expected 5 jobs, found $JOB_COUNT"
        echo "Jobs found:"
        grep "^  [a-z-]*:$" "$WORKFLOW_FILE" | grep -v "push:" | grep -v "pull_request:"
        return 1
    fi
    
    echo "✅ PASS: Workflow optimized to 5 essential jobs (reduced from 7)"
    return 0
}

# YAML syntax validation
test_yaml_syntax() {
    echo "Test AC5: YAML syntax validation"
    
    # Use python to validate YAML syntax
    if command -v python3 > /dev/null; then
        python3 -c "import yaml; yaml.safe_load(open('$WORKFLOW_FILE', 'r'))" 2>/dev/null
        if [ $? -eq 0 ]; then
            echo "✅ PASS: YAML syntax is valid"
            return 0
        else
            echo "❌ FAIL: YAML syntax errors detected"
            return 1
        fi
    else
        echo "⚠️  SKIP: Python not available for YAML validation"
        return 0
    fi
}

# Run all tests
echo "Running acceptance criteria tests..."
echo ""

TOTAL_TESTS=0
PASSED_TESTS=0

run_test() {
    local test_name=$1
    local test_func=$2
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if $test_func; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    echo ""
}

run_test "Security Removal" test_security_removal
run_test "Docker Removal" test_docker_removal
run_test "Essential Jobs" test_essential_jobs_retained
run_test "Performance" test_workflow_performance
run_test "YAML Syntax" test_yaml_syntax

echo "📊 TEST RESULTS:"
echo "Passed: $PASSED_TESTS/$TOTAL_TESTS"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo "🎉 ALL ACCEPTANCE CRITERIA TESTS PASSED!"
    exit 0
else
    echo "❌ Some tests failed. Review implementation."
    exit 1
fi