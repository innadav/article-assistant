package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"article-chat-system/internal/article"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/strategies"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	articleSvc       *article.Service
	plannerSvc       *planner.Service
	strategyExecutor *strategies.Executor
	promptFactory    *prompts.Factory
	processingFacade *processing.Facade
}

func NewHandler(
	articleSvc *article.Service,
	plannerSvc *planner.Service,
	strategyExecutor *strategies.Executor,
	promptFactory *prompts.Factory,
	processingFacade *processing.Facade,
) *Handler {
	return &Handler{
		articleSvc:       articleSvc,
		plannerSvc:       plannerSvc,
		strategyExecutor: strategyExecutor,
		promptFactory:    promptFactory,
		processingFacade: processingFacade,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Post("/chat", h.handleChat)
	r.Post("/articles", h.handleAddArticle)
	return r
}

type ChatRequest struct {
	Query string `json:"query"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

type AddArticleRequest struct {
	URL string `json:"url"`
}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	availableArticles := h.articleSvc.GetAllArticles()
	var plannerArticles []*prompts.Article
	for _, art := range availableArticles {
		plannerArticles = append(plannerArticles, &prompts.Article{
			URL:   art.URL,
			Title: art.Title,
		})
	}

	plan, err := h.plannerSvc.CreatePlan(r.Context(), req.Query, plannerArticles)
	if err != nil {
		http.Error(w, "Failed to create a query plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	answer, err := h.strategyExecutor.ExecutePlan(r.Context(), plan, h.articleSvc, h.promptFactory)
	if err != nil {
		http.Error(w, "Failed to execute the plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Answer: answer})
}

func (h *Handler) handleAddArticle(w http.ResponseWriter, r *http.Request) {
	var req AddArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newArticle, err := h.processingFacade.AddNewArticle(r.Context(), req.URL)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}
		http.Error(w, "Failed to process article: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(newArticle)
}
