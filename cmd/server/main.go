package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"article-assistant/internal/article"
	"article-assistant/internal/config"
	"article-assistant/internal/executor"
	"article-assistant/internal/llm"
	"article-assistant/internal/planner"
	"article-assistant/internal/processing"
	"article-assistant/internal/prompts"
	"article-assistant/internal/repository"
	"article-assistant/internal/transport/http/handler"
)

func main() {
	ctx := context.Background()

	// 1. Load Configuration
	cfg := config.New()
	log.Printf("Configuration loaded. LLM Provider: %s, Prompt Version: %s", cfg.LLMProvider, cfg.PromptVersion)

	// 2. Initialize Core Components
	llmClient, err := llm.NewClientFactory(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	promptLoader, err := prompts.NewLoader(cfg.PromptVersion)
	if err != nil {
		log.Fatalf("Failed to load prompts: %v", err)
	}
	promptFactory, err := prompts.NewFactory(prompts.ModelGemini15Flash, promptLoader)
	if err != nil {
		log.Fatalf("Failed to create prompt factory: %v", err)
	}

	// Initialize the PostgreSQL repository
	repo, err := repository.NewPostgresRepo(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL repository: %v", err)
	}

	// 3. Initialize Services
	// These must be created in the correct order based on their dependencies.
	articleSvc := article.NewService(llmClient)
	plannerSvc := planner.NewService(llmClient, promptFactory, articleSvc)
	processingFacade := processing.NewFacade(llmClient, articleSvc, repo)
	strategyExecutor := executor.NewRegistry() // Corrected to just Executor

	// 4. Initialize the Transport Layer
	apiHandler := handler.NewHandler(
		articleSvc,
		plannerSvc,
		strategyExecutor,
		promptFactory,
		processingFacade,
	)

	// 4.5. Start Background Processes
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

	// 5. Start HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      apiHandler.Routes(), // Corrected to use the Routes method
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v", cfg.Port, err)
		}
	}()

	// 6. Graceful Shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down server...")

	downCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(downCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped")
}
