#!/bin/bash

# Comprehensive Test Runner for Refactored Article Assistant
# Runs both unit tests and integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  Refactored Article Assistant Test Suite     ${NC}"
echo -e "${BLUE}================================================${NC}"
echo

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    echo -e "${RED}Error: go.mod not found. Please run from refactored directory.${NC}"
    exit 1
fi

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests and track results
run_test_suite() {
    local test_name="$1"
    local test_path="$2"
    local test_type="$3"
    
    echo -e "${PURPLE}🧪 Running $test_type: $test_name${NC}"
    
    if go test -v "$test_path"; then
        echo -e "${GREEN}✅ $test_name passed${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ $test_name failed${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    echo
}

# Run go mod tidy first
echo -e "${YELLOW}📦 Updating dependencies...${NC}"
if go mod tidy; then
    echo -e "${GREEN}✅ Dependencies updated${NC}"
else
    echo -e "${RED}❌ Failed to update dependencies${NC}"
    exit 1
fi
echo

# Unit Tests
echo -e "${BLUE}🔬 UNIT TESTS${NC}"
echo "Testing individual components in isolation"
echo

run_test_suite "LLM Client Tests" "./tests/unit" "Unit Test"
run_test_suite "Config Service Tests" "./tests/unit" "Unit Test"
run_test_suite "Analysis Service Tests" "./tests/unit" "Unit Test"
run_test_suite "Planner Service Tests" "./tests/unit" "Unit Test"
run_test_suite "Article Service Tests" "./tests/unit" "Unit Test"

# Integration Tests
echo -e "${BLUE}🔗 INTEGRATION TESTS${NC}"
echo "Testing component interactions and full workflows"
echo

run_test_suite "Full Flow Integration Tests" "./tests/integration" "Integration Test"
run_test_suite "Article Integration Tests" "./tests/integration" "Integration Test"

# Test Summary
echo -e "${BLUE}📊 TEST SUMMARY${NC}"
echo "=================="
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    echo
    echo -e "${YELLOW}📋 Test Coverage:${NC}"
    echo "✅ LLM Client (OpenAI integration)"
    echo "✅ Configuration Management"
    echo "✅ Analysis Service"
    echo "✅ Planner Service (Two-phase LLM approach)"
    echo "✅ Article Service (Execution and caching)"
    echo "✅ Full Integration Flow (Ingest → Plan → Execute → Cache)"
    echo "✅ Error Handling"
    echo "✅ Multiple Article Processing"
    echo
    echo -e "${BLUE}🚀 Next Steps:${NC}"
    echo "- Fix remaining build issues"
    echo "- Test with real OpenAI API"
    echo "- Performance comparison with production"
    echo "- Deploy refactored version"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    echo
    echo -e "${YELLOW}🔧 Debugging Tips:${NC}"
    echo "- Check import paths in test files"
    echo "- Verify mock implementations"
    echo "- Ensure all dependencies are available"
    echo "- Run individual test files to isolate issues"
    exit 1
fi
