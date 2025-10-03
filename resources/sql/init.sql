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
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX articles_embedding_idx
  ON articles USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX articles_url_idx ON articles(url);
