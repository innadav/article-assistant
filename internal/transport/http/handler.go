package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"article-assistant/internal/article"
	"article-assistant/internal/executor"
	"article-assistant/internal/planner"
	"article-assistant/internal/processing"
	"article-assistant/internal/prompts"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler is the top-level component that orchestrates all HTTP requests.
type Handler struct {
	articleSvc       *article.Service
	plannerSvc       *planner.Service
	strategyExecutor *executor.Registry
	promptFactory    *prompts.Factory
	processingFacade *processing.Facade
}

// NewHandler creates a new handler with all its required dependencies.
func NewHandler(
	articleSvc *article.Service,
	plannerSvc *planner.Service,
	strategyExecutor *executor.Registry,
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

// Routes sets up all the API endpoints for the service.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer) // Add standard middleware

	r.Post("/chat", h.handleChat)
	r.Post("/articles", h.handleAddArticle)

	return r
}

// --- Request/Response Structs ---

type ChatRequest struct {
	Query string `json:"query"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

type AddArticleRequest struct {
	URL string `json:"url"`
}

// --- Handler Methods ---

// handleChat orchestrates the planner -> executor -> strategy flow.
func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Create a plan from the raw user query.
	plan, err := h.plannerSvc.CreatePlan(r.Context(), req.Query)
	if err != nil {
		http.Error(w, "Failed to create a query plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Pass the plan to the executor, which runs the correct strategy.
	answer, err := h.strategyExecutor.ExecutePlan(r.Context(), plan, h.articleSvc, h.promptFactory)
	if err != nil {
		http.Error(w, "Failed to execute the plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Return the final response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Answer: answer})
}

// handleAddArticle uses the Facade to process a new article.
func (h *Handler) handleAddArticle(w http.ResponseWriter, r *http.Request) {
	var req AddArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// The handler calls the Facade, hiding all the complex processing logic.
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
