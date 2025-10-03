package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"article-assistant/internal/domain"
	"article-assistant/internal/repository"
)

// Service handles chat request/response caching
type Service struct {
	Repo *repository.Repo
}

// NewService creates a new cache service
func NewService(repo *repository.Repo) *Service {
	return &Service{Repo: repo}
}

// calculateRequestHash computes SHA-256 hash of the request for caching
func calculateRequestHash(request interface{}) (string, error) {
	// Marshal request to JSON for consistent hashing
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	hash := sha256.Sum256(requestJSON)
	return fmt.Sprintf("%x", hash), nil
}

// GetCachedResponse retrieves cached response for a request
func (s *Service) GetCachedResponse(ctx context.Context, request interface{}) (*domain.ChatResponse, error) {
	requestHash, err := calculateRequestHash(request)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate request hash: %w", err)
	}

	cache, err := s.Repo.GetChatCache(ctx, requestHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	if cache == nil {
		log.Printf("💾 Cache miss for request hash: %s", requestHash[:8])
		return nil, nil // Cache miss
	}

	log.Printf("💾 Cache hit for request hash: %s", requestHash[:8])

	// Convert cached response back to ChatResponse
	var response domain.ChatResponse
	responseJSON, err := json.Marshal(cache.ResponseJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cached response: %w", err)
	}

	err = json.Unmarshal(responseJSON, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &response, nil
}

// SetCachedResponse stores a request/response pair in cache
func (s *Service) SetCachedResponse(ctx context.Context, request interface{}, response *domain.ChatResponse) error {
	requestHash, err := calculateRequestHash(request)
	if err != nil {
		return fmt.Errorf("failed to calculate request hash: %w", err)
	}

	err = s.Repo.SetChatCache(ctx, requestHash, request, response)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	log.Printf("💾 Cached response for request hash: %s", requestHash[:8])
	return nil
}

// CleanExpiredCache removes expired cache entries
func (s *Service) CleanExpiredCache(ctx context.Context) error {
	err := s.Repo.CleanExpiredChatCache(ctx)
	if err != nil {
		return fmt.Errorf("failed to clean expired cache: %w", err)
	}

	log.Println("🧹 Cleaned expired cache entries")
	return nil
}

// StartCacheCleanup starts a background goroutine to clean expired cache entries
func (s *Service) StartCacheCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("🛑 Cache cleanup stopped")
				return
			case <-ticker.C:
				if err := s.CleanExpiredCache(ctx); err != nil {
					log.Printf("❌ Failed to clean expired cache: %v", err)
				}
			}
		}
	}()

	log.Printf("🔄 Started cache cleanup with interval: %v", interval)
}
