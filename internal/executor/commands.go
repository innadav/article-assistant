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
	Repo              *repository.Repo
	ResponseGenerator *ResponseGenerator
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
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "Article URL required for summary"), nil
	}

	// Get article by URL
	articles, err := c.Repo.GetArticlesByURLs(ctx, []string{targetURL})
	if err != nil {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "Error retrieving article: "+targetURL), nil
	}

	if len(articles) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "Article not found: "+targetURL), nil
	}

	return c.ResponseGenerator.CreateSingleArticleResponse(ctx, articles[0].Summary, plan.Command, &articles[0])
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

// KeywordsOrTopics Command
type FetchKeywordsOrTopicsCommand struct {
	Repo              *repository.Repo
	ResponseGenerator *ResponseGenerator
}

func (c *FetchKeywordsOrTopicsCommand) Execute(ctx context.Context, plan *domain.Plan, _ string) (*domain.ChatResponse, error) {
	targetURLs := extractURLs(plan)
	if len(targetURLs) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "URLs required to extract keywords/topics"), nil
	}

	keywords, topics, err := c.Repo.GetKeywordsAndTopics(ctx, targetURLs, 5)
	if err != nil {
		return nil, err
	}

	if len(keywords) == 0 && len(topics) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "No keywords/topics found"), nil
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

	return c.ResponseGenerator.CreateTextResponse(ctx, result.String(), plan.Command, targetURLs)
}

// Sentiment Command
type FetchSentimentCommand struct {
	Repo              *repository.Repo
	ResponseGenerator *ResponseGenerator
}

func (c *FetchSentimentCommand) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
	// Extract URLs from args
	targetURLs := extractURLs(plan)
	if len(targetURLs) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "URLs required for sentiment analysis"), nil
	}

	// Fetch articles by URLs to get sentiment data
	arts, err := c.Repo.GetArticlesByURLs(ctx, targetURLs)
	if err != nil {
		return nil, err
	}

	if len(arts) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "No articles found for the provided URLs"), nil
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

	// Create sources from articles
	var sources []domain.Source
	for _, article := range arts {
		sources = append(sources, domain.Source{
			ID:    article.ID,
			URL:   article.URL,
			Title: article.Title,
		})
	}

	return &domain.ChatResponse{
		Answer:       result,
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Compare Command
type CompareCommand struct {
	Repo              *repository.Repo
	LLM               *llm.OpenAIClient
	ResponseGenerator *ResponseGenerator
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

	// Create sources from articles
	var sources []domain.Source
	for _, article := range articles {
		sources = append(sources, domain.Source{
			ID:    article.ID,
			URL:   article.URL,
			Title: article.Title,
		})
	}

	return &domain.ChatResponse{
		Answer:       comparison,
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Tone Command
type ToneKeyDfferencesCommand struct {
	Repo              *repository.Repo
	LLM               *llm.OpenAIClient
	ResponseGenerator *ResponseGenerator
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

	// Create sources from articles
	var sources []domain.Source
	for _, article := range articles {
		sources = append(sources, domain.Source{
			ID:    article.ID,
			URL:   article.URL,
			Title: article.Title,
		})
	}

	return &domain.ChatResponse{
		Answer:       toneDiff,
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// MorePositive Command
type FetchMostPositivesByFilter struct {
	Repo              *repository.Repo
	LLM               *llm.OpenAIClient
	ResponseGenerator *ResponseGenerator
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
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "Filter required for finding most positive article"), nil
	}

	// Step 1: Embed the filter and find similar articles
	embedding, err := c.LLM.Embed(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	candidates, err := c.Repo.GetArticlesByVectorSearch(ctx, embedding, 2, []string{})
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "No articles found for the given filter"), nil
	}

	// Step 2: LLM validation - filter candidates that actually discuss the topic
	var validatedCandidates []domain.Article
	for _, article := range candidates {
		prompt := fmt.Sprintf("Does this article explicitly discuss %s?\n\nTitle: %s\nSummary: %s\n\nAnswer with only 'YES' or 'NO'.",
			filter, article.Title, article.Summary)

		fmt.Printf("🔍 LLM Validation Prompt: %s\n", prompt)
		response, err := c.LLM.GenerateText(ctx, prompt)
		if err != nil {
			fmt.Printf("❌ LLM Error: %v\n", err)
			// Include article if LLM fails
			validatedCandidates = append(validatedCandidates, article)
			continue
		}
		fmt.Printf("🤖 LLM Response: %s\n", response)

		if strings.Contains(strings.ToUpper(response), "YES") {
			validatedCandidates = append(validatedCandidates, article)
		}
	}

	if len(validatedCandidates) == 0 {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, fmt.Sprintf("No articles found that explicitly discuss '%s'", filter)), nil
	}

	// Step 3: Find the article with the highest sentiment score among validated candidates
	var best *domain.Article
	bestScore := -1.0
	for _, a := range validatedCandidates {
		if a.SentimentScore > bestScore {
			bestScore = a.SentimentScore
			best = &a
		}
	}

	if best == nil {
		return c.ResponseGenerator.CreateErrorResponse(plan.Command, "No articles with sentiment data found"), nil
	}

	result := fmt.Sprintf("Most positive article about '%s' (validated from %d candidates):\n%s\nTitle: %s\nSentiment: %s (%.2f)",
		filter, len(validatedCandidates), best.URL, best.Title, best.Sentiment, best.SentimentScore)

	// Create sources from the best article
	sources := []domain.Source{
		{
			ID:    best.ID,
			URL:   best.URL,
			Title: best.Title,
		},
	}

	return &domain.ChatResponse{
		Answer:       result,
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// TopEntities Command
type FetchTopEntitiesFromDBCommand struct {
	Repo              *repository.Repo
	ResponseGenerator *ResponseGenerator
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

	// For top entities, we don't have specific article sources, but we can indicate
	// that this is aggregated data from all articles (or filtered articles)
	var sources []domain.Source
	if len(targetURLs) > 0 {
		// If URLs were provided, get those articles as sources
		articles, err := c.Repo.GetArticlesByURLs(ctx, targetURLs)
		if err == nil {
			for _, article := range articles {
				sources = append(sources, domain.Source{
					ID:    article.ID,
					URL:   article.URL,
					Title: article.Title,
				})
			}
		}
	}
	// If no URLs provided or error getting articles, sources will be empty
	// which indicates this is aggregated data from all articles

	return &domain.ChatResponse{
		Answer:       result.String(),
		Sources:      sources,
		ResponseType: domain.ResponseText,
		Task:         plan.Command,
	}, nil
}

// Search Command
type FetchArticlesDiscussingSpecificTopic struct {
	Repo              *repository.Repo
	LLM               *llm.OpenAIClient
	ResponseGenerator *ResponseGenerator
}

func (c *FetchArticlesDiscussingSpecificTopic) Execute(ctx context.Context, plan *domain.Plan, query string) (*domain.ChatResponse, error) {
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

	arts, err := c.Repo.GetArticlesByVectorSearch(ctx, embedding, 2, []string{})
	if err != nil {
		return nil, err
	}

	fmt.Printf("🔍 Vector search found %d articles for filter: %s\n", len(arts), filter)

	if len(arts) == 0 {
		return &domain.ChatResponse{
			Answer: "No articles found for the given filter",
			Task:   plan.Command,
		}, nil
	}

	// Filter articles using LLM to check if they actually discuss the topic
	var filteredArticles []domain.Article
	for _, article := range arts {
		// Create prompt to check if article discusses the filter topic
		prompt := fmt.Sprintf("Does this article explicitly discuss %s?\n\nTitle: %s\nSummary: %s\n\nAnswer with only 'YES' or 'NO'.",
			filter, article.Title, article.Summary)

		fmt.Printf("🔍 LLM Verification Prompt: %s\n", prompt)

		response, err := c.LLM.GenerateText(ctx, prompt)
		if err != nil {
			fmt.Printf("❌ LLM Error: %v\n", err)
			// If LLM call fails, include the article to be safe
			filteredArticles = append(filteredArticles, article)
			continue
		}

		fmt.Printf("🤖 LLM Response: %s\n", response)

		// Check if LLM response indicates the article discusses the topic
		if strings.Contains(strings.ToUpper(response), "YES") {
			filteredArticles = append(filteredArticles, article)
		}
	}

	if len(filteredArticles) == 0 {
		return &domain.ChatResponse{
			Answer: fmt.Sprintf("No articles found that explicitly discuss %s", filter),
			Task:   plan.Command,
		}, nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Articles about %s:\n", filter))
	for i, a := range filteredArticles {
		result.WriteString(fmt.Sprintf("%d. %s\n   %s\n", i+1, a.Title, a.URL))
	}

	// Convert articles to sources
	var sources []domain.Source
	for _, a := range filteredArticles {
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
