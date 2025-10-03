#!/bin/bash

# Comprehensive Article Assistant Test Script
# Tests all query types with timing and result storage

API_BASE="http://localhost:8080"
RESULTS_DIR="results/comprehensive_test"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/results_$TIMESTAMP.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

echo -e "${BLUE}🚀 Comprehensive Article Assistant Test${NC}"
echo "=============================================="
echo "Timestamp: $(date)"
echo "Results will be saved to: $RESULTS_FILE"
echo ""

# Initialize results JSON
cat > "$RESULTS_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "test_summary": {
    "total_queries": 0,
    "successful_queries": 0,
    "failed_queries": 0,
    "total_time": 0
  },
  "queries": []
}
EOF

# Function to test API endpoint and measure performance
test_api_query() {
    local query_name="$1"
    local query="$2"
    local urls="$3"
    local expected_task="$4"
    
    echo -e "${YELLOW}Testing: $query_name${NC}"
    echo "Query: $query"
    echo "URLs: $urls"
    
    # Prepare JSON payload
    local json_payload
    if [ -n "$urls" ]; then
        json_payload="{\"query\": \"$query\", \"urls\": [$urls]}"
    else
        json_payload="{\"query\": \"$query\"}"
    fi
    
    # Measure time and make request
    local start_time=$(date +%s.%N)
    local response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
        -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "$json_payload")
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Parse response
    local http_code=$(echo "$response" | tail -n 2 | head -n 1)
    local curl_time=$(echo "$response" | tail -n 1)
    local json_response=$(echo "$response" | sed '$d' | sed '$d')
    
    # Extract task from response (only the first occurrence, not from plan)
    local actual_task=$(echo "$json_response" | jq -r '.task')
    local answer=$(echo "$json_response" | jq -r '.answer')
    local sources_count=$(echo "$json_response" | jq '.sources | length')
    
    # Check if test passed
    local status="FAILED"
    if [ "$http_code" = "200" ] && [ "$actual_task" = "$expected_task" ]; then
        echo -e "${GREEN}✅ PASSED${NC}"
        status="PASSED"
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo "HTTP Code: $http_code"
        echo "Expected Task: $expected_task, Got: $actual_task"
    fi
    
    echo "Response Time: ${curl_time}s"
    echo "Duration: ${duration}s"
    echo "Answer Length: ${#answer} chars"
    echo "Sources Count: $sources_count"
    echo "Answer: $answer"
    echo "---"
    echo ""
    
    # Add to results JSON
    local temp_file=$(mktemp)
    jq --arg name "$query_name" \
       --arg query "$query" \
       --arg urls "$urls" \
       --arg status "$status" \
       --arg http_code "$http_code" \
       --arg actual_task "$actual_task" \
       --arg expected_task "$expected_task" \
       --arg answer "$answer" \
       --arg response_time "$curl_time" \
       --arg duration "$duration" \
       --argjson sources_count "$sources_count" \
       '.queries += [{
         "name": $name,
         "query": $query,
         "urls": $urls,
         "status": $status,
         "http_code": $http_code,
         "actual_task": $actual_task,
         "expected_task": $expected_task,
         "answer": $answer,
         "response_time_seconds": ($response_time | tonumber),
         "duration_seconds": ($duration | tonumber),
         "answer_length": ($answer | length),
         "sources_count": $sources_count
       }]' "$RESULTS_FILE" > "$temp_file" && mv "$temp_file" "$RESULTS_FILE"
    
    # Return status for summary
    echo "$status"
}

# Function to get random URLs for comparison
get_random_urls() {
    local count=$1
    local urls=()
    
    # Read URLs from data file
    while IFS= read -r line; do
        line=$(echo "$line" | xargs) # trim whitespace
        if [[ "$line" =~ ^https?:// ]]; then
            urls+=("$line")
        fi
    done < "resources/data/startup_articles.txt"
    
    # Shuffle and take first $count (macOS compatible)
    local shuffled=($(printf '%s\n' "${urls[@]}" | sort -R))
    local result=()
    for ((i=0; i<count && i<${#shuffled[@]}; i++)); do
        result+=("\"${shuffled[i]}\"")
    done
    
    # Join with commas
    printf '%s,' "${result[@]}" | sed 's/,$//'
}

# Check if API is running
echo -e "${YELLOW}Checking if API is running...${NC}"
if ! curl -s "$API_BASE/health" > /dev/null; then
    echo -e "${RED}❌ API is not running at $API_BASE${NC}"
    echo "Please start the API with: docker-compose up -d"
    exit 1
fi
echo -e "${GREEN}✅ API is running${NC}"
echo ""

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ jq is required but not installed${NC}"
    echo "Please install jq: brew install jq (on macOS) or apt-get install jq (on Ubuntu)"
    exit 1
fi

# Test Results Array
declare -a test_results=()
total_start_time=$(date +%s.%N)

echo -e "${YELLOW}Starting Comprehensive Tests...${NC}"
echo ""

# Get URLs from data file for testing
echo "Reading URLs from resources/data/startup_articles.txt..."
urls_array=()
while IFS= read -r line; do
    line=$(echo "$line" | xargs) # trim whitespace
    if [[ "$line" =~ ^https?:// ]]; then
        urls_array+=("$line")
    fi
done < "resources/data/startup_articles.txt"

echo "Found ${#urls_array[@]} URLs in data file"
echo ""

# Test 1: Summary of specific article
if [ ${#urls_array[@]} -gt 0 ]; then
    result=$(test_api_query \
        "Article Summary" \
        "Give me a summary of ${urls_array[0]}" \
        "\"${urls_array[0]}\"" \
        "summary")
    test_results+=("Summary: $result")
fi

# Test 2: Extract keywords
if [ ${#urls_array[@]} -gt 0 ]; then
    result=$(test_api_query \
        "Keywords Extraction" \
        "Extract keywords from ${urls_array[0]}" \
        "\"${urls_array[0]}\"" \
        "keywords_or_topics")
    test_results+=("Keywords: $result")
fi

# Test 3: Sentiment analysis
if [ ${#urls_array[@]} -gt 0 ]; then
    result=$(test_api_query \
        "Sentiment Analysis" \
        "What is the sentiment of ${urls_array[0]}?" \
        "\"${urls_array[0]}\"" \
        "get_sentiment")
    test_results+=("Sentiment: $result")
fi

# Test 4: Compare multiple articles (random 2)
if [ ${#urls_array[@]} -ge 2 ]; then
    random_urls=$(get_random_urls 2)
    result=$(test_api_query \
        "Article Comparison" \
        "Compare ${urls_array[0]} and ${urls_array[1]}" \
        "$random_urls" \
        "compare_articles")
    test_results+=("Compare: $result")
fi

# Test 5: Tone differences between two sources
if [ ${#urls_array[@]} -ge 2 ]; then
    random_urls=$(get_random_urls 2)
    result=$(test_api_query \
        "Tone Differences" \
        "What are the key differences in tone between ${urls_array[0]} and ${urls_array[1]}?" \
        "$random_urls" \
        "ton_key_differences")
    test_results+=("Tone Differences: $result")
fi

# Test 6: Search by topic (economic trends)
result=$(test_api_query \
    "Topic Search - Economic Trends" \
    "What articles discuss economic trends?" \
    "" \
    "get_list_articles")
test_results+=("Search: $result")

# Test 7: Tone differences between two sources (duplicate test - removing)
# This was a duplicate of Test 5, so removing it

# Test 7: Top entities across all articles
result=$(test_api_query \
    "Top Entities" \
    "What are the most commonly discussed entities across the articles?" \
    "" \
    "get_top_entities")
test_results+=("Top Entities: $result")

# Test 8: AI regulation more positive (specific query as per requirements)
if [ ${#urls_array[@]} -ge 2 ]; then
    random_urls=$(get_random_urls 2)
    result=$(test_api_query \
        "AI Regulation More Positive" \
        "Which article is more positive about the topic of AI regulation?" \
        "$random_urls" \
        "get_article")
    test_results+=("AI Regulation More Positive: $result")
fi

# Calculate total time
total_end_time=$(date +%s.%N)
total_duration=$(echo "$total_end_time - $total_start_time" | bc)

# Summary
echo -e "${YELLOW}📊 Test Summary${NC}"
echo "==============="

passed=0
total=${#test_results[@]}

for result in "${test_results[@]}"; do
    echo "$result"
    if [[ "$result" == *"PASSED"* ]]; then
        ((passed++))
    fi
done

echo ""
echo "Results: $passed/$total tests passed"
echo "Total test duration: ${total_duration}s"

if [ $passed -eq $total ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
else
    echo -e "${RED}⚠️  Some tests failed${NC}"
fi

# Update final summary in JSON
temp_file=$(mktemp)
jq --argjson total "$total" \
   --argjson passed "$passed" \
   --argjson failed "$((total - passed))" \
   --argjson total_time "$total_duration" \
   '.test_summary.total_queries = $total |
    .test_summary.successful_queries = $passed |
    .test_summary.failed_queries = $failed |
    .test_summary.total_time = $total_time' "$RESULTS_FILE" > "$temp_file" && mv "$temp_file" "$RESULTS_FILE"

echo ""
echo -e "${BLUE}📁 Detailed results saved to: $RESULTS_FILE${NC}"
echo -e "${BLUE}To view results: cat $RESULTS_FILE | jq .${NC}"
echo -e "${BLUE}To view summary: cat $RESULTS_FILE | jq .test_summary${NC}"
