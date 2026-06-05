package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/philiaspace/mondaiphi/internal/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
)

// ============================================================
// ARCHIVE METHODS — chronological exam browsing
// ============================================================

// ListExams returns exams for a given level, most recent first.
func (r *QuestionRepository) ListExams(ctx context.Context, level examd.JLPTLevel, limit int) ([]domain.Exam, error) {
	query := `SELECT id, level, year, month, date_label, is_practice FROM exams WHERE 1=1`
	args := []interface{}{}
	argN := 1

	if level != "" {
		query += fmt.Sprintf(" AND level = $%d", argN)
		args = append(args, string(level))
		argN++
	}

	query += fmt.Sprintf(" ORDER BY year DESC, month DESC LIMIT $%d", argN)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exams []domain.Exam
	for rows.Next() {
		var e domain.Exam
		if err := rows.Scan(&e.ID, &e.Level, &e.Year, &e.Month, &e.DateLabel, &e.IsPractice); err != nil {
			return nil, err
		}
		exams = append(exams, e)
	}
	return exams, rows.Err()
}

// FindQuestionsByExam returns all questions for a given exam, ordered by section_order then id.
func (r *QuestionRepository) FindQuestionsByExam(ctx context.Context, examID string) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, level, section, prompt, context, answer_value, answer_note,
		       passage_id, source_group_key, version,
		       year, month, date_label, question_type, section_order, section_title,
		       source_url, is_practice, point_weight
		FROM questions WHERE exam_id = $1
		ORDER BY section_order, id
	`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanQuestionsFull(rows)
}

// FindExamByDate returns a single exam by level and date_label.
func (r *QuestionRepository) FindExamByDate(ctx context.Context, level examd.JLPTLevel, dateLabel string) (*domain.Exam, error) {
	var e domain.Exam
	err := r.db.QueryRowContext(ctx, `
		SELECT id, level, year, month, date_label, is_practice
		FROM exams WHERE level = $1 AND date_label = $2
	`, level, dateLabel).Scan(&e.ID, &e.Level, &e.Year, &e.Month, &e.DateLabel, &e.IsPractice)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ============================================================
// BULK IMPORT — optimized for archive CSV ingestion
// ============================================================

// ArchiveQuestion is a raw question from the archive CSV, used for bulk import.
type ArchiveQuestion struct {
	Level          string
	Year           int
	Month          int
	DateLabel      string
	Section        string
	IsPractice     bool
	SourceURL      string
	SectionOrder   int
	SectionTitle   string
	QuestionNumber int
	PassageText    string
	QuestionText   string
	Option1        string
	Option2        string
	Option3        string
	Option4        string
	CorrectAnswer  int    // 1-4, 0 if unknown
	QuestionType   int
	PointWeight    float64
	Explanation    string
	AudioPath      string
	ImagePaths     []string
}

// BulkImport inserts questions, options, exams, and assets in a single transaction.
func (r *QuestionRepository) BulkImport(ctx context.Context, archiveQuestions []ArchiveQuestion) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Track unique exams to insert
	examCache := make(map[string]string) // key "N1:2025-12" -> exam_id

	imported := 0
	for _, aq := range archiveQuestions {
		// Skip questions with missing critical data
		if aq.QuestionText == "" && aq.PassageText == "" {
			continue
		}

		// Ensure exam exists
		examKey := fmt.Sprintf("%s:%s", aq.Level, aq.DateLabel)
		examID, ok := examCache[examKey]
		if !ok {
			examID = fmt.Sprintf("exm_%s_%s", strings.ToLower(aq.Level), strings.ReplaceAll(aq.DateLabel, "-", "_"))
			_, err := tx.ExecContext(ctx, `
				INSERT INTO exams (id, level, year, month, date_label, is_practice)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (level, date_label) DO NOTHING
			`, examID, aq.Level, aq.Year, aq.Month, aq.DateLabel, aq.IsPractice)
			if err != nil {
				return imported, fmt.Errorf("insert exam %s: %w", examID, err)
			}
			examCache[examKey] = examID
		}

		// Generate question ID
		qID := fmt.Sprintf("qst_%s_%s_%02d", strings.ToLower(aq.Level), aq.DateLabel, aq.QuestionNumber)

		// Normalize section name
		section := aq.Section
		if section == "vocab" {
			section = "vocabulary"
		}

		answerValue := ""
		if aq.CorrectAnswer > 0 && aq.CorrectAnswer <= 4 {
			answerValue = fmt.Sprintf("%d", aq.CorrectAnswer)
		}

		// Insert question
		_, err := tx.ExecContext(ctx, `
			INSERT INTO questions (
				id, level, section, prompt, context, answer_value, answer_note,
				year, month, date_label, question_type, section_order, section_title,
				source_url, is_practice, point_weight, exam_id, version
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				$8, $9, $10, $11, $12, $13,
				$14, $15, $16, $17, 1
			)
			ON CONFLICT (id) DO UPDATE SET
				prompt = EXCLUDED.prompt,
				context = EXCLUDED.context,
				answer_value = EXCLUDED.answer_value,
				answer_note = EXCLUDED.answer_note,
				year = EXCLUDED.year,
				month = EXCLUDED.month,
				date_label = EXCLUDED.date_label,
				question_type = EXCLUDED.question_type,
				section_order = EXCLUDED.section_order,
				section_title = EXCLUDED.section_title,
				source_url = EXCLUDED.source_url,
				updated_at = NOW()
		`,
			qID, aq.Level, section, aq.QuestionText, nullStr(aq.PassageText),
			nullStr(answerValue), nullStr(aq.Explanation),
			nullInt(aq.Year), nullInt(aq.Month), aq.DateLabel, nullInt(aq.QuestionType),
			nullInt(aq.SectionOrder), nullStr(aq.SectionTitle),
			nullStr(aq.SourceURL), aq.IsPractice, aq.PointWeight, examID,
		)
		if err != nil {
			return imported, fmt.Errorf("insert question %s: %w", qID, err)
		}

		// Insert options (if any)
		optionValues := []struct {
			value int
			label string
		}{
			{1, aq.Option1},
			{2, aq.Option2},
			{3, aq.Option3},
			{4, aq.Option4},
		}

		optIdx := 0
		for _, ov := range optionValues {
			if ov.label == "" {
				continue
			}
			optIdx++
			_, err := tx.ExecContext(ctx, `
				INSERT INTO options (question_id, value, label, sort_order)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (question_id, value) DO UPDATE SET label = EXCLUDED.label
			`, qID, fmt.Sprintf("%d", ov.value), ov.label, optIdx)
			if err != nil {
				return imported, fmt.Errorf("insert option for %s: %w", qID, err)
			}
		}

		// Insert audio asset
		if aq.AudioPath != "" {
			assetID := fmt.Sprintf("ast_audio_%s_%s_%02d", strings.ToLower(aq.Level), aq.DateLabel, aq.QuestionNumber)
			_, err := tx.ExecContext(ctx, `
				INSERT INTO assets (id, type, source_url, s3_key, local_path, question_id)
				VALUES ($1, 'audio', $2, $3, $4, $5)
				ON CONFLICT (id) DO NOTHING
			`, assetID, "", aq.AudioPath, aq.AudioPath, qID)
			if err != nil {
				return imported, fmt.Errorf("insert audio asset for %s: %w", qID, err)
			}
		}

		// Insert image assets
		for _, imgPath := range aq.ImagePaths {
			if imgPath == "" {
				continue
			}
			// Extract simple filename as key
			imgKey := fmt.Sprintf("jlpt/images/%s/%s", aq.Level, imgPath)
			assetID := fmt.Sprintf("ast_img_%s_%s_%s", strings.ToLower(aq.Level), aq.DateLabel, sanitizeFilename(imgPath))
			_, err := tx.ExecContext(ctx, `
				INSERT INTO assets (id, type, source_url, s3_key, local_path, question_id)
				VALUES ($1, 'image', $2, $3, $4, $5)
				ON CONFLICT (id) DO NOTHING
			`, assetID, "", imgKey, imgPath, qID)
			if err != nil {
				continue // non-critical
			}
		}

		imported++
	}

	if err := tx.Commit(); err != nil {
		return imported, err
	}

	return imported, nil
}

// ============================================================
// UPDATED SCAN — includes archive fields
// ============================================================

func (r *QuestionRepository) scanQuestionsFull(rows *sql.Rows) ([]domain.Question, error) {
	var questions []domain.Question
	for rows.Next() {
		var q domain.Question
		var passageID, sourceGroupKey, answerNote, context, answerValue sql.NullString
		var sourceURL, sectionTitle sql.NullString
		var year, month, questionType, sectionOrder sql.NullInt32
		var pointWeight sql.NullFloat64
		var isPractice sql.NullBool

		if err := rows.Scan(
			&q.ID, &q.Level, &q.Section, &q.Prompt, &context, &answerValue,
			&answerNote, &passageID, &sourceGroupKey, &q.Version,
			&year, &month, &q.DateLabel, &questionType,
			&sectionOrder, &sectionTitle, &sourceURL,
			&isPractice, &pointWeight,
		); err != nil {
			return nil, err
		}
		q.PassageID = examd.PassageID(passageID.String)
		q.SourceGroupKey = sourceGroupKey.String
		q.AnswerNote = answerNote.String
		q.Context = context.String
		q.AnswerValue = answerValue.String
		q.SourceURL = sourceURL.String
		q.SectionTitle = sectionTitle.String
		if year.Valid {
			q.Year = int(year.Int32)
		}
		if month.Valid {
			q.Month = int(month.Int32)
		}
		if questionType.Valid {
			q.QuestionType = int(questionType.Int32)
		}
		if sectionOrder.Valid {
			q.SectionOrder = int(sectionOrder.Int32)
		}
		if isPractice.Valid {
			q.IsPractice = isPractice.Bool
		}
		if pointWeight.Valid {
			q.PointWeight = pointWeight.Float64
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// ============================================================
// HELPERS
// ============================================================

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt(i int) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	if len(s) > 50 {
		s = s[len(s)-50:]
	}
	return s
}
