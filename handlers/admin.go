package handlers

import (
	"net/http"

	"github.com/philiaspace/mondaiphi/internal/domain"
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

func (h *AdminHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) CreatePassage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) UpdatePassage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *AdminHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}
