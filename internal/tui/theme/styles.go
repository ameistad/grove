package theme

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#7dbd8a")
	ColorAccent    = lipgloss.Color("#a8d8b9")
	ColorSecondary = lipgloss.Color("#6c9bd2")
	ColorSubtle    = lipgloss.Color("#7a7a7a")
	ColorWarning   = lipgloss.Color("#e0a458")
	ColorDanger    = lipgloss.Color("#d46a6a")
	ColorMuted     = lipgloss.Color("#555555")
	ColorFg        = lipgloss.Color("#d4d4d4")
	ColorFgDim     = lipgloss.Color("#888888")
	ColorBorder    = lipgloss.Color("#444444")
	ColorBg        = lipgloss.Color("#1a1a1a")
	ColorBgSurface = lipgloss.Color("#252525")

	StyleLogo = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	StyleSelected = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleNormal = lipgloss.NewStyle().
			Foreground(ColorFg)

	StyleDirty = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleHelpKey = lipgloss.NewStyle().
			Foreground(ColorFgDim)

	StyleHelpSep = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleDivider = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorDanger)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleBranch = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StyleHarness = lipgloss.NewStyle().
			Foreground(ColorFgDim)

	StyleCursor = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleEmptyState = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	StyleSectionLabel = lipgloss.NewStyle().
				Foreground(ColorSubtle).
				Bold(true)
)
