package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Connect establishes a PostgreSQL connection.
func Connect(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations executes the database schema SQL to ensure all tables exist.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Base schema (001_schema.sql)
	baseSchema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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

	CREATE TABLE IF NOT EXISTS options (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
		value TEXT NOT NULL CHECK (value IN ('1','2','3','4')),
		label TEXT NOT NULL,
		sort_order INT NOT NULL DEFAULT 0,
		UNIQUE(question_id, value)
	);

	CREATE TABLE IF NOT EXISTS passages (
		id TEXT PRIMARY KEY,
		passage_number INT,
		title TEXT,
		content TEXT NOT NULL,
		level TEXT NOT NULL,
		section TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

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

	CREATE TABLE IF NOT EXISTS package_templates (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		level TEXT NOT NULL,
		section_counts JSONB NOT NULL,
		total_questions INT NOT NULL,
		is_default BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);`

	_, err := db.ExecContext(migrationCtx, baseSchema)
	if err != nil {
		return fmt.Errorf("failed to run base migrations: %w", err)
	}

	// Archive schema (002_archive_schema.sql)
	archiveSchema := `
	-- Extend questions table with archive metadata
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS year INTEGER;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS month INTEGER;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS date_label VARCHAR(20);
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_type INTEGER;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS section_order INTEGER;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS section_title TEXT;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS source_url TEXT;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS is_practice BOOLEAN DEFAULT FALSE;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS point_weight REAL DEFAULT 1;
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_number INTEGER;

	-- Relax constraints for archive data compatibility
	ALTER TABLE questions DROP CONSTRAINT IF EXISTS questions_section_check;
	ALTER TABLE questions ADD CONSTRAINT questions_section_check 
	    CHECK (section IN ('grammar','reading','listening','vocabulary','vocab'));

	ALTER TABLE questions ALTER COLUMN answer_value DROP NOT NULL;

	ALTER TABLE options DROP CONSTRAINT IF EXISTS options_value_check;
	ALTER TABLE options ADD CONSTRAINT options_value_check 
	    CHECK (value IN ('1','2','3','4'));

	-- Exams table
	CREATE TABLE IF NOT EXISTS exams (
	    id          TEXT PRIMARY KEY,
	    level       VARCHAR(2) NOT NULL,
	    year        INTEGER NOT NULL,
	    month       INTEGER NOT NULL,
	    date_label  VARCHAR(20) NOT NULL,
	    is_practice BOOLEAN DEFAULT FALSE,
	    created_at  TIMESTAMPTZ DEFAULT NOW(),
	    UNIQUE(level, date_label)
	);

	-- Link questions to exams
	ALTER TABLE questions ADD COLUMN IF NOT EXISTS exam_id TEXT REFERENCES exams(id);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_questions_exam        ON questions(exam_id);
	CREATE INDEX IF NOT EXISTS idx_questions_year        ON questions(year);
	CREATE INDEX IF NOT EXISTS idx_questions_date_label  ON questions(date_label);
	CREATE INDEX IF NOT EXISTS idx_questions_chrono      ON questions(level, year, month, section_order);
	CREATE INDEX IF NOT EXISTS idx_exams_level_date      ON exams(level, year, month);
	CREATE INDEX IF NOT EXISTS idx_questions_practice    ON questions(is_practice);
	CREATE INDEX IF NOT EXISTS idx_questions_qtype       ON questions(question_type);
	`

	_, err = db.ExecContext(migrationCtx, archiveSchema)
	if err != nil {
		return fmt.Errorf("failed to run archive migrations: %w", err)
	}

	return nil
}
