#!/bin/bash

# Single Query Tests
# Tests that work with individual articles or general queries

# Test helper function
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
    
    # Check if response contains expected task
    if echo "$response" | grep -q "\"task\":\"$expected_task\""; then
        log_success "✅ Task correct: $expected_task"
    else
        log_error "❌ Wrong task. Expected: $expected_task"
        echo "Response: $response"
        ((FAILED_TESTS++))
        return 1
    fi
    
    # Check if response contains expected content
    if [[ -n "$should_contain" ]]; then
        if echo "$response" | grep -qi "$should_contain"; then
            log_success "✅ Contains expected content: $should_contain"
        else
            log_warning "⚠️  Missing expected content: $should_contain"
        fi
    fi
    
    # Log timing
    log_info "⏱️  Response time: ${duration}s"
    
    ((TOTAL_TESTS++))
    ((PASSED_TESTS++))
    
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
    
    # Check if response contains expected task
    if echo "$response" | grep -q "\"task\":\"$expected_task\""; then
        log_success "✅ Task correct: $expected_task"
    else
        log_error "❌ Wrong task. Expected: $expected_task"
        echo "Response: $response"
        ((FAILED_TESTS++))
        return 1
    fi
    
    # Log timing
    log_info "⏱️  Response time: ${duration}s"
    
    ((TOTAL_TESTS++))
    ((PASSED_TESTS++))
    
    echo
}

# Main single query test function
run_single_tests() {
    log_info "=== Single Query Tests ==="
    echo
    
    # Test 1: Summary of specific article
    test_api_query_with_url \
        "Summary Test" \
        "Give me a summary of this article" \
        "https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/" \
        "summary"
    
    # Test 2: Keywords extraction
    test_api_query_with_url \
        "Keywords Test" \
        "Extract keywords from this article" \
        "https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/" \
        "keywords_or_topics"
    
    # Test 3: Sentiment analysis
    test_api_query_with_url \
        "Sentiment Test" \
        "What is the sentiment of this article?" \
        "https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/" \
        "get_sentiment"
    
    # Test 4: General article search
    test_api_query \
        "Article Search Test" \
        "What articles discuss AI technology and innovation?" \
        "filter_by_specific_topic" \
        "articles"
    
    # Test 5: Top entities
    test_api_query \
        "Top Entities Test" \
        "What are the most commonly discussed entities across all articles?" \
        "get_top_entities" \
        "entities"
    
    # Test 6: Topic-based search
    test_api_query \
        "Topic Search Test" \
        "What articles discuss cybersecurity and data protection?" \
        "filter_by_specific_topic" \
        "cyber"
    
    # Test 7: Sentiment-based search
    test_api_query \
        "Sentiment Search Test" \
        "Which article is most positive about technology innovation?" \
        "most_positive_article_for_filter" \
        "positive"
    
    # Test 8: Business articles search
    test_api_query \
        "Business Articles Test" \
        "What business and finance articles are available?" \
        "filter_by_specific_topic" \
        "business"
    
    # Test 9: Tech articles search
    test_api_query \
        "Tech Articles Test" \
        "Show me technology and innovation articles" \
        "filter_by_specific_topic" \
        "tech"
    
    # Test 10: Recent articles
    test_api_query \
        "Recent Articles Test" \
        "What are the most recent tech articles?" \
        "filter_by_specific_topic" \
        "articles"
    
    log_info "Single query tests completed: $PASSED_TESTS passed, $FAILED_TESTS failed"
    echo
}
