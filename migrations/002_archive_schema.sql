-- MondaiPhi Archive Schema Migration
-- Adds chronological exam support and JLPT archive-specific columns
-- Run: psql -U phi -d mondaiphi -f 002_archive_schema.sql

BEGIN;

-- ============================================================
-- 1. EXTEND questions table with archive metadata
-- ============================================================

ALTER TABLE questions ADD COLUMN IF NOT EXISTS year INTEGER;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS month INTEGER;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS date_label VARCHAR(20);
ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_type INTEGER;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS section_order INTEGER;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS section_title TEXT;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS source_url TEXT;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS is_practice BOOLEAN DEFAULT FALSE;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS point_weight REAL DEFAULT 1;

-- ============================================================
-- 2. RELAX constraints for archive data compatibility
-- ============================================================

-- Allow 'vocabulary' and 'vocab' in section (archive uses 'vocab')
ALTER TABLE questions DROP CONSTRAINT IF EXISTS questions_section_check;
ALTER TABLE questions ADD CONSTRAINT questions_section_check 
    CHECK (section IN ('grammar','reading','listening','vocabulary','vocab'));

-- Allow NULL answer_value for listening 問題3 (no printed options)
ALTER TABLE questions ALTER COLUMN answer_value DROP NOT NULL;

-- Allow 3-option questions (N5 listening sections)
ALTER TABLE options DROP CONSTRAINT IF EXISTS options_value_check;
ALTER TABLE options ADD CONSTRAINT options_value_check 
    CHECK (value IN ('1','2','3','4'));

-- ============================================================
-- 3. EXAMS table — chronological exam catalog
-- ============================================================

CREATE TABLE IF NOT EXISTS exams (
    id          TEXT PRIMARY KEY,                -- exm_ prefix + ULID
    level       VARCHAR(2) NOT NULL,             -- N1-N5
    year        INTEGER NOT NULL,
    month       INTEGER NOT NULL,                -- 7 or 12
    date_label  VARCHAR(20) NOT NULL,            -- "2025-12" or "practice-01_1"
    is_practice BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(level, date_label)
);

-- ============================================================
-- 4. LINK questions to exams
-- ============================================================

ALTER TABLE questions ADD COLUMN IF NOT EXISTS exam_id TEXT REFERENCES exams(id);

-- ============================================================
-- 4b. QUESTION_NUMBER — explicit ordering within section
ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_number INTEGER;

-- 5. INDEXES for chronological & archive queries
-- ============================================================

-- Chronological browsing
CREATE INDEX IF NOT EXISTS idx_questions_exam        ON questions(exam_id);
CREATE INDEX IF NOT EXISTS idx_questions_year        ON questions(year);
CREATE INDEX IF NOT EXISTS idx_questions_date_label  ON questions(date_label);
CREATE INDEX IF NOT EXISTS idx_questions_chrono      ON questions(level, year, month, section_order);
CREATE INDEX IF NOT EXISTS idx_exams_level_date      ON exams(level, year, month);

-- Archive-specific lookups
CREATE INDEX IF NOT EXISTS idx_questions_practice    ON questions(is_practice);
CREATE INDEX IF NOT EXISTS idx_questions_qtype       ON questions(question_type);

COMMIT;
