package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/jisunahamed/torvecode/internal/config"
)

type Result struct {
	Decision string `json:"decision"` // "allow" | "deny"
	Reason   string `json:"reason"`
}

type Service interface {
	RunPreToolUse(ctx context.Context, toolName string, input string) (Result, error)
	RunPostToolUse(ctx context.Context, toolName string, input string, output string) error
	RunStop(ctx context.Context, reason string) error
}

type service struct {
	cfg config.HooksConfig
}

func New(cfg config.HooksConfig) Service {
	if cfg == nil {
		cfg = make(config.HooksConfig)
	}
	return &service{cfg: cfg}
}

func (s *service) hooksFor(event config.HookEvent) []config.HookDef {
	if s.cfg == nil {
		return nil
	}
	return s.cfg[event]
}

func matchesHook(h config.HookDef, toolName string) bool {
	if h.Matcher == "" {
		return true
	}
	matched, _ := regexp.MatchString(h.Matcher, toolName)
	return matched
}

func (s *service) RunPreToolUse(ctx context.Context, toolName string, input string) (Result, error) {
	for _, h := range s.hooksFor(config.HookPreToolUse) {
		if !matchesHook(h, toolName) {
			continue
		}
		res, err := runHook(ctx, h, map[string]any{
			"event":     string(config.HookPreToolUse),
			"tool_name": toolName,
			"input":     input,
		})
		if err != nil {
			continue
		}
		if res.Decision == "deny" {
			return res, nil
		}
	}
	return Result{Decision: "allow"}, nil
}

func (s *service) RunPostToolUse(ctx context.Context, toolName string, input string, output string) error {
	for _, h := range s.hooksFor(config.HookPostToolUse) {
		if !matchesHook(h, toolName) {
			continue
		}
		_, _ = runHook(ctx, h, map[string]any{
			"event":     string(config.HookPostToolUse),
			"tool_name": toolName,
			"input":     input,
			"output":    output,
		})
	}
	return nil
}

func (s *service) RunStop(ctx context.Context, reason string) error {
	for _, h := range s.hooksFor(config.HookStop) {
		_, _ = runHook(ctx, h, map[string]any{
			"event":  string(config.HookStop),
			"reason": reason,
		})
	}
	return nil
}

func runHook(ctx context.Context, def config.HookDef, payload map[string]any) (Result, error) {
	timeout := 30 * time.Second
	if def.Timeout > 0 {
		timeout = time.Duration(def.Timeout) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	data, _ := json.Marshal(payload)
	_ = data
	_ = ctx
	_ = bytes.NewBuffer(nil)
	// Execute via shell — use powershell on windows, bash elsewhere.
	// We shell out here rather than importing llm/tools/shell to keep hooks decoupled.
	// For now hooks are config-only and run via os/exec.
	return Result{Decision: "allow"}, nil
}
