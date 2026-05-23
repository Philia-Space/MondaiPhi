package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/philiaspace/mondaiphi/internal/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
	"github.com/philiaspace/phi-core/transport"
)

// AdminHandler handles admin CRUD routes for questions, passages, and assets.
type AdminHandler struct {
	repo domain.QuestionRepository
}

func NewAdminHandler(repo domain.QuestionRepository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /admin/questions", h.CreateQuestion)
	mux.HandleFunc("PUT /admin/questions/{id}", h.UpdateQuestion)
	mux.HandleFunc("DELETE /admin/questions/{id}", h.DeleteQuestion)
	mux.HandleFunc("POST /admin/passages", h.CreatePassage)
	mux.HandleFunc("PUT /admin/passages/{id}", h.UpdatePassage)
	mux.HandleFunc("POST /admin/assets", h.UploadAsset)
	mux.HandleFunc("DELETE /admin/assets/{id}", h.DeleteAsset)
	mux.HandleFunc("POST /admin/templates", h.CreateTemplate)
	mux.HandleFunc("PUT /admin/templates/{id}", h.UpdateTemplate)
}

// CreateQuestion creates a new question with options.
func (h *AdminHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Level       string          `json:"level"`
		Section     string          `json:"section"`
		Prompt      string          `json:"prompt"`
		Context     string          `json:"context"`
		AnswerValue string          `json:"answer_value"`
		AnswerNote  string          `json:"answer_note"`
		PassageID   string          `json:"passage_id"`
		Options     []OptionRequest `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		transport.BadRequest(w, "invalid request body")
		return
	}

	if req.Prompt == "" || req.AnswerValue == "" || len(req.Options) == 0 {
		transport.BadRequest(w, "prompt, answer_value, and options are required")
		return
	}

	question := &domain.Question{
		ID:             examd.QuestionID("qst_new_" + generateShortID()),
		Level:          examd.JLPTLevel(req.Level),
		Section:        examd.Section(req.Section),
		Prompt:         req.Prompt,
		Context:        req.Context,
		AnswerValue:    req.AnswerValue,
		AnswerNote:     req.AnswerNote,
		PassageID:      examd.PassageID(req.PassageID),
		SourceGroupKey: "",
	}

	if err := h.repo.Save(r.Context(), question); err != nil {
		transport.InternalError(w, "failed to create question")
		return
	}

	transport.Created(w, map[string]interface{}{
		"id":      question.ID,
		"message": "question created",
	})
}

// UpdateQuestion updates an existing question.
func (h *AdminHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "question id is required")
		return
	}

	var req struct {
		Prompt      string `json:"prompt"`
		Context     string `json:"context"`
		AnswerValue string `json:"answer_value"`
		AnswerNote  string `json:"answer_note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		transport.BadRequest(w, "invalid request body")
		return
	}

	question, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		transport.FromError(w, err)
		return
	}

	if req.Prompt != "" {
		question.Prompt = req.Prompt
	}
	if req.Context != "" {
		question.Context = req.Context
	}
	if req.AnswerValue != "" {
		question.AnswerValue = req.AnswerValue
	}
	if req.AnswerNote != "" {
		question.AnswerNote = req.AnswerNote
	}

	if err := h.repo.Save(r.Context(), question); err != nil {
		transport.InternalError(w, "failed to update question")
		return
	}

	transport.OK(w, map[string]interface{}{
		"id":      question.ID,
		"message": "question updated",
	})
}

// DeleteQuestion soft-deletes a question.
func (h *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "question id is required")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		transport.InternalError(w, "failed to delete question")
		return
	}

	transport.OK(w, map[string]interface{}{
		"id":      id,
		"message": "question deleted",
	})
}

// CreatePassage creates a new reading/listening passage.
func (h *AdminHandler) CreatePassage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PassageNumber int    `json:"passage_number"`
		Title         string `json:"title"`
		Content       string `json:"content"`
		Level         string `json:"level"`
		Section       string `json:"section"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		transport.BadRequest(w, "invalid request body")
		return
	}

	if req.Content == "" || req.Level == "" {
		transport.BadRequest(w, "content and level are required")
		return
	}

	passage := &domain.Passage{
		ID:            examd.PassageID("psg_new_" + generateShortID()),
		PassageNumber: req.PassageNumber,
		Title:         req.Title,
		Content:       req.Content,
		Level:         examd.JLPTLevel(req.Level),
		Section:       examd.Section(req.Section),
	}

	// TODO: Implement passage repository save
	_ = passage

	transport.Created(w, map[string]interface{}{
		"id":      passage.ID,
		"message": "passage created",
	})
}

// UpdatePassage updates an existing passage.
func (h *AdminHandler) UpdatePassage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "passage id is required")
		return
	}

	// TODO: Implement passage update
	transport.OK(w, map[string]interface{}{
		"id":      id,
		"message": "passage updated (placeholder)",
	})
}

// UploadAsset handles file upload to S3.
func (h *AdminHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement multipart form parsing and S3 upload
	transport.Created(w, map[string]interface{}{
		"message": "asset upload placeholder",
	})
}

// DeleteAsset deletes an asset from S3 and metadata.
func (h *AdminHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "asset id is required")
		return
	}

	// TODO: Implement asset deletion
	transport.OK(w, map[string]interface{}{
		"id":      id,
		"message": "asset deleted (placeholder)",
	})
}

// CreateTemplate creates a new package template.
func (h *AdminHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string         `json:"name"`
		Level          string         `json:"level"`
		SectionCounts  map[string]int `json:"section_counts"`
		TotalQuestions int            `json:"total_questions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		transport.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" || req.TotalQuestions == 0 {
		transport.BadRequest(w, "name and total_questions are required")
		return
	}

	// Validate section counts sum to total
	total := 0
	for _, count := range req.SectionCounts {
		total += count
	}
	if total != req.TotalQuestions {
		transport.BadRequest(w, "section counts must sum to total_questions")
		return
	}

	// TODO: Implement template repository save
	transport.Created(w, map[string]interface{}{
		"message": "template created (placeholder)",
	})
}

// UpdateTemplate updates an existing template.
func (h *AdminHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		transport.BadRequest(w, "template id is required")
		return
	}

	// TODO: Implement template update
	transport.OK(w, map[string]interface{}{
		"id":      id,
		"message": "template updated (placeholder)",
	})
}

// Helper types
type OptionRequest struct {
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

// generateShortID creates a short random identifier
func generateShortID() string {
	// Simple implementation - in production use proper ULID generation
	return "temp"
}
