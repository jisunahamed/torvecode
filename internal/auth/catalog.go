package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jisunahamed/torvecode/internal/llm/models"
)

type CatalogModel struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	Protocol        string `json:"protocol"`
	Operation       string `json:"operation"`
	ContextWindow   int64  `json:"context_window"`
	MaxOutputTokens int64  `json:"max_output_tokens"`
	ToolUse         bool   `json:"tool_use"`
}

func FetchModels(ctx context.Context, credential Credential) ([]CatalogModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, APIURL()+"/cli/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+credential.AccessToken)
	req.Header.Set("accept", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("model catalog request failed: HTTP %d", resp.StatusCode)
	}
	var envelope struct {
		Data []CatalogModel `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func RegisterModels(entries []CatalogModel) (models.ModelID, error) {
	converted := make([]models.Model, 0, len(entries))
	for _, item := range entries {
		if !item.ToolUse {
			continue
		}
		provider := models.ProviderTorve
		if item.Protocol == "anthropic" && item.Operation == "messages" {
			provider = models.ProviderTorveAnthropic
		} else if item.Protocol != "openai" || item.Operation != "chat.completions" {
			continue
		}
		converted = append(converted, models.Model{ID: models.ModelID(item.ID), Name: item.DisplayName, Provider: provider, APIModel: item.ID, ContextWindow: item.ContextWindow, DefaultMaxTokens: item.MaxOutputTokens})
	}
	if len(converted) == 0 {
		return "", fmt.Errorf("your account has no tool-capable OpenAI-compatible models")
	}
	models.RegisterTorveModels(converted)
	return converted[0].ID, nil
}
