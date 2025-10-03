#!/bin/bash

# Multi Query Tests
# Tests that require 2+ articles for comparison or analysis

# Test helper function for multi-article queries
test_multi_query() {
    local test_name="$1"
    local query="$2"
    local urls="$3"
    local expected_task="$4"
    local should_contain="$5"
    
    log_test "Test: $test_name"
    
    local response
    local start_time
    local end_time
    
    start_time=$(date +%s.%N)
    
    # Build JSON payload with URLs array
    local json_payload="{\"query\": \"$query\", \"urls\": $urls}"
    
    response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "$json_payload")
    
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

# Main multi query test function
run_multi_tests() {
    log_info "=== Multi Query Tests ==="
    echo
    
    # Define some test URLs (first few from our data)
    local url1="https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/"
    local url2="https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/"
    local url3="https://edition.cnn.com/2025/07/25/tech/meta-ai-superintelligence-team-who-its-hiring"
    local url4="https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/"
    
    # Test 1: Compare two articles
    test_multi_query \
        "Article Comparison Test" \
        "Compare these two articles" \
        "[\"$url1\", \"$url2\"]" \
        "compare_articles" \
        "comparison"
    
    # Test 2: Tone differences between articles
    test_multi_query \
        "Tone Differences Test" \
        "What are the key differences in tone between these articles?" \
        "[\"$url1\", \"$url2\"]" \
        "ton_key_differences" \
        "tone"
    
    # Test 3: Compare three articles
    test_multi_query \
        "Three Article Comparison Test" \
        "Compare these three articles and highlight key differences" \
        "[\"$url1\", \"$url2\", \"$url3\"]" \
        "compare_articles" \
        "comparison"
    
    # Test 4: Sentiment comparison
    test_multi_query \
        "Sentiment Comparison Test" \
        "Which of these articles has the most positive sentiment about AI?" \
        "[\"$url1\", \"$url2\", \"$url3\"]" \
        "get_article" \
        "positive"
    
    # Test 5: Technology focus comparison
    test_multi_query \
        "Technology Focus Test" \
        "Compare the technology focus of these articles" \
        "[\"$url2\", \"$url3\", \"$url4\"]" \
        "compare_articles" \
        "technology"
    
    # Test 6: Business vs Tech comparison
    test_multi_query \
        "Business vs Tech Test" \
        "What are the differences between the business and technology articles?" \
        "[\"$url1\", \"$url2\"]" \
        "compare_articles" \
        "differences"
    
    # Test 7: AI-related articles comparison
    test_multi_query \
        "AI Articles Comparison Test" \
        "Compare these AI-related articles" \
        "[\"$url1\", \"$url2\", \"$url3\"]" \
        "compare_articles" \
        "AI"
    
    # Test 8: Tone analysis across multiple articles
    test_multi_query \
        "Multi-Article Tone Test" \
        "Analyze the tone differences across these articles" \
        "[\"$url1\", \"$url2\", \"$url4\"]" \
        "ton_key_differences" \
        "tone"
    
    # Test 9: Sentiment analysis across articles
    test_multi_query \
        "Multi-Article Sentiment Test" \
        "Which article has the most negative sentiment?" \
        "[\"$url1\", \"$url2\", \"$url3\", \"$url4\"]" \
        "get_article" \
        "sentiment"
    
    # Test 10: Comprehensive comparison
    test_multi_query \
        "Comprehensive Comparison Test" \
        "Provide a comprehensive comparison of these articles covering topics, tone, and sentiment" \
        "[\"$url1\", \"$url2\", \"$url3\"]" \
        "compare_articles" \
        "comprehensive"
    
    log_info "Multi query tests completed: $PASSED_TESTS passed, $FAILED_TESTS failed"
    echo
}
