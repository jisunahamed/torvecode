package core

import (
	"strings"
	"testing"
)

func TestFormatTokensAndCostShowsInputOutputAndContext(t *testing.T) {
	formatted := formatTokensAndCost(20_441, 733, 128_000, 0)
	for _, want := range []string{"IN 20.4K", "OUT 733", "21.2K/128K", "(16%)"} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("formatTokensAndCost() = %q, missing %q", formatted, want)
		}
	}
	if strings.Contains(formatted, "$0.00") {
		t.Fatalf("unknown zero price was presented as a measured cost: %q", formatted)
	}
}
