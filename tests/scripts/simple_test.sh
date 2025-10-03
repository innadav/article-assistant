#!/bin/bash

# Simple API Test Script
API_BASE="http://localhost:8080"

echo "🚀 Simple API Test"
echo "=================="

# Test 1: Summary
echo "Test 1: Summary"
response=$(curl -s -X POST "$API_BASE/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is this article about?",
    "urls": ["https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"]
  }')

echo "Response: $response"
echo "Task: $(echo "$response" | grep -o '"task":"[^"]*"' | cut -d'"' -f4)"
echo ""

# Test 2: Keywords
echo "Test 2: Keywords"
response=$(curl -s -X POST "$API_BASE/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What are the main keywords?",
    "urls": ["https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/"]
  }')

echo "Response: $response"
echo "Task: $(echo "$response" | grep -o '"task":"[^"]*"' | cut -d'"' -f4)"
echo ""

# Test 3: Top Entities
echo "Test 3: Top Entities"
response=$(curl -s -X POST "$API_BASE/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What are the most commonly discussed entities across the articles?"
  }')

echo "Response: $response"
echo "Task: $(echo "$response" | grep -o '"task":"[^"]*"' | cut -d'"' -f4)"
echo ""

echo "✅ Tests completed"
