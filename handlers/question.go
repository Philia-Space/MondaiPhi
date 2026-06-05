package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/philiaspace/mondaiphi/internal/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
	"github.com/philiaspace/phi-core/transport"
	storage "github.com/philiaspace/phi-storage/s3"
)

// QuestionHandler handles public question routes.
type QuestionHandler struct {
	repo          domain.QuestionRepository
	storage       *storage.S3Client
	serviceSecret string
}

func NewQuestionHandler(repo domain.QuestionRepository, s3Client *storage.S3Client, serviceSecret string) *QuestionHandler {
	return &QuestionHandler{repo: repo, storage: s3Client, serviceSecret: serviceSecret}
}

func (h *QuestionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /questions", h.List)
	mux.HandleFunc("GET /questions/{id}", h.Get)
	mux.HandleFunc("GET /internal/questions/{id}", h.GetInternal)
	mux.HandleFunc("GET /passages/{id}", h.GetPassage)
	mux.HandleFunc("GET /templates", h.ListTemplates)
	mux.HandleFunc("GET /assets/{id}", h.GetAsset)
	// Archive / chronological endpoints
	mux.HandleFunc("GET /exams", h.ListExams)
	mux.HandleFunc("GET /exams/{id}/questions", h.ListExamQuestions)
}

// List returns a list of questions filtered by level and section.
func (h *QuestionHandler) List(w http.ResponseWriter, r *http.Request) {
	levelParam := r.URL.Query().Get("level")
	sectionParam := r.URL.Query().Get("section")
	limitParam := r.URL.Query().Get("limit")

	if levelParam == "" {
		transport.BadRequest(w, "level parameter is required")
		return
	}

	level := examd.JLPTLevel(levelParam)
	if !isValidLevel(level) {
		transport.BadRequest(w, "invalid level: must be N1, N2, N3, N4, or N5")
		return
	}

	limit := 50
	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 300 {
			limit = l
		}
	}

	var questions []domain.Question
	var err error

	if sectionParam != "" {
		section := examd.Section(sectionParam)
		if !isValidSection(section) {
			transport.BadRequest(w, "invalid section: must be grammar, reading, or listening")
			return
		}
		questions, err = h.repo.FindByLevelAndSection(r.Context(), level, section, limit)
	} else {
		// Get all sections
		for _, section := range []examd.Section{examd.Grammar, examd.Reading, examd.Listening} {
			qs, err := h.repo.FindByLevelAndSection(r.Context(), level, section, limit/3+10)
			if err != nil {
				transport.InternalError(w, "failed to fetch questions")
				return
			}
			questions = append(questions, qs...)
		}
	}

	if err != nil {
		transport.InternalError(w, "failed to fetch questions")
		return
	}

	// Sanitize: remove answer data
	sanitized := make([]QuestionResponse, 0, len(questions))
	for _, q := range questions {
		sanitized = append(sanitized, sanitizeQuestion(q))
	}

	transport.OK(w, map[string]interface{}{
		"questions": sanitized,
		"count":     len(sanitized),
	})
}

// Get returns a single question by ID (sanitized).
func (h *QuestionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "question id is required")
		return
	}

	question, options, err := h.repo.FindWithOptions(r.Context(), examd.QuestionID(id))
	if err != nil {
		status := transport.FromError(w, err)
		if status == 0 {
			transport.InternalError(w, "failed to fetch question")
		}
		return
	}

	var sanitizedOptions []OptionResponse
	for _, opt := range options {
		sanitizedOptions = append(sanitizedOptions, OptionResponse{
			ID:        opt.ID,
			Value:     opt.Value,
			Label:     opt.Label,
			SortOrder: opt.SortOrder,
		})
	}

	assets, _ := h.repo.FindAssetsByQuestionID(r.Context(), string(question.ID))
	var assetResponses []AssetResponse
	for _, a := range assets {
		assetResponses = append(assetResponses, AssetResponse{
			ID:   string(a.ID),
			Type: a.Type,
		})
	}

	transport.OK(w, map[string]interface{}{
		"question": sanitizeQuestion(*question),
		"options":  sanitizedOptions,
		"assets":   assetResponses,
	})
}

// GetPassage returns a passage by ID with its questions (sanitized).
func (h *QuestionHandler) GetPassage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "passage id is required")
		return
	}

	passage, err := h.repo.FindPassageByID(r.Context(), examd.PassageID(id))
	if err != nil {
		status := transport.FromError(w, err)
		if status == 0 {
			transport.InternalError(w, "failed to fetch passage")
		}
		return
	}

	questions, err := h.repo.FindByPassageID(r.Context(), examd.PassageID(id))
	if err != nil {
		transport.InternalError(w, "failed to fetch passage questions")
		return
	}

	var sanitizedQuestions []QuestionResponse
	for _, q := range questions {
		sanitizedQuestions = append(sanitizedQuestions, sanitizeQuestion(q))
	}

	transport.OK(w, map[string]interface{}{
		"passage":   passage,
		"questions": sanitizedQuestions,
	})
}

// ListTemplates returns all package templates.
func (h *QuestionHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	levelParam := r.URL.Query().Get("level")
	
	var templates []domain.PackageTemplate
	var err error
	
	if levelParam != "" {
		level := examd.JLPTLevel(levelParam)
		if !isValidLevel(level) {
			transport.BadRequest(w, "invalid level")
			return
		}
		templates, err = h.repo.ListTemplates(r.Context(), level)
	} else {
		// List templates for all levels
		for _, level := range []examd.JLPTLevel{examd.N5, examd.N4, examd.N3, examd.N2, examd.N1} {
			ts, err := h.repo.ListTemplates(r.Context(), level)
			if err != nil {
				transport.InternalError(w, "failed to fetch templates")
				return
			}
			templates = append(templates, ts...)
		}
	}

	if err != nil {
		transport.InternalError(w, "failed to fetch templates")
		return
	}

	transport.OK(w, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetAsset redirects to presigned S3 URL or serves locally.
func (h *QuestionHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "asset id is required")
		return
	}

	// Fetch asset metadata from database
	asset, err := h.repo.FindAssetByID(r.Context(), id)
	if err != nil {
		transport.FromError(w, err)
		return
	}

	// Try local file if available
	if asset.LocalPath != "" {
		localFile := h.findLocalAsset(asset.LocalPath)
		if localFile != "" {
			http.ServeFile(w, r, localFile)
			return
		}
	}

	if h.storage != nil && asset.S3Key != "" {
		// Generate presigned URL for private assets
		presignedURL, err := h.storage.PresignedURL(r.Context(), asset.S3Key, 15*time.Minute)
		if err == nil {
			http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
			return
		}
		// Fallback: return direct public URL
		http.Redirect(w, r, h.storage.ObjectURL(asset.S3Key), http.StatusTemporaryRedirect)
		return
	}

	// No S3 configured or asset not uploaded: return metadata
	transport.OK(w, map[string]interface{}{
		"asset_id":   asset.ID,
		"type":       asset.Type,
		"source_url": asset.SourceURL,
		"s3_key":     asset.S3Key,
	})
}

// Helper types and functions

type QuestionResponse struct {
	ID             string `json:"id"`
	Level          string `json:"level"`
	Section        string `json:"section"`
	Prompt         string `json:"prompt"`
	Context        string `json:"context,omitempty"`
	PassageID      string `json:"passage_id,omitempty"`
	SourceGroupKey string `json:"source_group_key,omitempty"`
	// Archive fields
	Year         int    `json:"year,omitempty"`
	Month        int    `json:"month,omitempty"`
	DateLabel    string `json:"date_label,omitempty"`
	QuestionType int    `json:"question_type,omitempty"`
	SectionOrder int    `json:"section_order,omitempty"`
	SectionTitle string `json:"section_title,omitempty"`
	IsPractice   bool   `json:"is_practice,omitempty"`
}

type InternalQuestionResponse struct {
	ID             string `json:"id"`
	Level          string `json:"level"`
	Section        string `json:"section"`
	Prompt         string `json:"prompt"`
	Context        string `json:"context,omitempty"`
	AnswerValue    string `json:"answer_value"`
	PassageID      string `json:"passage_id,omitempty"`
	SourceGroupKey string `json:"source_group_key,omitempty"`
}

type OptionResponse struct {
	ID        string `json:"id"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

type AssetResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func sanitizeQuestion(q domain.Question) QuestionResponse {
	return QuestionResponse{
		ID:             string(q.ID),
		Level:          string(q.Level),
		Section:        string(q.Section),
		Prompt:         q.Prompt,
		Context:        q.Context,
		PassageID:      string(q.PassageID),
		SourceGroupKey: q.SourceGroupKey,
		Year:           q.Year,
		Month:          q.Month,
		DateLabel:      q.DateLabel,
		QuestionType:   q.QuestionType,
		SectionOrder:   q.SectionOrder,
		SectionTitle:   q.SectionTitle,
		IsPractice:     q.IsPractice,
	}
}

func (h *QuestionHandler) GetInternal(w http.ResponseWriter, r *http.Request) {
	// Service-to-service auth: require X-Service-Secret header
	secret := r.Header.Get("X-Service-Secret")
	if secret == "" || secret != h.serviceSecret {
		transport.Forbidden(w, "service auth required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "question id is required")
		return
	}

	question, options, err := h.repo.FindWithOptions(r.Context(), examd.QuestionID(id))
	if err != nil {
		status := transport.FromError(w, err)
		if status == 0 {
			transport.InternalError(w, "failed to fetch question")
		}
		return
	}

	var optResponses []OptionResponse
	for _, opt := range options {
		optResponses = append(optResponses, OptionResponse{
			ID:        opt.ID,
			Value:     opt.Value,
			Label:     opt.Label,
			SortOrder: opt.SortOrder,
		})
	}

	transport.OK(w, map[string]interface{}{
		"question": InternalQuestionResponse{
			ID:             string(question.ID),
			Level:          string(question.Level),
			Section:        string(question.Section),
			Prompt:         question.Prompt,
			Context:        question.Context,
			AnswerValue:    question.AnswerValue,
			PassageID:      string(question.PassageID),
			SourceGroupKey: question.SourceGroupKey,
		},
		"options": optResponses,
	})
}

func isValidLevel(level examd.JLPTLevel) bool {
	for _, l := range examd.AllLevels() {
		if l == level {
			return true
		}
	}
	return false
}

func isValidSection(section examd.Section) bool {
	for _, s := range examd.AllSections() {
		if s == section {
			return true
		}
	}
	return false
}

// ============================================================
// ARCHIVE ENDPOINTS — chronological exam browsing
// ============================================================

// ListExams returns available exams, optionally filtered by level.
func (h *QuestionHandler) ListExams(w http.ResponseWriter, r *http.Request) {
	levelParam := r.URL.Query().Get("level")
	var level examd.JLPTLevel
	if levelParam != "" {
		level = examd.JLPTLevel(levelParam)
		if !isValidLevel(level) {
			transport.BadRequest(w, "invalid level")
			return
		}
	}

	exams, err := h.repo.ListExams(r.Context(), level, 50)
	if err != nil {
		transport.InternalError(w, "failed to fetch exams")
		return
	}

	transport.OK(w, map[string]interface{}{
		"exams": exams,
		"count": len(exams),
	})
}

// ListExamQuestions returns all questions for a specific exam in chronological order.
func (h *QuestionHandler) ListExamQuestions(w http.ResponseWriter, r *http.Request) {
	examID := r.PathValue("id")
	if examID == "" {
		transport.BadRequest(w, "exam id is required")
		return
	}

	questions, err := h.repo.FindQuestionsByExam(r.Context(), examID)
	if err != nil {
		transport.InternalError(w, "failed to fetch exam questions")
		return
	}

	// Sanitize: remove answer data
	sanitized := make([]QuestionResponse, 0, len(questions))
	for _, q := range questions {
		sanitized = append(sanitized, sanitizeQuestion(q))
	}

	transport.OK(w, map[string]interface{}{
		"exam_id":   examID,
		"questions": sanitized,
		"count":     len(sanitized),
	})
}

// findLocalAsset searches for a media file on the local filesystem using the base filename.
func (h *QuestionHandler) findLocalAsset(localPath string) string {
	baseName := filepath.Base(localPath)

	// Search paths relative to common archive media roots
	searchRoots := []string{
		"/home/nexsal/Project/philiaspace/De Thi Tieng Nhat Archive",
	}

	// Also check if the localPath itself is absolute and exists
	if filepath.IsAbs(localPath) {
		if _, err := os.Stat(localPath); err == nil {
			return localPath
		}
	}

	for _, root := range searchRoots {
		var found string
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && d.Name() == baseName {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
		if err == nil && found != "" {
			return found
		}
	}

	// Check if the path contains level name (N1-N5) and try to extract it
	// e.g. "dethitiengnhat_N1/listening/listening_audio/file.mp3"
	// or "dethitiengnhat_N1/listening/listening_images/file.png"
	// or "reading_images/file.jpg"
	for _, level := range []string{"N1", "N2", "N3", "N4", "N5"} {
		if strings.Contains(localPath, level) {
			// Try reading images
			dirPath := filepath.Join(searchRoots[0], "dethitiengnhat_"+level, "reading", "reading_images", baseName)
			if _, err := os.Stat(dirPath); err == nil {
				return dirPath
			}
			// Try audio directory
			dirPath = filepath.Join(searchRoots[0], "dethitiengnhat_"+level, "listening", "listening_audio", baseName)
			if _, err := os.Stat(dirPath); err == nil {
				return dirPath
			}
			// Try images directory
			dirPath = filepath.Join(searchRoots[0], "dethitiengnhat_"+level, "listening", "listening_images", baseName)
			if _, err := os.Stat(dirPath); err == nil {
				return dirPath
			}
		}
	}

	return ""
}
