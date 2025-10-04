#!/bin/bash

# Article Assistant Test Runner with Response Storage
# Captures and stores all API responses for analysis

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
START_TIME=$(date +%s)

# Results storage
TEST_RESULTS=()

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

# Test function that captures responses
test_api_query() {
    local test_name="$1"
    local query="$2"
    local expected_task="$3"
    local should_contain="$4"
    
    log_test "Test: $test_name"
    
    local response
    local start_time
    local end_time
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$query\"}")
    
    end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Parse response to get actual task
    local actual_task=$(echo "$response" | jq -r '.task // "unknown"')
    
    # Check if response contains expected task
    if [[ "$actual_task" == "$expected_task" ]]; then
        log_success "✅ Task correct: $expected_task"
        local status="PASSED"
    else
        log_error "❌ Wrong task. Expected: $expected_task, Got: $actual_task"
        echo "Response: $response"
        local status="FAILED"
        ((FAILED_TESTS++))
    fi
    
    # Check if response contains expected content
    local content_check=""
    if [[ -n "$should_contain" ]]; then
        if echo "$response" | grep -qi "$should_contain"; then
            log_success "✅ Contains expected content: $should_contain"
            content_check="PASSED"
        else
            log_warning "⚠️  Missing expected content: $should_contain"
            content_check="FAILED"
        fi
    fi
    
    # Log timing
    log_info "⏱️  Response time: ${duration}s"
    
    # Store test result
    local test_result=$(jq -n --arg name "$test_name" --arg query "$query" --arg expected_task "$expected_task" --arg actual_task "$actual_task" --arg status "$status" --arg duration "$duration" --arg should_contain "$should_contain" --arg content_check "$content_check" --argjson response "$response" '
        {
            name: $name,
            query: $query,
            expected_task: $expected_task,
            actual_task: $actual_task,
            status: $status,
            duration: ($duration | tonumber),
            should_contain: $should_contain,
            content_check: $content_check,
            response: $response
        }
    ')
    
    TEST_RESULTS+=("$test_result")
    
    if [[ "$status" == "PASSED" ]]; then
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    echo
}

# Test with specific URL
test_api_query_with_url() {
    local test_name="$1"
    local query="$2"
    local url="$3"
    local expected_task="$4"
    
    log_test "Test: $test_name"
    
    local response
    local start_time
    local end_time
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$query\", \"urls\": [\"$url\"]}")
    
    end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Parse response to get actual task
    local actual_task=$(echo "$response" | jq -r '.task // "unknown"')
    
    # Check if response contains expected task
    if [[ "$actual_task" == "$expected_task" ]]; then
        log_success "✅ Task correct: $expected_task"
        local status="PASSED"
    else
        log_error "❌ Wrong task. Expected: $expected_task, Got: $actual_task"
        echo "Response: $response"
        local status="FAILED"
        ((FAILED_TESTS++))
    fi
    
    # Log timing
    log_info "⏱️  Response time: ${duration}s"
    
    # Store test result
    local test_result=$(jq -n --arg name "$test_name" --arg query "$query" --arg url "$url" --arg expected_task "$expected_task" --arg actual_task "$actual_task" --arg status "$status" --arg duration "$duration" --argjson response "$response" '
        {
            name: $name,
            query: $query,
            url: $url,
            expected_task: $expected_task,
            actual_task: $actual_task,
            status: $status,
            duration: ($duration | tonumber),
            response: $response
        }
    ')
    
    TEST_RESULTS+=("$test_result")
    
    if [[ "$status" == "PASSED" ]]; then
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    echo
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
    log_info "=== Single Query Tests ==="
    echo
    
    # Test 1: Summary of specific article
    test_api_query \
        "Summary Test" \
        "Summarize the article https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/" \
        "summary" \
        "summary"
    
    # Test 2: Keywords extraction
    test_api_query \
        "Keywords Test" \
        "Extract keywords from the article https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/" \
        "keywords_or_topics" \
        "keywords"
    
    # Test 3: Sentiment analysis
    test_api_query \
        "Sentiment Test" \
        "What is the sentiment of the article https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/?" \
        "get_sentiment" \
        "sentiment"
    
    # Test 4: General article search
    test_api_query \
        "Article positive about Search Test" \
        "Which article is more positive about the topic of AI regulation?" \
        "most_positive_article_for_filter" \
        "articles"
    
    # Test 5: Top entities
    test_api_query \
        "Top Entities Test" \
        "What are the most commonly discussed entities across all articles?" \
        "get_top_entities" \
        "entities"
    
    # Test 6: Economic trends search
    test_api_query \
        "Economic Trends Search Test" \
        "What articles discuss economic trends?" \
        "filter_by_specific_topic" \
        "articles"
    
    log_info "Single query tests completed: $PASSED_TESTS passed, $FAILED_TESTS failed"
    echo
}

# Run multi query tests
run_multi_query_tests() {
    log_info "=== Multi Query Tests ==="
    echo
    
    # Define some test URLs
    local url1="https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/"
    local url2="https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/"
    
    # Test 1: Compare two articles
    log_test "Test: Article Comparison Test"
    
    local response
    local start_time
    local end_time
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"Compare these two articles: $url1 and $url2\"}")
    
    end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    local actual_task=$(echo "$response" | jq -r '.task // "unknown"')
    
    if [[ "$actual_task" == "compare_articles" ]]; then
        log_success "✅ Task correct: compare_articles"
        local status="PASSED"
    else
        log_error "❌ Wrong task. Expected: compare_articles, Got: $actual_task"
        local status="FAILED"
        ((FAILED_TESTS++))
    fi
    
    log_info "⏱️  Response time: ${duration}s"
    
    # Store test result
    local test_result=$(jq -n --arg name "Article Comparison Test" --arg query "Compare these two articles: $url1 and $url2" --arg expected_task "compare_articles" --arg actual_task "$actual_task" --arg status "$status" --arg duration "$duration" --argjson response "$response" '
        {
            name: $name,
            query: $query,
            expected_task: $expected_task,
            actual_task: $actual_task,
            status: $status,
            duration: ($duration | tonumber),
            response: $response
        }
    ')
    
    TEST_RESULTS+=("$test_result")
    
    if [[ "$status" == "PASSED" ]]; then
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    echo
    
    # Test 2: Tone differences between sources
    log_test "Test: Tone Differences Test"
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"What are the key differences in tone between $url1 and $url2\"}")
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    actual_task=$(echo "$response" | jq -r '.task // "unknown"')
    
    if [[ "$actual_task" == "ton_key_differences" ]]; then
        log_success "✅ Task correct: ton_key_differences"
        status="PASSED"
    else
        log_error "❌ Wrong task. Expected: ton_key_differences, Got: $actual_task"
        status="FAILED"
        ((FAILED_TESTS++))
    fi
    
    log_info "⏱️  Response time: ${duration}s"
    
    # Store test result
    test_result=$(jq -n --arg name "Tone Differences Test" --arg query "What are the key differences in tone between $url1 and $url2" --arg expected_task "ton_key_differences" --arg actual_task "$actual_task" --arg status "$status" --arg duration "$duration" --argjson response "$response" '
        {
            name: $name,
            query: $query,
            expected_task: $expected_task,
            actual_task: $actual_task,
            status: $status,
            duration: ($duration | tonumber),
            response: $response
        }
    ')
    
    TEST_RESULTS+=("$test_result")
    
    if [[ "$status" == "PASSED" ]]; then
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    echo
    
    log_info "Multi query tests completed: $PASSED_TESTS passed, $FAILED_TESTS failed"
    echo
}

# Save results to JSON file
save_results() {
    log_info "Saving test results with responses..."
    
    local end_time=$(date +%s)
    local total_time=$((end_time - START_TIME))
    
    # Create JSON array from test results
    local tests_json="["
    for i in "${!TEST_RESULTS[@]}"; do
        if [[ $i -gt 0 ]]; then
            tests_json+=","
        fi
        tests_json+="${TEST_RESULTS[$i]}"
    done
    tests_json+="]"
    
    # Create final results JSON
    local results_json=$(jq -n --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --arg total "$TOTAL_TESTS" --arg passed "$PASSED_TESTS" --arg failed "$FAILED_TESTS" --arg duration "$total_time" --argjson tests "$tests_json" '
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
            },
            tests: $tests
        }
    ')
    
    echo "$results_json" > "$RESULTS_FILE"
    log_success "Results with responses saved to: $RESULTS_FILE"
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

# Run main function
main "$@"
