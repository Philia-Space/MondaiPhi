package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/philiaspace/mondaiphi/internal/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
	"github.com/philiaspace/phi-utils/errors"
)

// QuestionRepository implements domain.QuestionRepository using PostgreSQL.
type QuestionRepository struct {
	db *sql.DB
}

// NewQuestionRepository creates a new PostgreSQL-backed question repository.
func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// FindByID implements domain.Repository.
func (r *QuestionRepository) FindByID(ctx context.Context, id string) (*domain.Question, error) {
	var q domain.Question
	var passageID sql.NullString
	var sourceGroupKey sql.NullString
	var answerNote sql.NullString
	var context sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, level, section, prompt, context, answer_value, answer_note, passage_id, source_group_key, version
		FROM questions WHERE id = $1
	`, id).Scan(
		&q.ID, &q.Level, &q.Section, &q.Prompt, &context, &q.AnswerValue,
		&answerNote, &passageID, &sourceGroupKey, &q.Version,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("NOT_FOUND", fmt.Sprintf("question not found: %s", id))
	}
	if err != nil {
		return nil, err
	}

	q.PassageID = examd.PassageID(passageID.String)
	q.SourceGroupKey = sourceGroupKey.String
	q.AnswerNote = answerNote.String
	q.Context = context.String

	return &q, nil
}

// Save implements domain.Repository.
func (r *QuestionRepository) Save(ctx context.Context, q *domain.Question) error {
	// Upsert with optimistic concurrency
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO questions (id, level, section, prompt, context, answer_value, answer_note, passage_id, source_group_key, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			level = EXCLUDED.level,
			section = EXCLUDED.section,
			prompt = EXCLUDED.prompt,
			context = EXCLUDED.context,
			answer_value = EXCLUDED.answer_value,
			answer_note = EXCLUDED.answer_note,
			passage_id = EXCLUDED.passage_id,
			source_group_key = EXCLUDED.source_group_key,
			version = questions.version + 1,
			updated_at = NOW()
		WHERE questions.version = $10
	`, q.ID, q.Level, q.Section, q.Prompt, q.Context, q.AnswerValue, q.AnswerNote,
		sqlNullString(string(q.PassageID)), sqlNullString(q.SourceGroupKey), q.Version)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("CONFLICT", "optimistic concurrency conflict")
	}

	q.Version++
	return nil
}

// Delete implements domain.Repository.
func (r *QuestionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM questions WHERE id = $1`, id)
	return err
}

// FindByLevelAndSection returns questions filtered by JLPT level and section.
func (r *QuestionRepository) FindByLevelAndSection(ctx context.Context, level examd.JLPTLevel, section examd.Section, limit int) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, level, section, prompt, context, answer_value, answer_note, passage_id, source_group_key, version
		FROM questions WHERE level = $1 AND section = $2 LIMIT $3
	`, level, section, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

// FindByPassageID returns questions belonging to a passage.
func (r *QuestionRepository) FindByPassageID(ctx context.Context, id examd.PassageID) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, level, section, prompt, context, answer_value, answer_note, passage_id, source_group_key, version
		FROM questions WHERE passage_id = $1 ORDER BY source_group_key, id
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

// FindBySourceGroupKey returns questions sharing the same source group.
func (r *QuestionRepository) FindBySourceGroupKey(ctx context.Context, key string) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, level, section, prompt, context, answer_value, answer_note, passage_id, source_group_key, version
		FROM questions WHERE source_group_key = $1 ORDER BY id
	`, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

// FindWithOptions returns a question with all its options (for admin/internal use).
func (r *QuestionRepository) FindWithOptions(ctx context.Context, id examd.QuestionID) (*domain.Question, []domain.Option, error) {
	q, err := r.FindByID(ctx, string(id))
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question_id, value, label, sort_order
		FROM options WHERE question_id = $1 ORDER BY sort_order
	`, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var options []domain.Option
	for rows.Next() {
		var o domain.Option
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.Value, &o.Label, &o.SortOrder); err != nil {
			return nil, nil, err
		}
		options = append(options, o)
	}

	return q, options, rows.Err()
}

// FindByIDs returns multiple questions by their IDs, preserving order.
func (r *QuestionRepository) FindByIDs(ctx context.Context, ids []examd.QuestionID) ([]domain.Question, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = string(id)
	}

	query := fmt.Sprintf(`
		SELECT id, level, section, prompt, context, answer_value, answer_note, passage_id, source_group_key, version
		FROM questions WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions, err := r.scanQuestions(rows)
	if err != nil {
		return nil, err
	}

	// Reorder to match input IDs
	qMap := make(map[string]domain.Question)
	for _, q := range questions {
		qMap[string(q.ID)] = q
	}

	var ordered []domain.Question
	for _, id := range ids {
		if q, ok := qMap[string(id)]; ok {
			ordered = append(ordered, q)
		}
	}

	return ordered, nil
}

// FindPassageByID returns a passage by ID.
func (r *QuestionRepository) FindPassageByID(ctx context.Context, id examd.PassageID) (*domain.Passage, error) {
	var p domain.Passage
	var title sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, passage_number, title, content, level, section
		FROM passages WHERE id = $1
	`, id).Scan(&p.ID, &p.PassageNumber, &title, &p.Content, &p.Level, &p.Section)
	if err == sql.ErrNoRows {
		return nil, errors.New("NOT_FOUND", fmt.Sprintf("passage not found: %s", id))
	}
	if err != nil {
		return nil, err
	}

	p.Title = title.String
	return &p, nil
}

// FindAssetsForQuestions returns assets mapped by question ID.
func (r *QuestionRepository) FindAssetsForQuestions(ctx context.Context, ids []examd.QuestionID) (map[examd.QuestionID][]domain.Asset, error) {
	if len(ids) == 0 {
		return map[examd.QuestionID][]domain.Asset{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = string(id)
	}

	query := fmt.Sprintf(`
		SELECT id, type, source_url, s3_key, local_path, question_id, passage_id
		FROM assets WHERE question_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[examd.QuestionID][]domain.Asset)
	for rows.Next() {
		var a domain.Asset
		var qID, pID sql.NullString
		if err := rows.Scan(&a.ID, &a.Type, &a.SourceURL, &a.S3Key, &a.LocalPath, &qID, &pID); err != nil {
			return nil, err
		}
		if qID.Valid {
			qid := examd.QuestionID(qID.String)
			result[qid] = append(result[qid], a)
		}
	}

	return result, rows.Err()
}

// FindAssetsForPassages returns assets mapped by passage ID.
func (r *QuestionRepository) FindAssetsForPassages(ctx context.Context, ids []examd.PassageID) (map[examd.PassageID][]domain.Asset, error) {
	if len(ids) == 0 {
		return map[examd.PassageID][]domain.Asset{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = string(id)
	}

	query := fmt.Sprintf(`
		SELECT id, type, source_url, s3_key, local_path, question_id, passage_id
		FROM assets WHERE passage_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[examd.PassageID][]domain.Asset)
	for rows.Next() {
		var a domain.Asset
		var qID, pID sql.NullString
		if err := rows.Scan(&a.ID, &a.Type, &a.SourceURL, &a.S3Key, &a.LocalPath, &qID, &pID); err != nil {
			return nil, err
		}
		if pID.Valid {
			pid := examd.PassageID(pID.String)
			result[pid] = append(result[pid], a)
		}
	}

	return result, rows.Err()
}

// FindAssetByID returns a single asset by ID.
func (r *QuestionRepository) FindAssetByID(ctx context.Context, id string) (*domain.Asset, error) {
	var a domain.Asset
	var qID, pID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, type, source_url, s3_key, local_path, question_id, passage_id
		FROM assets WHERE id = $1
	`, id).Scan(&a.ID, &a.Type, &a.SourceURL, &a.S3Key, &a.LocalPath, &qID, &pID)

	if err == sql.ErrNoRows {
		return nil, errors.New("NOT_FOUND", fmt.Sprintf("asset not found: %s", id))
	}
	if err != nil {
		return nil, err
	}

	if qID.Valid {
		a.QuestionID = examd.QuestionID(qID.String)
	}
	if pID.Valid {
		a.PassageID = examd.PassageID(pID.String)
	}

	return &a, nil
}

// ListTemplates returns all package templates for a level.
func (r *QuestionRepository) ListTemplates(ctx context.Context, level examd.JLPTLevel) ([]domain.PackageTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, level, section_counts, total_questions, is_default
		FROM package_templates WHERE level = $1 ORDER BY is_default DESC
	`, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.PackageTemplate
	for rows.Next() {
		var t domain.PackageTemplate
		var sectionCounts []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.Level, &sectionCounts, &t.TotalQuestions, &t.IsDefault); err != nil {
			return nil, err
		}
		// Parse JSONB
		if len(sectionCounts) > 0 {
			t.SectionCounts = make(map[examd.Section]int)
			var raw map[string]interface{}
			if err := json.Unmarshal(sectionCounts, &raw); err == nil {
				for k, v := range raw {
					if n, ok := v.(float64); ok {
						t.SectionCounts[examd.Section(k)] = int(n)
					}
				}
			}
		}
		templates = append(templates, t)
	}

	return templates, rows.Err()
}

// scanQuestions scans question rows.
func (r *QuestionRepository) scanQuestions(rows *sql.Rows) ([]domain.Question, error) {
	var questions []domain.Question
	for rows.Next() {
		var q domain.Question
		var passageID, sourceGroupKey, answerNote, context sql.NullString
		if err := rows.Scan(&q.ID, &q.Level, &q.Section, &q.Prompt, &context, &q.AnswerValue,
			&answerNote, &passageID, &sourceGroupKey, &q.Version); err != nil {
			return nil, err
		}
		q.PassageID = examd.PassageID(passageID.String)
		q.SourceGroupKey = sourceGroupKey.String
		q.AnswerNote = answerNote.String
		q.Context = context.String
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
