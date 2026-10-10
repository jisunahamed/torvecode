package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/jisunahamed/torvecode/internal/config"
)

type skillTool struct{}

const SkillToolName = "skill"

func NewSkillTool() BaseTool { return &skillTool{} }

func (s *skillTool) Info() ToolInfo {
	return ToolInfo{
		Name:        SkillToolName,
		Description: "Read a skill's full SKILL.md by name. Use when user requests a skill or you need its instructions.",
		Parameters: map[string]any{
			"name": map[string]any{"type": "string", "description": "Skill name"},
		},
		Required: []string{"name"},
	}
}

func (s *skillTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	// Minimal JSON parse without import cycle - extract "name" field.
	name := extractJSONField(call.Input, "name")
	if name == "" {
		return NewTextErrorResponse("skill name is required"), nil
	}
	wd := config.WorkingDirectory()
	candidates := []string{
		filepath.Join(wd, ".torvecode", "skills", name, "SKILL.md"),
	}
	if dir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(dir, "torvecode", "skills", name, "SKILL.md"))
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			content := strings.TrimSpace(string(data))
			if len(content) > 12000 {
				content = content[:12000] + "\n\n[truncated]"
			}
			return NewTextResponse(content), nil
		}
	}
	return NewTextErrorResponse("skill not found: " + name), nil
}

func extractJSONField(input, field string) string {
	// Tiny helper to avoid importing encoding/json just for one field in this hot path.
	// Falls back to simple scan.
	needle := `"` + field + `"`
	idx := strings.Index(input, needle)
	if idx < 0 {
		return ""
	}
	rest := input[idx+len(needle):]
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return ""
	}
	rest = strings.TrimSpace(rest[colon+1:])
	if len(rest) == 0 || rest[0] != '"' {
		return ""
	}
	rest = rest[1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}
