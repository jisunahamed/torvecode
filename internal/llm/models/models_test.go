package models

import "testing"

func TestEffectiveContextWindowHonorsPlanTPM(t *testing.T) {
	model := Model{ContextWindow: 1_000_000, PlanTPM: 40_000}
	if got := model.EffectiveContextWindow(); got != 30_000 {
		t.Fatalf("EffectiveContextWindow() = %d, want 30000", got)
	}
	model = Model{ContextWindow: 16_000, PlanTPM: 40_000}
	if got := model.EffectiveContextWindow(); got != 16_000 {
		t.Fatalf("EffectiveContextWindow() = %d, want 16000", got)
	}
}
