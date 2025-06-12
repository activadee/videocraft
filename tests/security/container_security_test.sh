#!/bin/bash

# Container Security Test Suite
# Tests for Issue #13: Enforce Container Security Context

# Note: set -e is commented out to allow all tests to run and generate complete PASS/FAIL summary
# set -e

TEST_IMAGE="videocraft:security-test"
TEST_CONTAINER="videocraft-security-test"
PASS_COUNT=0
FAIL_COUNT=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test helper functions
test_start() {
    echo -e "${YELLOW}Testing: $1${NC}"
}

test_pass() {
    echo -e "${GREEN}✅ PASS: $1${NC}"
    ((PASS_COUNT++))
}

test_fail() {
    echo -e "${RED}❌ FAIL: $1${NC}"
    ((FAIL_COUNT++))
    return 1
}

# Cleanup function
cleanup() {
    echo "Cleaning up test containers..."
    docker stop "$TEST_CONTAINER" 2>/dev/null || true
    docker rm "$TEST_CONTAINER" 2>/dev/null || true
}

# Build test image
build_test_image() {
    echo "Building test image..."
    docker build -t "$TEST_IMAGE" .
}

# Test 1: Container runs as non-root user
test_non_root_user() {
    test_start "Container runs as non-root user"
    
    # Start container and check user ID
    docker run -d --name "$TEST_CONTAINER" "$TEST_IMAGE" sleep 60
    
    USER_ID=$(docker exec "$TEST_CONTAINER" id -u)
    USER_NAME=$(docker exec "$TEST_CONTAINER" whoami)
    
    if [ "$USER_ID" != "0" ] && [ "$USER_NAME" = "videocraft" ]; then
        test_pass "Container runs as non-root user (UID: $USER_ID, User: $USER_NAME)"
    else
        test_fail "Container is running as root (UID: $USER_ID, User: $USER_NAME)"
    fi
    
    cleanup
}

# Test 2: Security contexts are enforced
test_security_contexts() {
    test_start "Security contexts are enforced in docker-compose"
    
    # Check if docker-compose.yml has security_opt configured
    if grep -q "security_opt" docker-compose.yml; then
        test_pass "Security contexts found in docker-compose.yml"
    else
        test_fail "Security contexts missing in docker-compose.yml"
    fi
}

# Test 3: Read-only root filesystem
test_readonly_filesystem() {
    test_start "Read-only root filesystem is configured"
    
    # Check if docker-compose.yml has read_only configured
    if grep -q "read_only" docker-compose.yml; then
        test_pass "Read-only filesystem configured"
    else
        test_fail "Read-only filesystem not configured"
    fi
}

# Test 4: Capabilities are dropped
test_capability_dropping() {
    test_start "Unnecessary capabilities are dropped"
    
    # Check if docker-compose.yml has cap_drop configured
    if grep -q "cap_drop" docker-compose.yml; then
        test_pass "Capability dropping configured"
    else
        test_fail "Capability dropping not configured"
    fi
}

# Test 5: Resource limits are configured
test_resource_limits() {
    test_start "Resource limits are configured"
    
    # Check if docker-compose.yml has deploy.resources configured
    if grep -A 10 "deploy:" docker-compose.yml | grep -q "resources:"; then
        test_pass "Resource limits configured"
    else
        test_fail "Resource limits not configured"
    fi
}

# Test 6: Container can still function with security restrictions
test_container_functionality() {
    test_start "Container functionality with security restrictions"
    
    # This test will pass only after implementation
    # For now, we expect it to fail
    docker-compose up -d
    sleep 10
    
    # Check if health check passes
    HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' videocraft 2>/dev/null || echo "unhealthy")
    
    if [ "$HEALTH_STATUS" = "healthy" ]; then
        test_pass "Container is functional with security restrictions"
    else
        test_fail "Container is not functional with security restrictions (Health: $HEALTH_STATUS)"
    fi
    
    docker-compose down
}

# Test 7: No privileged mode
test_no_privileged_mode() {
    test_start "Container does not run in privileged mode"
    
    # Check if docker-compose.yml has privileged: false or no privileged setting
    if grep -q "privileged.*true" docker-compose.yml; then
        test_fail "Container runs in privileged mode"
    else
        test_pass "Container does not run in privileged mode"
    fi
}

# Main test execution
main() {
    echo "===========================================" 
    echo "🔒 Container Security Test Suite"
    echo "==========================================="
    echo "Testing Issue #13: Enforce Container Security Context"
    echo ""
    
    # Trap to ensure cleanup
    trap cleanup EXIT
    
    # Build image for testing
    build_test_image
    
    # Run all tests
    test_non_root_user
    test_security_contexts
    test_readonly_filesystem
    test_capability_dropping
    test_resource_limits
    test_no_privileged_mode
    test_container_functionality
    
    echo ""
    echo "==========================================="
    echo "Test Results:"
    echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
    echo -e "${RED}Failed: $FAIL_COUNT${NC}"
    echo "==========================================="
    
    if [ $FAIL_COUNT -gt 0 ]; then
        echo "❌ Tests failed! Security vulnerabilities detected."
        exit 1
    else
        echo "✅ All security tests passed!"
        exit 0
    fi
}

# Run tests
main "$@"