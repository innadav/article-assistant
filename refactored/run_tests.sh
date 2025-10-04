#!/bin/bash

# Test Runner for Refactored Article Assistant
# Tests ingestion and summary generation with the new two-phase LLM approach

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}===============================================${NC}"
echo -e "${BLUE}  Refactored Article Assistant Test Runner   ${NC}"
echo -e "${BLUE}===============================================${NC}"
echo

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    echo -e "${RED}Error: go.mod not found. Please run from refactored directory.${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Running Tests for Refactored Code${NC}"
echo "Testing: Ingestion → Planning → Execution → Database Storage"
echo

# Run unit tests
echo -e "${BLUE}🧪 Running Unit Tests...${NC}"
if go test -v ./internal/planner/...; then
    echo -e "${GREEN}✅ Planner tests passed${NC}"
else
    echo -e "${RED}❌ Planner tests failed${NC}"
    exit 1
fi

if go test -v ./internal/article/...; then
    echo -e "${GREEN}✅ Article service tests passed${NC}"
else
    echo -e "${RED}❌ Article service tests failed${NC}"
    exit 1
fi

echo

# Run integration tests
echo -e "${BLUE}🔗 Running Integration Tests...${NC}"
if go test -v ./internal/article/ -run TestArticleService_IngestAndSummarize_Integration; then
    echo -e "${GREEN}✅ Ingestion and summary integration test passed${NC}"
else
    echo -e "${RED}❌ Integration test failed${NC}"
    exit 1
fi

if go test -v ./internal/article/ -run TestArticleService_ProcessInitialArticles; then
    echo -e "${GREEN}✅ Batch processing test passed${NC}"
else
    echo -e "${RED}❌ Batch processing test failed${NC}"
    exit 1
fi

if go test -v ./internal/article/ -run TestArticleService_DatabasePersistence; then
    echo -e "${GREEN}✅ Database persistence test passed${NC}"
else
    echo -e "${RED}❌ Database persistence test failed${NC}"
    exit 1
fi

echo

# Test summary
echo -e "${BLUE}📊 Test Summary${NC}"
echo "✅ Unit Tests: Planner Service"
echo "✅ Unit Tests: Article Service" 
echo "✅ Integration Test: Ingest and Summarize"
echo "✅ Integration Test: Batch Processing"
echo "✅ Integration Test: Database Persistence"
echo

echo -e "${GREEN}🎉 All refactored code tests passed!${NC}"
echo
echo -e "${YELLOW}📋 What was tested:${NC}"
echo "1. 📥 Article ingestion and storage"
echo "2. 🧠 Two-phase LLM planning (intent recognition)"
echo "3. ⚡ Plan execution and summary generation"
echo "4. 💾 Summary caching in database"
echo "5. 🔄 Cache hit behavior (no duplicate LLM calls)"
echo "6. 📦 Batch article processing"
echo "7. 💽 Data persistence simulation"
echo

echo -e "${BLUE}🚀 Next Steps:${NC}"
echo "- Fix remaining build issues"
echo "- Test with real OpenAI API"
echo "- Compare performance with production system"
echo "- Deploy refactored version"
