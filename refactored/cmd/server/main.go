package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"article-chat-system/internal/article"
	"article-chat-system/internal/config"
	"article-chat-system/internal/llm"
	"article-chat-system/internal/planner"
	"article-chat-system/internal/processing"
	"article-chat-system/internal/prompts"
	"article-chat-system/internal/strategies"
	"article-chat-system/internal/transport/http/handler"
)

func main() {
	ctx := context.Background()

	// --- 1. Load Configuration First ---
	// This is the foundation and has no dependencies.
	cfg := config.New()
	log.Printf("Configuration loaded. LLM Provider: %s, Prompt Version: %s", cfg.LLMProvider, cfg.PromptVersion)

	// --- 2. Initialize Core, Independent Components ---
	// These components have few or no dependencies on other parts of our application.
	llmClient, err := llm.NewClientFactory(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	promptLoader, err := prompts.NewLoader(cfg.PromptVersion)
	if err != nil {
		log.Fatalf("Failed to load prompts: %v", err)
	}

	promptFactory := prompts.NewFactory(prompts.ModelGemini15Flash, promptLoader)
	strategyExecutor := strategies.NewExecutor()

	// --- 3. Initialize Services that Depend on Core Components ---
	// These services depend on the components created in step 2.
	articleSvc := article.NewService(llmClient)
	plannerSvc := planner.NewService(llmClient, promptFactory)
	processingFacade := processing.NewFacade(llmClient, articleSvc)

	// --- 4. Initialize the Transport Layer Last ---
	// The handler is the top layer; it depends on all the services, so it must be created last.
	apiHandler := handler.NewHandler(
		articleSvc,
		plannerSvc,
		strategyExecutor,
		promptFactory,
		processingFacade,
	)

	// --- 5. Start Background Processes and the Server ---
	// Now that all components are correctly initialized, we can start the application.
	go func() {
		log.Println("Processing initial articles in the background...")
		for _, url := range cfg.InitialArticleURLs {
			_, err := processingFacade.AddNewArticle(context.Background(), url)
			if err != nil {
				log.Printf("Failed to process initial URL %s: %v", url, err)
			}
		}
		log.Println("Initial article processing complete.")
	}()

	server := &http.Server{
		Addr:    ":8080",
		Handler: apiHandler.Routes(),
	}

	go func() {
		log.Println("Starting server on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	// --- 6. Handle Graceful Shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped.")
}
