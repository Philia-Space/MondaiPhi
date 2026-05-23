-- MondaiPhi PostgreSQL Schema
-- Run: psql -U phi -d mondaiphi -f schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Questions table
CREATE TABLE IF NOT EXISTS questions (
    id TEXT PRIMARY KEY,
    level TEXT NOT NULL CHECK (level IN ('N5','N4','N3','N2','N1')),
    section TEXT NOT NULL CHECK (section IN ('grammar','reading','listening')),
    prompt TEXT NOT NULL,
    context TEXT,
    answer_value TEXT NOT NULL CHECK (answer_value IN ('1','2','3','4')),
    answer_note TEXT,
    passage_id TEXT,
    source_group_key TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

-- Options table
CREATE TABLE IF NOT EXISTS options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    value TEXT NOT NULL CHECK (value IN ('1','2','3','4')),
    label TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE(question_id, value)
);

-- Passages table
CREATE TABLE IF NOT EXISTS passages (
    id TEXT PRIMARY KEY,
    passage_number INT,
    title TEXT,
    content TEXT NOT NULL,
    level TEXT NOT NULL,
    section TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Assets table
CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('audio','image')),
    source_url TEXT,
    s3_key TEXT NOT NULL,
    local_path TEXT,
    question_id TEXT REFERENCES questions(id),
    passage_id TEXT REFERENCES passages(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Package templates table
CREATE TABLE IF NOT EXISTS package_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    level TEXT NOT NULL,
    section_counts JSONB NOT NULL,
    total_questions INT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_questions_level_section ON questions(level, section);
CREATE INDEX IF NOT EXISTS idx_questions_passage ON questions(passage_id);
CREATE INDEX IF NOT EXISTS idx_questions_group ON questions(source_group_key);
CREATE INDEX IF NOT EXISTS idx_options_question ON options(question_id);
CREATE INDEX IF NOT EXISTS idx_assets_question ON assets(question_id);
CREATE INDEX IF NOT EXISTS idx_assets_passage ON assets(passage_id);
CREATE INDEX IF NOT EXISTS idx_templates_level ON package_templates(level);
