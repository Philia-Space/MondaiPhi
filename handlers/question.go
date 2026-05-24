package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/philiaspace/mondaiphi/internal/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
	"github.com/philiaspace/phi-core/transport"
	storage "github.com/philiaspace/phi-storage/s3"
)

// QuestionHandler handles public question routes.
type QuestionHandler struct {
	repo   domain.QuestionRepository
	storage *storage.S3Client
}

func NewQuestionHandler(repo domain.QuestionRepository, s3Client *storage.S3Client) *QuestionHandler {
	return &QuestionHandler{repo: repo, storage: s3Client}
}

func (h *QuestionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /questions", h.List)
	mux.HandleFunc("GET /questions/{id}", h.Get)
	mux.HandleFunc("GET /passages/{id}", h.GetPassage)
	mux.HandleFunc("GET /templates", h.ListTemplates)
	mux.HandleFunc("GET /assets/{id}", h.GetAsset)
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
	var sanitized []QuestionResponse
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

	// Sanitize options: remove is_correct
	var sanitizedOptions []OptionResponse
	for _, opt := range options {
		sanitizedOptions = append(sanitizedOptions, OptionResponse{
			ID:        opt.ID,
			Value:     opt.Value,
			Label:     opt.Label,
			SortOrder: opt.SortOrder,
		})
	}

	transport.OK(w, map[string]interface{}{
		"question": sanitizeQuestion(*question),
		"options":  sanitizedOptions,
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

// GetAsset redirects to presigned S3 URL or returns direct URL.
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
}

type OptionResponse struct {
	ID        string `json:"id"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
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
	}
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
