package executor

import (
	"article-assistant/internal/domain"
	"article-assistant/internal/llm"
	"article-assistant/internal/repository"
	"context"
	"fmt"
	"strings"
)

// Summary Command
type SummaryCommand struct {
	Repo *repository.Repo
}

func (c *SummaryCommand) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract URL from args
	var targetURL string
	if urlsVal, ok := plan.Args["urls"]; ok {
		if urlSlice, ok := urlsVal.([]interface{}); ok && len(urlSlice) > 0 {
			if urlStr, ok := urlSlice[0].(string); ok {
				targetURL = urlStr
			}
		}
	} else {
		return &domain.ChatResponse{
			Answer: "Article URL required for summary",
			Task:   plan.Command,
		}, nil
	}

	// Get article by URL
	articles, err := c.Repo.GetArticlesByURLs(ctx, []string{targetURL})
	if err != nil {
		return &domain.ChatResponse{
			Answer: "Error retrieving article: " + targetURL,
			Task:   plan.Command,
		}, nil
	}

	if len(articles) == 0 {
		return &domain.ChatResponse{
			Answer: "Article not found: " + targetURL,
			Task:   plan.Command,
		}, nil
	}

	return &domain.ChatResponse{
		Answer:       articles[0].Summary,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Helper functions
func extractURLs(plan *domain.Plan) []string {
	var targetURLs []string
	if urlsVal, ok := plan.Args["urls"]; ok {
		if urlSlice, ok := urlsVal.([]interface{}); ok {
			for _, u := range urlSlice {
				if urlStr, ok := u.(string); ok {
					targetURLs = append(targetURLs, urlStr)
				}
			}
		}
	}
	return targetURLs
}

func errorResponse(command, message string) *domain.ChatResponse {
	return &domain.ChatResponse{
		Answer:       message,
		ResponseType: domain.ResponseText,
		Task:         command,
	}
}

// KeywordsOrTopics Command
type FetchKeywordsOrTopicsCommand struct {
	Repo *repository.Repo
}

func (c *FetchKeywordsOrTopicsCommand) Execute(ctx context.Context, plan *domain.Plan, _ string) (*domain.ChatResponse, error) {
	targetURLs := extractURLs(plan)
	if len(targetURLs) == 0 {
		return errorResponse(plan.Command, "URLs required to extract keywords/topics"), nil
	}

	keywords, topics, err := c.Repo.GetKeywordsAndTopics(ctx, targetURLs, 5)
	if err != nil {
		return nil, err
	}

	if len(keywords) == 0 && len(topics) == 0 {
		return errorResponse(plan.Command, "No keywords/topics found"), nil
	}

	var result strings.Builder
	if len(keywords) > 0 {
		result.WriteString("Top Keywords:\n")
		for i, k := range keywords {
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, k.Term))
		}
	}
	if len(topics) > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("Top Topics:\n")
		for i, t := range topics {
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, t.Name))
		}
	}

	return &domain.ChatResponse{
		Answer:       result.String(),
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Sentiment Command
type FetchSentimentCommand struct {
	Repo *repository.Repo
}

func (c *FetchSentimentCommand) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract URLs from args
	targetURLs := extractURLs(plan)
	if len(targetURLs) == 0 {
		return errorResponse(plan.Command, "URLs required for sentiment analysis"), nil
	}

	// Fetch articles by URLs to get sentiment data
	arts, err := c.Repo.GetArticlesByURLs(ctx, targetURLs)
	if err != nil {
		return nil, err
	}

	if len(arts) == 0 {
		return errorResponse(plan.Command, "No articles found for the provided URLs"), nil
	}

	var sentiments []string
	var totalScore float64
	for _, a := range arts {
		sentiments = append(sentiments, fmt.Sprintf("%s: %s (%.2f)", a.URL, a.Sentiment, a.SentimentScore))
		totalScore += a.SentimentScore
	}

	avgScore := totalScore / float64(len(arts))
	var overallSentiment string
	if avgScore > 0.6 {
		overallSentiment = "positive"
	} else if avgScore < 0.4 {
		overallSentiment = "negative"
	} else {
		overallSentiment = "neutral"
	}

	result := fmt.Sprintf("Overall sentiment: %s (%.2f)\nArticles:\n%s",
		overallSentiment, avgScore, strings.Join(sentiments, "\n"))

	return &domain.ChatResponse{
		Answer:       result,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Compare Command
type CompareCommand struct {
	Repo *repository.Repo
	LLM  *llm.OpenAIClient
}

func (c *CompareCommand) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract URLs from args
	var targetURLs []string
	if urlsVal, ok := plan.Args["urls"]; ok {
		if urlSlice, ok := urlsVal.([]interface{}); ok {
			for _, u := range urlSlice {
				if urlStr, ok := u.(string); ok {
					targetURLs = append(targetURLs, urlStr)
				}
			}
		}
	}

	if len(targetURLs) < 2 {
		return &domain.ChatResponse{
			Answer: "At least 2 URLs required for comparison",
			Task:   plan.Command,
		}, nil
	}

	// Get articles for comparison
	articles, err := c.Repo.GetArticlesByURLs(ctx, targetURLs)
	if err != nil {
		return &domain.ChatResponse{
			Answer: "Error retrieving articles for comparison",
			Task:   plan.Command,
		}, nil
	}

	if len(articles) < 2 {
		return &domain.ChatResponse{
			Answer: "Could not find at least 2 articles for comparison",
			Task:   plan.Command,
		}, nil
	}

	var summaries []string
	for _, article := range articles {
		summaries = append(summaries, article.Summary)
	}

	// Use LLM to compare summaries
	comparison, err := c.LLM.GenerateText(ctx, fmt.Sprintf("Compare these articles:\n1. %s\n2. %s", summaries[0], summaries[1]))
	if err != nil {
		return &domain.ChatResponse{
			Answer: "Error generating comparison",
			Task:   plan.Command,
		}, nil
	}

	return &domain.ChatResponse{
		Answer:       comparison,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Tone Command
type ToneKeyDfferencesCommand struct {
	Repo *repository.Repo
	LLM  *llm.OpenAIClient
}

func (c *ToneKeyDfferencesCommand) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract URLs from args
	var targetURLs []string
	if urlsVal, ok := plan.Args["urls"]; ok {
		if urlSlice, ok := urlsVal.([]interface{}); ok {
			for _, u := range urlSlice {
				if urlStr, ok := u.(string); ok {
					targetURLs = append(targetURLs, urlStr)
				}
			}
		}
	}

	if len(targetURLs) < 2 {
		return &domain.ChatResponse{
			Answer: "At least 2 URLs required for tone comparison",
			Task:   plan.Command,
		}, nil
	}

	// Get articles for tone comparison
	articles, err := c.Repo.GetArticlesByURLs(ctx, targetURLs)
	if err != nil {
		return &domain.ChatResponse{
			Answer: "Error retrieving articles for tone comparison",
			Task:   plan.Command,
		}, nil
	}

	if len(articles) < 2 {
		return &domain.ChatResponse{
			Answer: "Could not find at least 2 articles for tone comparison",
			Task:   plan.Command,
		}, nil
	}

	var summaries []string
	for _, article := range articles {
		summaries = append(summaries, article.Summary)
	}

	// Use LLM to compare tone
	toneDiff, err := c.LLM.ToneCompare(ctx, summaries[0], summaries[1])
	if err != nil {
		return &domain.ChatResponse{
			Answer: "Error comparing tone",
			Task:   plan.Command,
		}, nil
	}

	return &domain.ChatResponse{
		Answer:       toneDiff,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// MorePositive Command
type FetchMostPositivesByFilter struct {
	Repo *repository.Repo
	LLM  *llm.OpenAIClient
}

func (c *FetchMostPositivesByFilter) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract filter from args
	var filter string
	if filterVal, ok := plan.Args["filter"]; ok {
		if filterStr, ok := filterVal.(string); ok {
			filter = filterStr
		}
	}

	if filter == "" {
		return errorResponse(plan.Command, "Filter required for finding most positive article"), nil
	}

	// Step 1: Embed the filter and find similar articles
	embedding, err := c.LLM.Embed(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	candidates, err := c.Repo.GetArticlesByVectorSearch(ctx, embedding, 10, []string{})
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return errorResponse(plan.Command, "No articles found for the given filter"), nil
	}

	// Step 2: Find the article with the highest sentiment score
	var best *domain.Article
	bestScore := -1.0
	for _, a := range candidates {
		if a.SentimentScore > bestScore {
			bestScore = a.SentimentScore
			best = &a
		}
	}

	if best == nil {
		return errorResponse(plan.Command, "No articles with sentiment data found"), nil
	}

	result := fmt.Sprintf("Most positive article about %s:\n%s\nTitle: %s\nSentiment: %s (%.2f)",
		filter, best.URL, best.Title, best.Sentiment, best.SentimentScore)

	return &domain.ChatResponse{
		Answer:       result,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// TopEntities Command
type FetchTopEntitiesFromDBCommand struct {
	Repo *repository.Repo
}

func (c *FetchTopEntitiesFromDBCommand) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract URLs from args for get_top_db_entities
	var targetURLs []string
	if urlsVal, ok := plan.Args["urls"]; ok {
		if urlSlice, ok := urlsVal.([]interface{}); ok {
			for _, u := range urlSlice {
				if urlStr, ok := u.(string); ok {
					targetURLs = append(targetURLs, urlStr)
				}
			}
		}
	}

	entities, err := c.Repo.GetTopEntities(ctx, 10, targetURLs)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return &domain.ChatResponse{
			Answer: "No entities found",
			Task:   plan.Command,
		}, nil
	}

	var result strings.Builder
	result.WriteString("Top entities:\n")
	for i, e := range entities {
		result.WriteString(fmt.Sprintf("%d. %s (confidence: %.2f)\n", i+1, e.Name, e.Confidence))
	}

	return &domain.ChatResponse{
		Answer:       result.String(),
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Search Command
type FilterFromVectorDBByFilter struct {
	Repo *repository.Repo
	LLM  *llm.OpenAIClient
}

func (c *FilterFromVectorDBByFilter) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract filter from args
	var filter string
	if filterVal, ok := plan.Args["filter"]; ok {
		if filterStr, ok := filterVal.(string); ok {
			filter = filterStr
		}
	}

	if filter == "" {
		return &domain.ChatResponse{
			Answer: "Filter required for article search",
			Task:   plan.Command,
		}, nil
	}

	// Embed filter and search vector DB
	embedding, err := c.LLM.Embed(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	arts, err := c.Repo.GetArticlesByVectorSearch(ctx, embedding, 5, []string{})
	if err != nil {
		return nil, err
	}

	if len(arts) == 0 {
		return &domain.ChatResponse{
			Answer: "No articles found for the given filter",
			Task:   plan.Command,
		}, nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Articles about %s:\n", filter))
	for i, a := range arts {
		result.WriteString(fmt.Sprintf("%d. %s\n   %s\n", i+1, a.Title, a.URL))
	}

	// Convert articles to sources
	var sources []domain.Source
	for _, a := range arts {
		sources = append(sources, domain.Source{
			ID:    a.ID,
			URL:   a.URL,
			Title: a.Title,
		})
	}

	return &domain.ChatResponse{
		Answer:       result.String(),
		ResponseType: domain.ResponseArticleList,
		Task:         plan.Command,
		Sources:      sources,
	}, nil
}
