CREATE TABLE articles (
    url TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    excerpt TEXT,
    summary TEXT,
    sentiment TEXT,
    topics TEXT[],
    processed_at TIMESTAMPTZ NOT NULL
);
