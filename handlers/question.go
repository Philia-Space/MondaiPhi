package handlers

import (
	"net/http"

	"github.com/philiaspace/mondaiphi/internal/domain"
)

// QuestionHandler handles public question routes.
type QuestionHandler struct {
	repo domain.QuestionRepository
}

func NewQuestionHandler(repo domain.QuestionRepository) *QuestionHandler {
	return &QuestionHandler{repo: repo}
}

func (h *QuestionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /questions", h.List)
	mux.HandleFunc("GET /questions/{id}", h.Get)
	mux.HandleFunc("GET /passages/{id}", h.GetPassage)
	mux.HandleFunc("GET /templates", h.ListTemplates)
	mux.HandleFunc("GET /assets/{id}", h.GetAsset)
}

func (h *QuestionHandler) List(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *QuestionHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *QuestionHandler) GetPassage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *QuestionHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *QuestionHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not implemented"}`))
}
