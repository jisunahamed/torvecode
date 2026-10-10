package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// TorveAmberTheme is the hero SaaS theme for Torvecode:
// warm amber primary on graphite/stone neutrals, not a terminal-hacker look.
type TorveAmberTheme struct {
	BaseTheme
}

func NewTorveAmberTheme() *TorveAmberTheme {
	// Amber / graphite / stone palette
	darkBackground := "#1A1A1E"
	darkSurface := "#252529"
	darkSurfaceHover := "#2E2E33"
	darkForeground := "#E7E5E4"
	darkMuted := "#A8A29E"
	darkAmber := "#F59E0B"
	darkAmberBright := "#FBBF24"
	darkStone := "#D6D3D1"
	darkError := "#F87171"
	darkWarning := "#FBBF24"
	darkSuccess := "#34D399"
	darkInfo := "#38BDF8"
	darkBorder := "#3A3936"
	darkBorderFocused := "#F59E0B"
	darkBorderDim := "#2E2E33"

	lightBackground := "#FDFBF7"
	lightSurface := "#F5F3EF"
	lightSurfaceHover := "#EDE9E3"
	lightForeground := "#1C1917"
	lightMuted := "#78716C"
	lightAmber := "#D97706"
	lightAmberBright := "#F59E0B"
	lightStone := "#44403C"
	lightError := "#DC2626"
	lightWarning := "#D97706"
	lightSuccess := "#059669"
	lightInfo := "#0284C7"
	lightBorder := "#E7E5E4"
	lightBorderFocused := "#D97706"
	lightBorderDim := "#F5F5F4"

	t := &TorveAmberTheme{}

	t.PrimaryColor = lipgloss.AdaptiveColor{Dark: darkAmber, Light: lightAmber}
	t.SecondaryColor = lipgloss.AdaptiveColor{Dark: darkStone, Light: lightStone}
	t.AccentColor = lipgloss.AdaptiveColor{Dark: darkAmberBright, Light: lightAmberBright}

	t.ErrorColor = lipgloss.AdaptiveColor{Dark: darkError, Light: lightError}
	t.WarningColor = lipgloss.AdaptiveColor{Dark: darkWarning, Light: lightWarning}
	t.SuccessColor = lipgloss.AdaptiveColor{Dark: darkSuccess, Light: lightSuccess}
	t.InfoColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}

	t.TextColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}
	t.TextMutedColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.TextEmphasizedColor = lipgloss.AdaptiveColor{Dark: darkAmberBright, Light: lightAmber}

	t.BackgroundColor = lipgloss.AdaptiveColor{Dark: darkBackground, Light: lightBackground}
	t.BackgroundSecondaryColor = lipgloss.AdaptiveColor{Dark: darkSurface, Light: lightSurface}
	t.BackgroundDarkerColor = lipgloss.AdaptiveColor{Dark: darkSurfaceHover, Light: lightSurfaceHover}

	t.BorderNormalColor = lipgloss.AdaptiveColor{Dark: darkBorder, Light: lightBorder}
	t.BorderFocusedColor = lipgloss.AdaptiveColor{Dark: darkBorderFocused, Light: lightBorderFocused}
	t.BorderDimColor = lipgloss.AdaptiveColor{Dark: darkBorderDim, Light: lightBorderDim}

	// Diff - warm, low-saturation
	t.DiffAddedColor = lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#059669"}
	t.DiffRemovedColor = lipgloss.AdaptiveColor{Dark: "#F87171", Light: "#DC2626"}
	t.DiffContextColor = lipgloss.AdaptiveColor{Dark: "#A8A29E", Light: "#78716C"}
	t.DiffHunkHeaderColor = lipgloss.AdaptiveColor{Dark: "#78716C", Light: "#A8A29E"}
	t.DiffHighlightAddedColor = lipgloss.AdaptiveColor{Dark: "#064E3B", Light: "#D1FAE5"}
	t.DiffHighlightRemovedColor = lipgloss.AdaptiveColor{Dark: "#7F1D1D", Light: "#FEE2E2"}
	t.DiffAddedBgColor = lipgloss.AdaptiveColor{Dark: "#1A2E26", Light: "#ECFDF5"}
	t.DiffRemovedBgColor = lipgloss.AdaptiveColor{Dark: "#2E1A1A", Light: "#FEF2F2"}
	t.DiffContextBgColor = lipgloss.AdaptiveColor{Dark: darkBackground, Light: lightBackground}
	t.DiffLineNumberColor = lipgloss.AdaptiveColor{Dark: "#78716C", Light: "#A8A29E"}
	t.DiffAddedLineNumberBgColor = lipgloss.AdaptiveColor{Dark: "#1A2E26", Light: "#D1FAE5"}
	t.DiffRemovedLineNumberBgColor = lipgloss.AdaptiveColor{Dark: "#2E1A1A", Light: "#FEE2E2"}

	// Markdown
	t.MarkdownTextColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}
	t.MarkdownHeadingColor = lipgloss.AdaptiveColor{Dark: darkAmber, Light: lightAmber}
	t.MarkdownLinkColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}
	t.MarkdownLinkTextColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}
	t.MarkdownCodeColor = lipgloss.AdaptiveColor{Dark: "#FDBA74", Light: "#9A3412"}
	t.MarkdownBlockQuoteColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.MarkdownEmphColor = lipgloss.AdaptiveColor{Dark: darkStone, Light: lightStone}
	t.MarkdownStrongColor = lipgloss.AdaptiveColor{Dark: darkAmberBright, Light: lightAmber}
	t.MarkdownHorizontalRuleColor = lipgloss.AdaptiveColor{Dark: darkBorder, Light: lightBorder}
	t.MarkdownListItemColor = lipgloss.AdaptiveColor{Dark: darkAmber, Light: lightAmber}
	t.MarkdownListEnumerationColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.MarkdownImageColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}
	t.MarkdownImageTextColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.MarkdownCodeBlockColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}

	// Syntax - warm, readable on graphite
	t.SyntaxCommentColor = lipgloss.AdaptiveColor{Dark: "#78716C", Light: "#A8A29E"}
	t.SyntaxKeywordColor = lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#D97706"}
	t.SyntaxFunctionColor = lipgloss.AdaptiveColor{Dark: "#38BDF8", Light: "#0284C7"}
	t.SyntaxVariableColor = lipgloss.AdaptiveColor{Dark: "#E7E5E4", Light: "#1C1917"}
	t.SyntaxStringColor = lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#059669"}
	t.SyntaxNumberColor = lipgloss.AdaptiveColor{Dark: "#FDBA74", Light: "#C2410C"}
	t.SyntaxTypeColor = lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#7C3AED"}
	t.SyntaxOperatorColor = lipgloss.AdaptiveColor{Dark: "#F472B6", Light: "#DB2777"}
	t.SyntaxPunctuationColor = lipgloss.AdaptiveColor{Dark: "#A8A29E", Light: "#78716C"}

	return t
}

func init() {
	RegisterTheme("torve-amber", NewTorveAmberTheme())
}
