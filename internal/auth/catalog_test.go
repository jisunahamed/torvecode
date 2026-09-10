package auth

import "testing"

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
