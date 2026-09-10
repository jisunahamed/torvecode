package auth

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jisunahamed/torvecode/internal/llm/models"
)

func TestCatalogCacheIsScopedToCredentialAndExpires(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	credential := Credential{AccessToken: "token-a"}
	entries := []CatalogModel{{ID: "model-a", Protocol: "openai", Operation: "chat.completions"}}
	if err := saveCatalogCache(credential, entries); err != nil {
		t.Fatal(err)
	}
	loaded, ok := loadCatalogCache(credential)
	if !ok || len(loaded) != 1 || loaded[0].ID != "model-a" {
		t.Fatalf("loadCatalogCache() = %#v, %v", loaded, ok)
	}
	if _, ok := loadCatalogCache(Credential{AccessToken: "token-b"}); ok {
		t.Fatal("cache was reused for another credential")
	}
	path, err := catalogCachePath()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cache catalogCache
	if err := json.Unmarshal(payload, &cache); err != nil {
		t.Fatal(err)
	}
	cache.FetchedAt = time.Now().Add(-catalogCacheTTL - time.Minute)
	payload, _ = json.Marshal(cache)
	if err := os.WriteFile(path, payload, 0600); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadCatalogCache(credential); ok {
		t.Fatal("expired cache was reused")
	}
}

func TestRegisterModelsAcceptsOpenAIChatCompatibilityWithoutToolFlag(t *testing.T) {
	first, err := RegisterModels([]CatalogModel{{
		ID:              "torve-model",
		DisplayName:     "Torve Model",
		Protocol:        "openai",
		Operation:       "chat.completions",
		ContextWindow:   128000,
		MaxOutputTokens: 8192,
		ToolUse:         false,
	}})
	if err != nil {
		t.Fatalf("RegisterModels returned an error: %v", err)
	}
	if first != "torve-model" {
		t.Fatalf("first model = %q, want %q", first, "torve-model")
	}
	if !models.SupportedModels[first].SupportsAttachments {
		t.Fatal("OpenAI-compatible Torve model did not enable image attachments")
	}
}

func TestRegisterModelsRejectsUnsupportedOperations(t *testing.T) {
	_, err := RegisterModels([]CatalogModel{{
		ID:        "responses-only",
		Protocol:  "openai",
		Operation: "responses",
		ToolUse:   true,
	}})
	if err == nil {
		t.Fatal("RegisterModels accepted an unsupported operation")
	}
}

func TestRegisterModelsConvertsMicrousdPricing(t *testing.T) {
	first, err := RegisterModels([]CatalogModel{{
		ID: "priced-model", DisplayName: "Priced", Protocol: "openai", Operation: "chat.completions",
		InputPrice: "3000000", OutputPrice: "15000000", CachedInputPrice: "300000",
	}})
	if err != nil {
		t.Fatal(err)
	}
	model := models.SupportedModels[first]
	if model.CostPer1MIn != 3 || model.CostPer1MOut != 15 || model.CostPer1MInCached != 0.3 {
		t.Fatalf("unexpected model prices: %#v", model)
	}
}
