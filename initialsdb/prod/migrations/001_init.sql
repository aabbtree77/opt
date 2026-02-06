-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- -----------------------------------------------------
-- LISTINGS
-- -----------------------------------------------------
CREATE TABLE listings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- main content
    body TEXT NOT NULL,

    -- moderation
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    hidden_at TIMESTAMPTZ,

    -- timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- bot / abuse signals
    ip_hash BYTEA NOT NULL, -- sha256(ip + server_salt)

    -- derived anti-spam helpers
    body_length INTEGER GENERATED ALWAYS AS (length(body)) STORED,
    has_links BOOLEAN GENERATED ALWAYS AS (body ~* '(https?://|www\.)') STORED,
    link_count INTEGER GENERATED ALWAYS AS (
    regexp_count(body, '(https?://|www\.)', 1, 'i')
) STORED,

    -- full text search
    body_tsv tsvector GENERATED ALWAYS AS (
        to_tsvector('simple', body)
    ) STORED
);

-- -----------------------------------------------------
-- INDEXES
-- -----------------------------------------------------

-- Primary read path: newest visible listings
CREATE INDEX idx_listings_visible_created_at
ON listings (created_at DESC)
WHERE is_hidden = FALSE;

-- Full text search on visible listings
CREATE INDEX idx_listings_visible_fts
ON listings
USING GIN (body_tsv)
WHERE is_hidden = FALSE;

-- Abuse analysis / rate limiting
CREATE INDEX idx_listings_ip_hash_created_at
ON listings (ip_hash, created_at DESC);

-- Optional: fast moderation queries
CREATE INDEX idx_listings_hidden_at
ON listings (hidden_at)
WHERE is_hidden = TRUE;

-- A tie-breaker index for cursor pagination
CREATE INDEX idx_listings_visible_created_at_id
ON listings (created_at DESC, id DESC)
WHERE is_hidden = FALSE;

-- Make search queries index-only reads
CREATE INDEX idx_listings_visible_created_at_id_cover
ON listings (created_at DESC, id DESC)
INCLUDE (body)
WHERE is_hidden = FALSE;
