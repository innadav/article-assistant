package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"article-assistant/internal/analysis"
	"article-assistant/internal/llm"
)

func main() {
	// Initialize LLM client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	llmClient := llm.New(apiKey)
	analysisService := analysis.NewAnalysisService(llmClient)

	ctx := context.Background()

	// Test content
	content := `
	OpenAI has announced a major breakthrough in artificial intelligence with the release of GPT-4. 
	The new model shows significant improvements in reasoning, creativity, and safety compared to previous versions. 
	Tech industry leaders like Elon Musk and Sam Altman have praised the advancement, calling it a milestone in AI development.
	The technology is expected to revolutionize various sectors including healthcare, education, and finance.
	`

	fmt.Println("🚀 Testing LLM Analysis Service")
	fmt.Println("================================")
	fmt.Println("Content:", content)
	fmt.Println()

	// Test 1: Comprehensive Analysis
	fmt.Println("1. Comprehensive Analysis:")
	result, err := analysisService.AnalyzeContent(ctx, content)
	if err != nil {
		log.Printf("Error in comprehensive analysis: %v", err)
	} else {
		fmt.Printf("Summary: %s\n", result.Summary)
		fmt.Printf("Entities: %+v\n", result.Entities)
		fmt.Printf("Keywords: %+v\n", result.Keywords)
		fmt.Printf("Topics: %+v\n", result.Topics)
	}
	fmt.Println()

	// Test 2: Entity Matching
	fmt.Println("2. Entity Matching:")
	referenceEntities := []string{"OpenAI", "Google", "Microsoft", "Tesla", "Apple", "Sam Altman", "Elon Musk"}
	entityMatches, err := analysisService.MatchEntities(ctx, content, referenceEntities)
	if err != nil {
		log.Printf("Error in entity matching: %v", err)
	} else {
		fmt.Printf("Entity Matches: %+v\n", entityMatches)
	}
	fmt.Println()

	// Test 3: Keyword Matching
	fmt.Println("3. Keyword Matching:")
	referenceKeywords := []string{"artificial intelligence", "machine learning", "technology", "innovation", "breakthrough", "healthcare", "education"}
	keywordMatches, err := analysisService.MatchKeywords(ctx, content, referenceKeywords)
	if err != nil {
		log.Printf("Error in keyword matching: %v", err)
	} else {
		fmt.Printf("Keyword Matches: %+v\n", keywordMatches)
	}
	fmt.Println()

	// Test 4: Topic Matching
	fmt.Println("4. Topic Matching:")
	referenceTopics := []string{"AI Development", "Technology Innovation", "Business Strategy", "Healthcare Technology", "Educational Technology"}
	topicMatches, err := analysisService.MatchTopics(ctx, content, referenceTopics)
	if err != nil {
		log.Printf("Error in topic matching: %v", err)
	} else {
		fmt.Printf("Topic Matches: %+v\n", topicMatches)
	}
	fmt.Println()

	// Test 5: Similarity Analysis
	fmt.Println("5. Similarity Analysis:")
	candidateContents := []string{
		"Google announced new AI capabilities in their search engine, focusing on better understanding user queries.",
		"Tesla is developing autonomous driving technology using advanced machine learning algorithms.",
		"Microsoft Azure offers cloud computing services for businesses worldwide.",
		"OpenAI's ChatGPT has revolutionized conversational AI and natural language processing.",
	}

	similarityMatches, err := analysisService.FindSimilarContent(ctx, content, candidateContents)
	if err != nil {
		log.Printf("Error in similarity analysis: %v", err)
	} else {
		fmt.Printf("Similarity Matches:\n")
		for _, match := range similarityMatches {
			fmt.Printf("  Index %d (Score: %.2f): %s...\n", match.Index, match.Similarity, match.Content[:50])
		}
	}

	fmt.Println("\n✅ Analysis service testing completed!")
}
