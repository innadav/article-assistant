# Article Chat System - Refactored Architecture

This is a refactored version of the article assistant system using a **two-phase LLM approach** with OpenAI's GPT models.

## 🏗️ Architecture Overview

### Query Planner Flow
1. **User Query**: Natural language request
2. **Query Planner (LLM Call #1)**: Creates structured execution plan
3. **Structured Plan**: JSON object with intent, targets, parameters
4. **Action Executor**: Executes plan based on intent
5. **Data Fetching**: Retrieves articles from in-memory store
6. **Synthesizer (LLM Call #2)**: Generates final response (optional)
7. **Final Response**: Returns answer to user

### Key Components

- **`planner/`**: The "brain" that creates execution plans
- **`llm/`**: Manages interaction with OpenAI's GPT API
- **`article/`**: Handles article processing and storage
- **`analysis/`**: Text analysis services
- **`transport/http/`**: HTTP API handlers

## 🚀 Features

- **Two-Phase LLM Approach**: Plan → Execute pattern
- **Structured Intent System**: Explicit intent mapping
- **Context-Aware Planning**: Full article context in planning phase
- **Modern LLM Integration**: GPT-4 Turbo with JSON mode
- **In-Memory Storage**: Fast article retrieval

## 📋 Supported Intents

- `SUMMARIZE`: Article summarization
- `KEYWORDS`: Keyword extraction
- `SENTIMENT`: Sentiment analysis
- `COMPARE_TONE`: Tone comparison between articles
- `FIND_BY_TOPIC`: Topic-based article search
- `COMPARE_POSITIVITY`: Positivity comparison
- `UNKNOWN`: Fallback for unclear queries

## 🛠️ Setup

### Prerequisites
- Go 1.21+
- OpenAI API key

### Environment Variables
```bash
export OPENAI_API_KEY="your_openai_api_key_here"
export PORT="8080"  # Optional, defaults to 8080
```

### Running Locally
```bash
cd tests/refactored
go mod tidy
go run ./cmd/server
```

### Running with Docker
```bash
cd tests/refactored
docker build -t article-chat-refactored .
docker run -p 8080:8080 -e OPENAI_API_KEY=$OPENAI_API_KEY article-chat-refactored
```

## 📡 API Endpoints

### POST /chat
Main chat endpoint implementing the planner-executor flow.

**Request:**
```json
{
  "query": "summarize the article about Tesla"
}
```

**Response:**
```json
{
  "answer": "This is a generated summary for 'Sample Title'."
}
```

### POST /add-article
Add a new article to the system.

**Request:**
```json
{
  "url": "https://example.com/article",
  "title": "Article Title",
  "content": "Article content..."
}
```

## 🔄 Comparison with Current System

| Aspect | Current System | Refactored System |
|--------|---------------|------------------|
| **LLM Calls** | 1 per request | 2 per request (Plan + Execute) |
| **Provider** | OpenAI GPT | Google Gemini |
| **Architecture** | Command Pattern | Planner-Executor Pattern |
| **Context** | Limited | Full article context |
| **Cost** | Lower | Higher (2x LLM calls) |
| **Latency** | Lower | Higher (additional round-trip) |
| **Flexibility** | Good | Excellent |
| **Maintainability** | Good | Complex |

## 🎯 Trade-offs

### Advantages
- ✅ Better context awareness
- ✅ More structured intent system
- ✅ Easier to extend with new intents
- ✅ Clear separation of concerns

### Disadvantages
- ❌ Higher cost (2x LLM calls)
- ❌ Increased latency
- ❌ More complex architecture
- ❌ Potential over-engineering

## 🧪 Testing

This refactored version is designed for comparison and experimentation. The current production system remains in the main directory with comprehensive test coverage.
