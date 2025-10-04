package repository

import (
	"article-assistant/internal/article"
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

const ArticleClassName = "Article"

// VectorRepository handles all communication with the Weaviate vector database.
type VectorRepository struct {
	client *weaviate.Client
}

// NewVectorRepository creates and initializes a new Weaviate client.
func NewVectorRepository(host, port string) (*VectorRepository, error) {
	cfg := weaviate.Config{
		Host:   host,
		Scheme: "http", // Assuming local, non-https connection
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create weaviate client: %w", err)
	}

	repo := &VectorRepository{client: client}
	// Ensure the required schema exists in Weaviate when the service starts.
	if err := repo.ensureSchemaExists(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure weaviate schema: %w", err)
	}

	return repo, nil
}

// ensureSchemaExists creates the "Article" class in Weaviate if it doesn't already exist.
func (r *VectorRepository) ensureSchemaExists(ctx context.Context) error {
	exists, err := r.client.Schema().ClassExistenceChecker().WithClassName(ArticleClassName).Do(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil // Schema is already in place.
	}

	// Define the class schema for storing articles.
	classObj := &models.Class{
		Class: ArticleClassName,
		Properties: []*models.Property{
			{
				Name:     "url",
				DataType: []string{"text"},
			},
			{
				Name:     "title",
				DataType: []string{"text"},
			},
		},
	}

	return r.client.Schema().ClassCreator().WithClass(classObj).Do(ctx)
}

// SaveArticleVector saves an article's vector representation to Weaviate.
func (r *VectorRepository) SaveArticleVector(ctx context.Context, art *article.Article, vector []float32) error {
	properties := map[string]interface{}{
		"url":   art.URL,
		"title": art.Title,
	}

	_, err := r.client.Data().Creator().
		WithClassName(ArticleClassName).
		WithProperties(properties).
		WithVector(vector).Do(ctx)

	return err
}

// SearchSimilarArticles finds articles with similar content based on a query vector.
func (r *VectorRepository) SearchSimilarArticles(ctx context.Context, queryVector []float32, limit int) ([]*article.Article, error) {
	// Perform the "nearVector" search.
	result, err := r.client.GraphQL().Get().WithClassName(ArticleClassName).WithNearVector(r.client.GraphQL().NearVectorArgBuilder().WithVector(queryVector)).WithLimit(limit).WithFields(graphql.Field{Name: "url"}, graphql.Field{Name: "title"}).Do(ctx)
	if err != nil {
		return nil, err
	}

	// Parse the GraphQL response.
	var articles []*article.Article
	get := result.Data["Get"].(map[string]interface{})
	items := get[ArticleClassName].([]interface{})

	for _, item := range items {
		itemMap := item.(map[string]interface{})
		articles = append(articles, &article.Article{
			URL:   itemMap["url"].(string),
			Title: itemMap["title"].(string),
		})
	}

	return articles, nil
}
