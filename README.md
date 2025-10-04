# Article Assistant

A comprehensive AI-powered article analysis system that ingests, analyzes, and provides intelligent insights about articles through a chat-based API.

## 🚀 Features

- **Article Ingestion**: Automatically downloads, summarizes, and extracts entities/keywords from URLs
- **Chat-based API**: Natural language queries for article analysis
- **Semantic Search**: Vector-based search using pgvector and OpenAI embeddings
- **LLM Analysis**: Advanced entity, keyword, and topic matching using OpenAI GPT
- **Multiple Query Types**: Summary, keywords, sentiment, tone, comparison, search, and more
- **Caching**: Intelligent caching of expensive LLM operations
- **Content Change Detection**: Avoid reprocessing unchanged articles using content hashing and HTTP headers
- **Docker Deployment**: Full containerized deployment with PostgreSQL

## 📊 Supported Query Types

1. **Summary** - "What is this article about?"
2. **Keywords** - "What are the main keywords?"
3. **Sentiment** - "What is the sentiment of this article?"
4. **Tone** - "What is the tone of this article?"
5. **Comparison** - "Compare these articles"
6. **Search** - "What articles discuss AI?"
7. **More Positive** - "Which article is more positive?"
8. **Top Entities** - "What are the most commonly discussed entities?"
9. **Unknown Query** - Proper error handling for unrecognized queries

## 🏗️ Architecture

```
article-assistant/
├── cmd/                    # Application entry points
│   ├── server/            # Main HTTP server
│   └── ingest/            # Standalone ingestion tool
├── internal/              # Internal packages
│   ├── analysis/          # LLM-based analysis service
│   ├── classify/          # Query classification logic
│   ├── domain/            # Domain models and types
│   ├── ingest/            # Article ingestion service
│   ├── llm/               # LLM client interface and implementations
│   ├── nlp/               # NLP utilities (entities, keywords)
│   └── repository/        # Database operations
├── data/                  # Startup articles and reference data
├── tests/                 # All test files and scripts
│   ├── integration/       # Integration tests
│   ├── unit/              # Unit tests
│   └── scripts/           # Test and analysis scripts
├── results/               # Test results and analysis outputs
└── docker-compose.yml     # Docker deployment configuration
```

## 🔄 Content Caching & Change Detection

The Article Assistant implements comprehensive caching mechanisms to avoid unnecessary reprocessing and API calls, significantly reducing costs and improving performance.

### Caching Strategy

**1. Article Ingestion Caching**
- Each article URL is hashed using SHA-256 for unique identification
- Before processing, the system checks if the URL has already been processed
- If the URL exists in the database, ingestion is skipped entirely
- No LLM calls are made for previously processed URLs

**2. Chat API Request/Response Caching**
- All `/chat` API requests are hashed using SHA-256 of the request payload
- Responses are cached for 24 hours with automatic expiration
- Identical queries return cached responses instantly without LLM processing
- Background cleanup removes expired cache entries every hour

### Database Schema

```sql
-- Articles table with URL-based caching
CREATE TABLE articles (
  -- ... existing fields ...
  url_hash TEXT UNIQUE NOT NULL, -- SHA-256 hash of the URL for caching
  -- ... other fields ...
);

-- Chat request/response cache table
CREATE TABLE chat_cache (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  request_hash TEXT UNIQUE NOT NULL, -- SHA-256 hash of the request
  request_json JSONB NOT NULL,        -- Full request payload
  response_json JSONB NOT NULL,      -- Full response payload
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours')
);

CREATE INDEX articles_url_hash_idx ON articles(url_hash);
CREATE INDEX chat_cache_request_hash_idx ON chat_cache(request_hash);
CREATE INDEX chat_cache_expires_at_idx ON chat_cache(expires_at);
```

### Implementation Details

**Article Ingestion Flow:**
1. Calculate URL hash
2. Check if article exists in database
3. Skip processing if found, or proceed with full LLM analysis
4. Store article with URL hash

**Chat API Flow:**
1. Calculate request hash from payload
2. Check cache for existing response
3. Return cached response if found, or process with LLM
4. Cache new response for future requests

**Benefits:**
- **Cost Reduction**: Avoids expensive LLM API calls for duplicate requests
- **Performance**: Instant responses for cached queries
- **Scalability**: Enables efficient bulk processing and high-frequency queries
- **Reliability**: Automatic cache expiration prevents stale data

**Logging:**
- `📄 Article already processed, skipping: [URL]`
- `💾 Cache hit for request hash: [hash]`
- `🔄 Processing new request: [query]`
- `💾 Cached response for request hash: [hash]`

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- OpenAI API key

### 1. Environment Setup

```bash
# Create .env file
echo "OPENAI_API_KEY=your_openai_api_key_here" > .env

# Optional: Configure OpenAI model (default: gpt-4-turbo)
echo "OPENAI_MODEL=gpt-4" >> .env
```

**Supported Models:**
- `gpt-4-turbo` (default) - Latest GPT-4 with higher context limit
- `gpt-3.5-turbo` - Fast and cost-effective
- `gpt-3.5-turbo-16k` - Higher context limit
- `gpt-4` - Most capable, higher cost

### 2. Start Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database with pgvector extension
- Article Assistant API server on port 8080
- Automatic ingestion of startup articles

### 3. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Ingest a new article
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article"}'

# Chat query
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What are the top entities?"}'
```

## ⚙️ Configuration

### Model Configuration

The Article Assistant supports configurable OpenAI models through environment variables:

```bash
# Set model in .env file
OPENAI_MODEL=gpt-4

# Or export in shell
export OPENAI_MODEL=gpt-3.5-turbo-16k
```

**Model Comparison:**

| Model | Context Limit | Speed | Cost | Best For |
|-------|---------------|-------|------|----------|
| `gpt-3.5-turbo` | 4K tokens | Fast | Low | General use, summaries |
| `gpt-3.5-turbo-16k` | 16K tokens | Fast | Low-Medium | Longer articles |
| `gpt-4` | 8K tokens | Medium | High | Complex analysis |
| `gpt-4-turbo` | 128K tokens | Medium | High | Very long content |

**Configuration Examples:**

```bash
# For development/testing (fast, cheap)
OPENAI_MODEL=gpt-3.5-turbo

# For production with longer articles
OPENAI_MODEL=gpt-3.5-turbo-16k

# For maximum quality analysis
OPENAI_MODEL=gpt-4

# For very long content analysis
OPENAI_MODEL=gpt-4-turbo
```

### Database Configuration

```bash
# Custom database URL
DATABASE_URL=postgres://user:pass@localhost:5432/article_assistant
```

## 🧪 Testing

### Run All Tests

```bash
# Unit tests
go test ./tests/unit/ -v

# Integration tests (requires running services)
go test ./tests/integration/ -v

# Performance tests
./tests/scripts/comprehensive_test.sh
```

### Analysis Scripts

```bash
# Analyze all articles with LLM
go run ./tests/scripts/analyze_all_articles.go

# Test analysis service
go run ./tests/scripts/test_analysis_service.go
```

## 📊 Performance

- **Average Response Time**: 0.239s
- **Cached Queries**: < 0.01s (summary, keywords, sentiment, tone)
- **LLM Queries**: 0.5-1.4s (comparison, tone analysis)
- **Semantic Search**: ~0.6s across all articles
- **Database Aggregation**: < 0.01s (top entities)

## 🔧 API Endpoints

### POST /ingest
Ingest a new article from URL.

**Request:**
```json
{
  "url": "https://example.com/article"
}
```

**Success Response:**
```json
{
  "status": "success",
  "message": "URL ingested successfully"
}
```

**Error Response:**
```json
{
  "error": "Failed to ingest URL: invalid URL format"
}
```

**Examples:**
```bash
# Ingest a news article
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{"url": "https://edition.cnn.com/2025/07/27/business/trump-us-eu-trade-deal"}'

# Ingest a tech article
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{"url": "https://techcrunch.com/2025/07/26/ai-startup-funding-news"}'
```

### POST /chat
Chat-based queries with natural language. The system automatically extracts URLs from queries when needed.

**Request:**
```json
{
  "query": "What is this article about?"
}
```

**Success Response:**
```json
{
  "answer": "Trump claims to have made the 'biggest deal ever' as the US and EU outline a trade framework.",
  "sources": [
    {
      "id": "uuid-here",
      "url": "https://edition.cnn.com/2025/07/27/business/trump-us-eu-trade-deal",
      "title": "Trump touts 'biggest deal ever made' as US and EU sketch trade framework"
    }
  ],
  "usage": {
    "tokens": 150,
    "cost": 0.0003
  },
  "task": "summary",
  "response_type": "text",
  "plan": {
    "command": "summary",
    "args": {
      "urls": ["https://edition.cnn.com/2025/07/27/business/trump-us-eu-trade-deal"]
    }
  }
}
```

**Error Response:**
```json
{
  "error": "Failed to create query plan: invalid query format"
}
```

**Query Examples:**

#### Summary Queries
```bash
# Summary with URL in query
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Summary of https://edition.cnn.com/2025/07/27/business/trump-us-eu-trade-deal"}'

# General summary request
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What is this article about?"}'
```

#### Keyword/Topic Extraction
```bash
# Extract keywords from specific article
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What are the main keywords in https://example.com/article?"}'

# Get topics from article
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What topics does this article cover?"}'
```

#### Sentiment Analysis
```bash
# Analyze sentiment
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the sentiment of this article?"}'
```

#### Article Comparison
```bash
# Compare two articles
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Compare https://example.com/article1 and https://example.com/article2"}'
```

#### Search Queries
```bash
# Find articles by topic
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What articles discuss AI?"}'

# Find most positive article about a topic
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Most positive article about AI regulation"}'
```

#### Entity Analysis
```bash
# Get top entities across all articles
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What are the top entities?"}'
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

## ⚠️ Exception Handling

### Common Error Scenarios

#### 1. Invalid URL Format
**Request:**
```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{"url": "invalid-url"}'
```

**Response:**
```json
{
  "error": "Failed to ingest URL: invalid URL format"
}
```

#### 2. Article Not Found
**Request:**
```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Summary of https://nonexistent.com/article"}'
```

**Response:**
```json
{
  "answer": "Article not found: https://nonexistent.com/article",
  "sources": null,
  "usage": {
    "tokens": 0,
    "cost": 0
  },
  "task": "summary",
  "response_type": "",
  "plan": {
    "command": "summary",
    "args": {
      "urls": ["https://nonexistent.com/article"]
    }
  }
}
```

#### 3. Missing Query Parameter
**Request:**
```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Response:**
```json
{
  "error": "Invalid request body"
}
```

#### 4. LLM Planning Failure
**Request:**
```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Invalid query format that breaks LLM parsing"}'
```

**Response:**
```json
{
  "error": "Failed to create query plan: failed to parse plan JSON"
}
```

#### 5. Unsupported Command
**Request:**
```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Execute unsupported command"}'
```

**Response:**
```json
{
  "answer": "Command not supported: unsupported_command",
  "task": "unsupported_command",
  "sources": null,
  "usage": {
    "tokens": 0,
    "cost": 0
  }
}
```

### Error Response Format

All error responses follow this structure:
```json
{
  "error": "Error message describing what went wrong",
  "status_code": 400,
  "timestamp": "2025-01-03T12:00:00Z"
}
```

### HTTP Status Codes

- **200 OK**: Successful request
- **400 Bad Request**: Invalid request format or parameters
- **405 Method Not Allowed**: Wrong HTTP method
- **500 Internal Server Error**: Server-side error (LLM failure, database issues, etc.)

### Rate Limiting

The API implements intelligent caching to prevent excessive LLM calls:
- **Cached Responses**: Returned instantly (< 0.01s)
- **New Queries**: Processed with LLM (0.5-1.4s)
- **Cache Duration**: 24 hours for chat responses
- **Background Cleanup**: Expired cache entries removed every hour

## 🗄️ Database Schema

```sql
CREATE TABLE articles (
    id UUID PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    embedding vector(1536),
    sentiment VARCHAR(50),
    tone TEXT,
    entities TEXT[],
    keywords TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🔍 LLM Analysis Service

Advanced analysis capabilities:
- **Entity Matching**: Match content against reference entities with confidence scores
- **Keyword Matching**: Find relevant keywords with relevance scores
- **Topic Matching**: Identify topics with descriptions and scores
- **Similarity Analysis**: Find similar content based on semantic similarity
- **Comprehensive Analysis**: Full content analysis with structured output

## 📈 Results

The system has been tested with 17 articles and shows:
- **100% Success Rate** in processing
- **Top Entities**: AI (6 mentions), United States (3), Intel (2)
- **Top Keywords**: technology (10), artificial intelligence (5)
- **Top Topics**: Technology Innovation (7), Business Strategy (4)

## 🛠️ Development

### Local Development

```bash
# Install dependencies
go mod download

# Run server locally
go run ./cmd/server

# Run ingestion
go run ./cmd/ingest
```

### Adding New Query Types

1. Add task constant to `internal/domain/domain.go`
2. Add classification pattern to `internal/classify/classify.go`
3. Add handler function to `cmd/server/main.go`
4. Add test cases to `tests/unit/classify_test.go`

## 📝 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📞 Support

For questions or issues, please open a GitHub issue or contact the development team.
