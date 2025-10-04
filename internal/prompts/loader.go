package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Prompt struct {
	Template string `yaml:"template"`
}

type Loader struct {
	Prompts map[string]Prompt
}

func NewLoader(version string) (*Loader, error) {
	promptDir := filepath.Join("configs", "prompts", version)
	files, err := os.ReadDir(promptDir)
	if err != nil {
		return nil, fmt.Errorf("could not read prompt directory '%s': %w", promptDir, err)
	}

	prompts := make(map[string]Prompt)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" {
			filePath := filepath.Join(promptDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read prompt file %s: %w", file.Name(), err)
			}

			var p Prompt
			if err := yaml.Unmarshal(data, &p); err != nil {
				return nil, fmt.Errorf("failed to parse prompt file %s: %w", file.Name(), err)
			}

			promptName := strings.TrimSuffix(file.Name(), ".yaml")
			prompts[promptName] = p
		}
	}
	return &Loader{Prompts: prompts}, nil
}

func (l *Loader) Get(name string) (string, error) {
	prompt, ok := l.Prompts[name]
	if !ok {
		return "", fmt.Errorf("prompt '%s' not found", name)
	}
	return prompt.Template, nil
}
