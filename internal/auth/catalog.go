package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jisunahamed/torvecode/internal/llm/models"
)

// A successful account catalog is stable enough to reuse between launches.
// Explicit `torve models` requests still refresh it immediately.
const catalogCacheTTL = 24 * time.Hour
const catalogCacheVersion = 2

type CatalogModel struct {
	ID               string `json:"id"`
	DisplayName      string `json:"display_name"`
	Protocol         string `json:"protocol"`
	Operation        string `json:"operation"`
	ContextWindow    int64  `json:"context_window"`
	MaxOutputTokens  int64  `json:"max_output_tokens"`
	ToolUse          bool   `json:"tool_use"`
	InputPrice       string `json:"input_microusd_per_million"`
	OutputPrice      string `json:"output_microusd_per_million"`
	CachedInputPrice string `json:"cached_input_microusd_per_million"`
	PlanTPM          int64  `json:"plan_tpm,omitempty"`
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
		Data   []CatalogModel `json:"data"`
		Limits struct {
			TPM int64 `json:"tpm"`
		} `json:"limits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	for index := range envelope.Data {
		envelope.Data[index].PlanTPM = envelope.Limits.TPM
	}
	_ = saveCatalogCache(credential, envelope.Data)
	return envelope.Data, nil
}

// FetchModelsCached keeps repeat interactive launches off the network. The cache
// is scoped to the active credential and time limited; explicit catalog commands
// still use FetchModels and refresh it.
func FetchModelsCached(ctx context.Context, credential Credential) ([]CatalogModel, error) {
	if entries, ok := loadCatalogCache(credential); ok {
		return entries, nil
	}
	return FetchModels(ctx, credential)
}

type catalogCache struct {
	Version    int            `json:"version"`
	Credential string         `json:"credential"`
	FetchedAt  time.Time      `json:"fetched_at"`
	Models     []CatalogModel `json:"models"`
}

func catalogCredentialID(credential Credential) string {
	sum := sha256.Sum256([]byte(credential.AccessToken))
	return fmt.Sprintf("%x", sum[:8])
}

func catalogCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "torvecode")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "models.json"), nil
}

func loadCatalogCache(credential Credential) ([]CatalogModel, bool) {
	path, err := catalogCachePath()
	if err != nil {
		return nil, false
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var cache catalogCache
	if json.Unmarshal(payload, &cache) != nil || cache.Version != catalogCacheVersion || cache.Credential != catalogCredentialID(credential) || time.Since(cache.FetchedAt) > catalogCacheTTL || len(cache.Models) == 0 {
		return nil, false
	}
	return cache.Models, true
}

func saveCatalogCache(credential Credential, entries []CatalogModel) error {
	if len(entries) == 0 {
		return nil
	}
	path, err := catalogCachePath()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(catalogCache{Version: catalogCacheVersion, Credential: catalogCredentialID(credential), FetchedAt: time.Now().UTC(), Models: entries})
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0600)
}

func RegisterModels(entries []CatalogModel) (models.ModelID, error) {
	converted := make([]models.Model, 0, len(entries))
	for _, item := range entries {
		provider := models.ProviderTorve
		if item.Protocol == "anthropic" && item.Operation == "messages" {
			if !item.ToolUse {
				continue
			}
			provider = models.ProviderTorveAnthropic
		} else if item.Protocol != "openai" || item.Operation != "chat.completions" {
			continue
		}
		converted = append(converted, models.Model{
			ID:                  models.ModelID(item.ID),
			Name:                item.DisplayName,
			Provider:            provider,
			APIModel:            item.ID,
			ContextWindow:       item.ContextWindow,
			DefaultMaxTokens:    item.MaxOutputTokens,
			CostPer1MIn:         priceUSD(item.InputPrice),
			CostPer1MOut:        priceUSD(item.OutputPrice),
			CostPer1MInCached:   priceUSD(item.CachedInputPrice),
			PlanTPM:             item.PlanTPM,
			SupportsAttachments: true,
		})
	}
	if len(converted) == 0 {
		return "", fmt.Errorf("your account has no compatible chat models")
	}
	models.RegisterTorveModels(converted)
	return converted[0].ID, nil
}

func priceUSD(microUSD string) float64 {
	value, err := strconv.ParseFloat(microUSD, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value / 1_000_000
}
