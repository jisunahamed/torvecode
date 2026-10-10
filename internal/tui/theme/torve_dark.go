package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// TorveDarkTheme — reference image palette: #0B0D0F bg, lime #A3E635 accent.
type TorveDarkTheme struct {
	BaseTheme
}

func NewTorveDarkTheme() *TorveDarkTheme {
	darkBackground := "#0B0D0F"
	darkSurface := "#15191D"
	darkSurfaceHover := "#1B2025"
	darkForeground := "#E8EAED"
	darkMuted := "#77818C"
	darkSecondary := "#A0A8B3"
	darkPrimary := "#A3E635"
	darkPrimaryDim := "#86C530"
	darkSuccess := "#A3E635"
	darkError := "#F87171"
	darkWarning := "#FBBF24"
	darkInfo := "#60A5FA"
	darkBorder := "#30363D"
	darkBorderDim := "#21262D"

	lightBackground := "#F9FAFB"
	lightSurface := "#F1F5F9"
	lightSurfaceHover := "#E2E8F0"
	lightForeground := "#0F172A"
	lightMuted := "#64748B"
	lightSecondary := "#475569"
	lightPrimary := "#65A30D"
	lightSuccess := "#65A30D"
	lightError := "#DC2626"
	lightWarning := "#D97706"
	lightInfo := "#0284C7"
	lightBorder := "#E2E8F0"
	lightBorderDim := "#F1F5F9"

	t := &TorveDarkTheme{}

	t.PrimaryColor = lipgloss.AdaptiveColor{Dark: darkPrimary, Light: lightPrimary}
	t.SecondaryColor = lipgloss.AdaptiveColor{Dark: darkSecondary, Light: lightSecondary}
	t.AccentColor = lipgloss.AdaptiveColor{Dark: darkPrimaryDim, Light: lightPrimary}

	t.ErrorColor = lipgloss.AdaptiveColor{Dark: darkError, Light: lightError}
	t.WarningColor = lipgloss.AdaptiveColor{Dark: darkWarning, Light: lightWarning}
	t.SuccessColor = lipgloss.AdaptiveColor{Dark: darkSuccess, Light: lightSuccess}
	t.InfoColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}

	t.TextColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}
	t.TextMutedColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.TextEmphasizedColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}

	t.BackgroundColor = lipgloss.AdaptiveColor{Dark: darkBackground, Light: lightBackground}
	t.BackgroundSecondaryColor = lipgloss.AdaptiveColor{Dark: darkSurface, Light: lightSurface}
	t.BackgroundDarkerColor = lipgloss.AdaptiveColor{Dark: darkSurfaceHover, Light: lightSurfaceHover}

	t.BorderNormalColor = lipgloss.AdaptiveColor{Dark: darkBorder, Light: lightBorder}
	t.BorderFocusedColor = lipgloss.AdaptiveColor{Dark: darkPrimary, Light: lightPrimary}
	t.BorderDimColor = lipgloss.AdaptiveColor{Dark: darkBorderDim, Light: lightBorderDim}

	t.DiffAddedColor = lipgloss.AdaptiveColor{Dark: "#A3E635", Light: "#65A30D"}
	t.DiffRemovedColor = lipgloss.AdaptiveColor{Dark: "#F87171", Light: "#DC2626"}
	t.DiffContextColor = lipgloss.AdaptiveColor{Dark: "#77818C", Light: "#64748B"}
	t.DiffHunkHeaderColor = lipgloss.AdaptiveColor{Dark: "#77818C", Light: "#64748B"}
	t.DiffHighlightAddedColor = lipgloss.AdaptiveColor{Dark: "#1A2E05", Light: "#DCFCE7"}
	t.DiffHighlightRemovedColor = lipgloss.AdaptiveColor{Dark: "#2E0A0A", Light: "#FEE2E2"}
	t.DiffAddedBgColor = lipgloss.AdaptiveColor{Dark: "#132A0A", Light: "#F0FDF4"}
	t.DiffRemovedBgColor = lipgloss.AdaptiveColor{Dark: "#2E0A0A", Light: "#FEF2F2"}
	t.DiffContextBgColor = lipgloss.AdaptiveColor{Dark: darkBackground, Light: lightBackground}
	t.DiffLineNumberColor = lipgloss.AdaptiveColor{Dark: "#77818C", Light: "#64748B"}
	t.DiffAddedLineNumberBgColor = lipgloss.AdaptiveColor{Dark: "#132A0A", Light: "#DCFCE7"}
	t.DiffRemovedLineNumberBgColor = lipgloss.AdaptiveColor{Dark: "#2E0A0A", Light: "#FEE2E2"}

	t.MarkdownTextColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}
	t.MarkdownHeadingColor = lipgloss.AdaptiveColor{Dark: darkPrimary, Light: lightPrimary}
	t.MarkdownLinkColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}
	t.MarkdownLinkTextColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}
	t.MarkdownCodeColor = lipgloss.AdaptiveColor{Dark: "#86EFAC", Light: "#166534"}
	t.MarkdownBlockQuoteColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.MarkdownEmphColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}
	t.MarkdownStrongColor = lipgloss.AdaptiveColor{Dark: darkPrimary, Light: lightPrimary}
	t.MarkdownHorizontalRuleColor = lipgloss.AdaptiveColor{Dark: darkBorder, Light: lightBorder}
	t.MarkdownListItemColor = lipgloss.AdaptiveColor{Dark: darkPrimary, Light: lightPrimary}
	t.MarkdownListEnumerationColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.MarkdownImageColor = lipgloss.AdaptiveColor{Dark: darkInfo, Light: lightInfo}
	t.MarkdownImageTextColor = lipgloss.AdaptiveColor{Dark: darkMuted, Light: lightMuted}
	t.MarkdownCodeBlockColor = lipgloss.AdaptiveColor{Dark: darkForeground, Light: lightForeground}

	t.SyntaxCommentColor = lipgloss.AdaptiveColor{Dark: "#77818C", Light: "#64748B"}
	t.SyntaxKeywordColor = lipgloss.AdaptiveColor{Dark: "#A3E635", Light: "#65A30D"}
	t.SyntaxFunctionColor = lipgloss.AdaptiveColor{Dark: "#60A5FA", Light: "#0284C7"}
	t.SyntaxVariableColor = lipgloss.AdaptiveColor{Dark: "#E8EAED", Light: "#0F172A"}
	t.SyntaxStringColor = lipgloss.AdaptiveColor{Dark: "#86EFAC", Light: "#166534"}
	t.SyntaxNumberColor = lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#D97706"}
	t.SyntaxTypeColor = lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#7C3AED"}
	t.SyntaxOperatorColor = lipgloss.AdaptiveColor{Dark: "#F472B6", Light: "#DB2777"}
	t.SyntaxPunctuationColor = lipgloss.AdaptiveColor{Dark: "#77818C", Light: "#64748B"}

	return t
}

func init() {
	RegisterTheme("torve", NewTorveDarkTheme())
}
