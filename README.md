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
```

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

```json
{
  "url": "https://example.com/article"
}
```

### POST /chat
Chat-based queries with natural language.

```json
{
  "query": "What is this article about?",
  "urls": ["https://example.com/article"] // optional
}
```

### GET /health
Health check endpoint.

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
