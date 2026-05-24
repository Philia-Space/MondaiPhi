package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/philiaspace/mondaiphi/internal/domain"
	"github.com/philiaspace/phi-core/transport"
	dashboard "github.com/philiaspace/phi-dashboard"
	examd "github.com/philiaspace/phi-exam-domain/domain"
	storage "github.com/philiaspace/phi-storage/s3"
)

// DashboardHandler mounts a phi-dashboard admin panel for MondaiPhi.
type DashboardHandler struct {
	repo      domain.QuestionRepository
	s3Client  *storage.S3Client
	dash      *dashboard.Dashboard
	adminUser string
	adminPass string
	jwtSecret []byte
}

func NewDashboardHandler(repo domain.QuestionRepository, s3Client *storage.S3Client) *DashboardHandler {
	_, b, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(b)
	staticPath := filepath.Join(baseDir, "..", "web", "dist")

	dash := dashboard.New(dashboard.Config{
		Prefix:      "/dashboard",
		StaticPath:  staticPath,
		DevMode:     false,
		Title:       "MondaiPhi Admin",
		RequireAuth: true,
	})

	dash.RegisterModule(dashboard.Module{Name: "Questions", Path: "/dashboard/questions", Icon: "📝", Priority: 1})
	dash.RegisterModule(dashboard.Module{Name: "Passages", Path: "/dashboard/passages", Icon: "📖", Priority: 2})
	dash.RegisterModule(dashboard.Module{Name: "Assets", Path: "/dashboard/assets", Icon: "🎵", Priority: 3})
	dash.RegisterModule(dashboard.Module{Name: "Templates", Path: "/dashboard/templates", Icon: "📋", Priority: 4})

	// Load admin credentials from env (same as AuthPhi superadmin)
	adminUser := os.Getenv("PHILIA_ADMIN_USERNAME")
	adminPass := os.Getenv("PHILIA_ADMIN_PASSWORD")
	if adminUser == "" {
		adminUser = "admin"
	}
	if adminPass == "" {
		adminPass = "admin"
	}

	return &DashboardHandler{
		repo:      repo,
		s3Client:  s3Client,
		dash:      dash,
		adminUser: adminUser,
		adminPass: adminPass,
		jwtSecret: []byte("mondaiphi-dashboard-secret-" + adminPass),
	}
}

// RegisterRoutes mounts dashboard + CRUD API routes.
func (h *DashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	// Redirect /dashboard -> /dashboard/
	mux.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard/", http.StatusMovedPermanently)
	})

	// Public: login
	mux.HandleFunc("POST /dashboard/api/login", h.login)

	// Protected: stats
	mux.HandleFunc("GET /dashboard/api/stats", h.withAuth(h.getStats))

	// Protected: CRUD APIs
	mux.HandleFunc("/dashboard/api/questions", h.withAuth(h.handleQuestions))
	mux.HandleFunc("/dashboard/api/questions/{id}", h.withAuth(h.handleQuestionByID))
	mux.HandleFunc("/dashboard/api/questions/{id}/assets", h.withAuth(h.getQuestionAssets))
	mux.HandleFunc("POST /dashboard/api/questions/assets/batch", h.withAuth(h.batchQuestionAssets))
	mux.HandleFunc("/dashboard/api/passages", h.withAuth(h.handlePassages))
	mux.HandleFunc("/dashboard/api/passages/{id}", h.withAuth(h.handlePassageByID))
	mux.HandleFunc("/dashboard/api/assets", h.withAuth(h.handleAssets))
	mux.HandleFunc("/dashboard/api/assets/{id}", h.withAuth(h.handleAssetByID))
	mux.HandleFunc("/dashboard/api/templates", h.withAuth(h.handleTemplates))
	mux.HandleFunc("/dashboard/api/templates/{id}", h.withAuth(h.handleTemplateByID))

	// Mount the dashboard SPA (must be last - catch-all)
	h.dash.Mount(mux)
}

// ============================
// Auth
// ============================

func (h *DashboardHandler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		transport.BadRequest(w, "invalid request body")
		return
	}

	if req.Username != h.adminUser || req.Password != h.adminPass {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   map[string]string{"code": "UNAUTHORIZED", "message": "invalid credentials"},
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  req.Username,
		"role": "admin",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		transport.InternalError(w, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   86400,
		},
	})
}

func (h *DashboardHandler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   map[string]string{"code": "UNAUTHORIZED", "message": "missing authorization header"},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   map[string]string{"code": "UNAUTHORIZED", "message": "invalid authorization header format"},
			})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return h.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   map[string]string{"code": "UNAUTHORIZED", "message": "invalid or expired token"},
			})
			return
		}

		next(w, r)
	}
}

// ============================
// Stats
// ============================

func (h *DashboardHandler) getStats(w http.ResponseWriter, r *http.Request) {
	var totalQuestions int

	for _, level := range []examd.JLPTLevel{examd.N5, examd.N4, examd.N3, examd.N2, examd.N1} {
		for _, section := range []examd.Section{examd.Grammar, examd.Reading, examd.Listening} {
			qs, _ := h.repo.FindByLevelAndSection(r.Context(), level, section, 10000)
			totalQuestions += len(qs)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"total_questions": totalQuestions,
			"total_passages":  "381",  // From migration
			"total_assets":    "2089", // From migration
			"total_templates": "5",    // Approximate
		},
	})
}

// ============================
// Questions (method router)
// ============================

func (h *DashboardHandler) handleQuestions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listQuestions(w, r)
	case http.MethodPost:
		h.createQuestion(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DashboardHandler) handleQuestionByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getQuestion(w, r)
	case http.MethodPut:
		h.updateQuestion(w, r)
	case http.MethodDelete:
		h.deleteQuestion(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ============================
// Passages (method router)
// ============================

func (h *DashboardHandler) handlePassages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listPassages(w, r)
	case http.MethodPost:
		h.createPassage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DashboardHandler) handlePassageByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getPassage(w, r)
	case http.MethodPut:
		h.updatePassage(w, r)
	case http.MethodDelete:
		h.deletePassage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ============================
// Assets (method router)
// ============================

func (h *DashboardHandler) handleAssets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAssets(w, r)
	case http.MethodPost:
		h.createAsset(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DashboardHandler) handleAssetByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAsset(w, r)
	case http.MethodPut:
		h.updateAsset(w, r)
	case http.MethodDelete:
		h.deleteAsset(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ============================
// Templates (method router)
// ============================

func (h *DashboardHandler) handleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTemplates(w, r)
	case http.MethodPost:
		h.createTemplate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DashboardHandler) handleTemplateByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTemplate(w, r)
	case http.MethodPut:
		h.updateTemplate(w, r)
	case http.MethodDelete:
		h.deleteTemplate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ============================
// Questions CRUD
// ============================

func (h *DashboardHandler) listQuestions(w http.ResponseWriter, r *http.Request) {
	levelParam := r.URL.Query().Get("level")
	sectionParam := r.URL.Query().Get("section")
	searchParam := r.URL.Query().Get("search")
	sortParam := r.URL.Query().Get("sort")
	sortDirParam := r.URL.Query().Get("sort_dir")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit := 50
	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	questions, total, err := h.repo.SearchQuestions(r.Context(), examd.JLPTLevel(levelParam), examd.Section(sectionParam), searchParam, sortParam, sortDirParam, limit, offset)
	if err != nil {
		transport.InternalError(w, "failed to fetch questions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    questions,
		"count":   len(questions),
		"total":   total,
	})
}

func (h *DashboardHandler) getQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	q, opts, err := h.repo.FindWithOptions(r.Context(), examd.QuestionID(id))
	if err != nil {
		transport.FromError(w, err)
		return
	}

	assets, _ := h.repo.FindAssetsByQuestionID(r.Context(), id)
	for i := range assets {
		if assets[i].S3Key != "" && h.s3Client != nil {
			url, err := h.s3Client.PresignedURL(r.Context(), assets[i].S3Key, time.Hour)
			if err == nil {
				assets[i].SourceURL = url
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"question": q,
			"options":  opts,
			"assets":   assets,
		},
	})
}

func (h *DashboardHandler) createQuestion(w http.ResponseWriter, r *http.Request) {
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

	var options []domain.Option
	for _, o := range req.Options {
		options = append(options, domain.Option{
			QuestionID: question.ID,
			Value:      o.Value,
			Label:      o.Label,
			SortOrder:  o.SortOrder,
		})
	}
	if err := h.repo.SaveOptions(r.Context(), string(question.ID), options); err != nil {
		transport.InternalError(w, "failed to save options")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":      question.ID,
			"message": "question created",
		},
	})
}

func (h *DashboardHandler) updateQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":      question.ID,
			"message": "question updated",
		},
	})
}

func (h *DashboardHandler) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		transport.InternalError(w, "failed to delete question")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":      id,
			"message": "question deleted",
		},
	})
}

// ============================
// Question Assets
// ============================

func (h *DashboardHandler) getQuestionAssets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	assets, err := h.repo.FindAssetsByQuestionID(r.Context(), id)
	if err != nil {
		transport.InternalError(w, "failed to fetch assets")
		return
	}

	for i := range assets {
		if assets[i].S3Key != "" && h.s3Client != nil {
			url, err := h.s3Client.PresignedURL(r.Context(), assets[i].S3Key, time.Hour)
			if err == nil {
				assets[i].SourceURL = url
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    assets,
		"count":   len(assets),
	})
}

func (h *DashboardHandler) batchQuestionAssets(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuestionIDs []string `json:"question_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.QuestionIDs) == 0 {
		transport.BadRequest(w, "question_ids required")
		return
	}

	ids := make([]examd.QuestionID, len(req.QuestionIDs))
	for i, id := range req.QuestionIDs {
		ids[i] = examd.QuestionID(id)
	}

	assetsMap, err := h.repo.FindAssetsForQuestions(r.Context(), ids)
	if err != nil {
		transport.InternalError(w, "failed to fetch assets")
		return
	}

	result := make(map[string]interface{})
	for qid, assets := range assetsMap {
		for i := range assets {
			if assets[i].S3Key != "" && h.s3Client != nil {
				url, err := h.s3Client.PresignedURL(r.Context(), assets[i].S3Key, time.Hour)
				if err == nil {
					assets[i].SourceURL = url
				}
			}
		}
		result[string(qid)] = assets
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// ============================
// Passages CRUD
// ============================

func (h *DashboardHandler) listPassages(w http.ResponseWriter, r *http.Request) {
	levelParam := r.URL.Query().Get("level")
	sectionParam := r.URL.Query().Get("section")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit := 50
	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	passages, err := h.repo.ListPassages(r.Context(), examd.JLPTLevel(levelParam), examd.Section(sectionParam), limit+offset)
	if err != nil {
		transport.InternalError(w, "failed to fetch passages")
		return
	}
	total := len(passages)
	if offset > 0 && offset < len(passages) {
		passages = passages[offset : min(offset+limit, len(passages))]
	} else if limit < len(passages) {
		passages = passages[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    passages,
		"count":   len(passages),
		"total":   total,
	})
}

func (h *DashboardHandler) getPassage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.repo.FindPassageByID(r.Context(), examd.PassageID(id))
	if err != nil {
		transport.FromError(w, err)
		return
	}

	questions, _ := h.repo.FindByPassageID(r.Context(), examd.PassageID(id))

	passageAssets, _ := h.repo.FindAssetsForPassages(r.Context(), []examd.PassageID{examd.PassageID(id)})
	assets := passageAssets[examd.PassageID(id)]
	for i := range assets {
		if assets[i].S3Key != "" && h.s3Client != nil {
			url, err := h.s3Client.PresignedURL(r.Context(), assets[i].S3Key, time.Hour)
			if err == nil {
				assets[i].SourceURL = url
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"data":     p,
		"questions": questions,
		"assets":   assets,
	})
}

func (h *DashboardHandler) createPassage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "passage creation not yet implemented"},
	})
}

func (h *DashboardHandler) updatePassage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "passage update not yet implemented"},
	})
}

func (h *DashboardHandler) deletePassage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "passage deletion not yet implemented"},
	})
}

// ============================
// Assets CRUD
// ============================

func (h *DashboardHandler) listAssets(w http.ResponseWriter, r *http.Request) {
	typeParam := r.URL.Query().Get("type")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit := 50
	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	assets, total, err := h.repo.ListAssets(r.Context(), typeParam, limit, offset)
	if err != nil {
		transport.InternalError(w, "failed to fetch assets")
		return
	}

	for i := range assets {
		if assets[i].S3Key != "" && h.s3Client != nil {
			url, err := h.s3Client.PresignedURL(r.Context(), assets[i].S3Key, time.Hour)
			if err == nil {
				assets[i].SourceURL = url
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    assets,
		"count":   len(assets),
		"total":   total,
	})
}

func (h *DashboardHandler) getAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a, err := h.repo.FindAssetByID(r.Context(), id)
	if err != nil {
		transport.FromError(w, err)
		return
	}
	if a.S3Key != "" && h.s3Client != nil {
		url, err := h.s3Client.PresignedURL(r.Context(), a.S3Key, time.Hour)
		if err == nil {
			a.SourceURL = url
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": a})
}

func (h *DashboardHandler) createAsset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "asset creation not yet implemented"},
	})
}

func (h *DashboardHandler) updateAsset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "asset update not yet implemented"},
	})
}

func (h *DashboardHandler) deleteAsset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "asset deletion not yet implemented"},
	})
}

// ============================
// Templates CRUD
// ============================

func (h *DashboardHandler) listTemplates(w http.ResponseWriter, r *http.Request) {
	levelParam := r.URL.Query().Get("level")
	var templates []domain.PackageTemplate
	var err error

	if levelParam != "" {
		templates, err = h.repo.ListTemplates(r.Context(), examd.JLPTLevel(levelParam))
	} else {
		for _, level := range []examd.JLPTLevel{examd.N5, examd.N4, examd.N3, examd.N2, examd.N1} {
			qs, _ := h.repo.ListTemplates(r.Context(), level)
			templates = append(templates, qs...)
		}
	}
	if err != nil {
		transport.InternalError(w, "failed to fetch templates")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    templates,
		"count":   len(templates),
	})
}

func (h *DashboardHandler) getTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "template get not yet implemented"},
	})
}

func (h *DashboardHandler) createTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "template creation not yet implemented"},
	})
}

func (h *DashboardHandler) updateTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "template update not yet implemented"},
	})
}

func (h *DashboardHandler) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    map[string]string{"message": "template deletion not yet implemented"},
	})
}
