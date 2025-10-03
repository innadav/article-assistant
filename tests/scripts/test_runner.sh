#!/bin/bash

# Article Assistant Test Runner
# Orchestrates comprehensive testing across all query types

set -e

# Configuration
API_BASE="http://localhost:8080"
RESULTS_DIR="results/comprehensive_test"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/results_$TIMESTAMP.json"
DATA_FILE="resources/data/startup_articles.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Results storage
START_TIME=$(date +%s)

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_test() {
    echo -e "${PURPLE}[TEST]${NC} $1"
}

# Check if server is running
check_server() {
    log_info "Checking if server is running..."
    if ! curl -s "$API_BASE/health" > /dev/null; then
        log_error "Server is not running at $API_BASE"
        log_info "Please start the server with: docker-compose up -d"
        exit 1
    fi
    log_success "Server is running"
}

# Ingest all articles from data file
ingest_all_articles() {
    log_info "Ingesting all articles from $DATA_FILE..."
    
    if [[ ! -f "$DATA_FILE" ]]; then
        log_error "Data file not found: $DATA_FILE"
        exit 1
    fi
    
    local ingested=0
    local failed=0
    
    while IFS= read -r url; do
        # Skip empty lines and comments
        if [[ -z "$url" || "$url" =~ ^# ]]; then
            continue
        fi
        
        log_info "Ingesting: $url"
        
        if curl -s -X POST "$API_BASE/ingest" \
            -H "Content-Type: application/json" \
            -d "{\"url\": \"$url\"}" | grep -q "success"; then
            ((ingested++))
            log_success "✅ Ingested: $url"
        else
            ((failed++))
            log_error "❌ Failed: $url"
        fi
        
        # Small delay to avoid overwhelming the server
        sleep 1
        
    done < "$DATA_FILE"
    
    log_info "Ingestion complete: $ingested success, $failed failed"
    
    if [[ $failed -gt 0 ]]; then
        log_warning "Some articles failed to ingest. Tests may have limited data."
    fi
}

# Run single query tests
run_single_query_tests() {
    log_info "Running single query tests..."
    
    if [[ ! -f "tests/scripts/single_query_tests.sh" ]]; then
        log_error "Single query test file not found"
        return 1
    fi
    
    chmod +x "tests/scripts/single_query_tests.sh"
    source "tests/scripts/single_query_tests.sh"
    
    run_single_tests
}

# Run multi query tests
run_multi_query_tests() {
    log_info "Running multi-query tests..."
    
    if [[ ! -f "tests/scripts/multi_query_tests.sh" ]]; then
        log_error "Multi query test file not found"
        return 1
    fi
    
    chmod +x "tests/scripts/multi_query_tests.sh"
    source "tests/scripts/multi_query_tests.sh"
    
    run_multi_tests
}

# Save results to JSON file
save_results() {
    log_info "Saving test results..."
    
    local end_time=$(date +%s)
    local total_time=$((end_time - START_TIME))
    
    # Create simple JSON results
    local results_json=$(jq -n --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --arg total "$TOTAL_TESTS" --arg passed "$PASSED_TESTS" --arg failed "$FAILED_TESTS" --arg duration "$total_time" '
        {
            timestamp: $timestamp,
            test_summary: {
                total_tests: ($total | tonumber),
                passed_tests: ($passed | tonumber),
                failed_tests: ($failed | tonumber),
                total_time: ($duration | tonumber)
            },
            ingestion: {
                data_file: "resources/data/startup_articles.txt",
                api_base: "http://localhost:8080"
            }
        }
    ')
    
    echo "$results_json" > "$RESULTS_FILE"
    log_success "Results saved to: $RESULTS_FILE"
}

# Generate final report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="$RESULTS_DIR/report_$TIMESTAMP.txt"
    
    cat > "$report_file" << EOF
Article Assistant Comprehensive Test Report
==========================================
Timestamp: $(date)
Total Tests: $TOTAL_TESTS
Passed: $PASSED_TESTS
Failed: $FAILED_TESTS
Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%

Test Categories:
- Single Query Tests: Tests that work with individual articles
- Multi Query Tests: Tests that require 2+ articles for comparison

Data Source: $DATA_FILE
API Base: $API_BASE
Results Directory: $RESULTS_DIR
Results File: $RESULTS_FILE
EOF

    log_success "Report generated: $report_file"
    
    if [[ $FAILED_TESTS -gt 0 ]]; then
        log_warning "Some tests failed. Check the report for details."
        exit 1
    else
        log_success "All tests passed! 🎉"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Article Assistant Test Runner  ${NC}"
    echo -e "${BLUE}================================${NC}"
    echo
    
    # Pre-flight checks
    check_server
    
    # Ingest all articles
    ingest_all_articles
    
    echo
    log_info "Starting comprehensive tests..."
    echo
    
    # Run test suites
    run_single_query_tests
    run_multi_query_tests
    
    echo
    log_info "Tests completed!"
    
    # Save results to JSON file
    save_results
    
    # Generate final report
    generate_report
}

# Export functions for use in other scripts
export -f log_info log_success log_error log_warning log_test
export API_BASE RESULTS_DIR TIMESTAMP RESULTS_FILE
export TOTAL_TESTS PASSED_TESTS FAILED_TESTS

# Run main function
main "$@"
