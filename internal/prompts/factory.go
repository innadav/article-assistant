package prompts

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"article-assistant/internal/article"
)

type ModelType string

const (
	ModelGemini15Flash ModelType = "gemini-1.5-flash"
)

// Factory loads and executes prompt templates.
type Factory struct {
	model     ModelType
	loader    *Loader
	templates map[string]*template.Template
}

// NewFactory parses all loaded YAML templates and prepares them for execution.
func NewFactory(model ModelType, loader *Loader) (*Factory, error) {
	templates := make(map[string]*template.Template)
	for name, p := range loader.Prompts {
		// Use Funcs to add custom functions if needed in the future
		tmpl, err := template.New(name).Parse(p.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		templates[name] = tmpl
	}
	return &Factory{
		model:     model,
		loader:    loader,
		templates: templates,
	}, nil
}

// executeTemplate is a generic helper to run a parsed template with given data.
func (f *Factory) executeTemplate(name string, data interface{}) (string, error) {
	tmpl, ok := f.templates[name]
	if !ok {
		return "", fmt.Errorf("template '%s' not found", name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	return buf.String(), nil
}

func (f *Factory) CreatePlannerPrompt(query string, articles []*article.Article) (string, error) {
	var articleContext strings.Builder
	for _, art := range articles {
		fmt.Fprintf(&articleContext, "- URL: %s, Title: %s\n", art.URL, art.Title)
	}
	data := map[string]interface{}{
		"Articles": articleContext.String(),
		"Query":    query,
	}
	return f.executeTemplate("planner", data)
}

func (f *Factory) CreateSummarizePrompt(content string) (string, error) {
	data := map[string]interface{}{"Content": content}
	return f.executeTemplate("summarize", data)
}

func (f *Factory) CreateKeywordsPrompt(title string) (string, error) {
	data := map[string]interface{}{"Title": title}
	return f.executeTemplate("keywords", data)
}

func (f *Factory) CreateSentimentPrompt(title string) (string, error) {
	data := map[string]interface{}{"Title": title}
	return f.executeTemplate("sentiment", data)
}

func (f *Factory) CreateCompareTonePrompt(article1 *article.Article, article2 *article.Article) (string, error) {
	data := map[string]interface{}{
		"Article1Title": article1.Title,
		"Article2Title": article2.Title,
	}
	return f.executeTemplate("compare_tone", data)
}

func (f *Factory) CreateFindTopicPrompt(topic string, allArticles []*article.Article) (string, error) {
	var articleContext strings.Builder
	for _, art := range allArticles {
		fmt.Fprintf(&articleContext, "- Title: %s, Excerpt: %s\n", art.Title, art.Excerpt)
	}
	data := map[string]interface{}{
		"Topic":    topic,
		"Articles": articleContext.String(),
	}
	return f.executeTemplate("find_topic", data)
}

func (f *Factory) CreateComparePositivityPrompt(topic string, article1 *article.Article, article2 *article.Article) (string, error) {
	data := map[string]interface{}{
		"Topic":           topic,
		"Article1Title":   article1.Title,
		"Article1Excerpt": article1.Excerpt,
		"Article2Title":   article2.Title,
		"Article2Excerpt": article2.Excerpt,
	}
	return f.executeTemplate("compare_positivity", data)
}

func (f *Factory) CreateFindCommonEntitiesPrompt(allArticles []*article.Article) (string, error) {
	var titleList []string
	for _, art := range allArticles {
		titleList = append(titleList, art.Title)
	}
	data := map[string]interface{}{
		"Titles": strings.Join(titleList, "\n- "),
	}
	return f.executeTemplate("find_common_entities", data)
}
