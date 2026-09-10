package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jisunahamed/torvecode/internal/config"
	"github.com/jisunahamed/torvecode/internal/llm/models"
	"github.com/jisunahamed/torvecode/internal/skills"
)

func GetAgentPrompt(agentName config.AgentName, provider models.ModelProvider) string {
	basePrompt := ""
	switch agentName {
	case config.AgentCoder:
		basePrompt = CoderPrompt(provider)
	case config.AgentTitle:
		basePrompt = TitlePrompt(provider)
	case config.AgentTask:
		basePrompt = TaskPrompt(provider)
	case config.AgentSummarizer:
		basePrompt = SummarizerPrompt(provider)
	default:
		basePrompt = "You are a helpful assistant"
	}

	if agentName == config.AgentCoder || agentName == config.AgentTask {
		// Add context from project-specific instruction files if they exist
		contextContent := getContextFromPaths()
		contextContent += skills.Catalog(config.WorkingDirectory())
		if contextContent != "" {
			return fmt.Sprintf("%s\n\n# Project-Specific Context\n Make sure to follow the instructions in the context below\n%s", basePrompt, contextContent)
		}
	}
	return basePrompt
}

var (
	onceContext    sync.Once
	contextContent string
)

func getContextFromPaths() string {
	onceContext.Do(func() {
		var (
			cfg          = config.Get()
			workDir      = cfg.WorkingDir
			contextPaths = cfg.ContextPaths
		)

		contextContent = processContextPaths(workDir, contextPaths)
	})

	return contextContent
}

func processContextPaths(workDir string, paths []string) string {
	processedFiles := make(map[string]bool)
	results := make([]string, 0)
	appendFile := func(path string) {
		key := strings.ToLower(filepath.Clean(path))
		if processedFiles[key] {
			return
		}
		processedFiles[key] = true
		if result := processFile(path); result != "" {
			results = append(results, result)
		}
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "/") || strings.HasSuffix(path, `\`) {
			_ = filepath.WalkDir(filepath.Join(workDir, path), func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !entry.IsDir() {
					appendFile(path)
				}
				return nil
			})
			continue
		}
		appendFile(filepath.Join(workDir, path))
	}

	return strings.Join(results, "\n")
}

func processFile(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return "# From:" + filePath + "\n" + string(content)
}
