CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop existing table to ensure clean migration
DROP TABLE IF EXISTS articles CASCADE;

CREATE TABLE articles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  url TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  summary TEXT,
  embedding vector(1536),
  sentiment VARCHAR(50),
  sentiment_score DECIMAL(3,2) DEFAULT 0.5,
  tone TEXT,
  entities JSONB DEFAULT '[]'::jsonb,
  keywords JSONB DEFAULT '[]'::jsonb,
  topics JSONB DEFAULT '[]'::jsonb,
  url_hash TEXT UNIQUE NOT NULL, -- SHA-256 hash of the URL for caching
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX articles_embedding_idx
  ON articles USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX articles_url_idx ON articles(url);
CREATE INDEX articles_url_hash_idx ON articles(url_hash);

-- Chat request/response cache table
CREATE TABLE chat_cache (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  request_hash TEXT UNIQUE NOT NULL, -- SHA-256 hash of the request
  request_json JSONB NOT NULL,        -- Full request payload
  response_json JSONB NOT NULL,      -- Full response payload
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours')
);

CREATE INDEX chat_cache_request_hash_idx ON chat_cache(request_hash);
CREATE INDEX chat_cache_expires_at_idx ON chat_cache(expires_at);
