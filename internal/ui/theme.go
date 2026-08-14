package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	Title        = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#58A6FF"}
	Muted        = lipgloss.AdaptiveColor{Light: "#6E7781", Dark: "#8B949E"}
	String       = lipgloss.AdaptiveColor{Light: "#1F2328", Dark: "#E6EDF3"}
	Number       = lipgloss.AdaptiveColor{Light: "#9A6700", Dark: "#E3B341"}
	Assignee     = lipgloss.AdaptiveColor{Light: "#8250DF", Dark: "#BC8CFF"}
	Success      = lipgloss.AdaptiveColor{Light: "#1A7F37", Dark: "#3FB950"}
	Danger       = lipgloss.AdaptiveColor{Light: "#CF222E", Dark: "#F85149"}
	Warn         = lipgloss.AdaptiveColor{Light: "#9A6700", Dark: "#D29922"}
	Merged       = lipgloss.AdaptiveColor{Light: "#8250DF", Dark: "#A371F7"}
	Draft        = lipgloss.AdaptiveColor{Light: "#6E7781", Dark: "#8B949E"}
	Border       = lipgloss.AdaptiveColor{Light: "#D0D7DE", Dark: "#2A3038"}
	BorderActive = lipgloss.AdaptiveColor{Light: "#0969DA", Dark: "#4C86F0"}
	Surface      = lipgloss.AdaptiveColor{Light: "#DDEBFF", Dark: "#12233A"}
	FundoSel     = Surface
)

var heroGradient = []string{"#3BE8B0", "#4C86F0", "#8A63F4"}

func SetTheme(escuro bool) {
	if escuro {
		heroGradient = []string{"#3BE8B0", "#4C86F0", "#8A63F4"}
	} else {
		heroGradient = []string{"#1A9E7A", "#0969DA", "#7A3FF2"}
	}
}

func HeroGradient() (string, string, string) {
	return heroGradient[0], heroGradient[1], heroGradient[2]
}

var (
	StyleTitle  = lipgloss.NewStyle().Foreground(Title).Bold(true)
	StyleMuted  = lipgloss.NewStyle().Foreground(Muted)
	StyleText   = lipgloss.NewStyle().Foreground(String)
	StyleNumber = lipgloss.NewStyle().Foreground(Number)
	StyleKey    = lipgloss.NewStyle().Foreground(Title).Bold(true)

	StyleTabActive   = lipgloss.NewStyle().Foreground(Title).Bold(true).Underline(true)
	StyleTabInactive = lipgloss.NewStyle().Foreground(Muted)

	StyleHeader = lipgloss.NewStyle().Foreground(Muted).Bold(true)
	StyleAccent = lipgloss.NewStyle().Foreground(Title)

	StyleSelected = lipgloss.NewStyle().Background(Surface)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)
	StyleBorderActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(BorderActive)

	StyleToastOk  = lipgloss.NewStyle().Foreground(Success).Bold(true)
	StyleToastErr = lipgloss.NewStyle().Foreground(Danger).Bold(true)
	StyleWarn     = lipgloss.NewStyle().Foreground(Warn).Bold(true)
)

func LabelColor(hex string) lipgloss.Color {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return lipgloss.Color("#FFFFFF")
	}
	r, g, b := hexPair(hex[0:2]), hexPair(hex[2:4]), hexPair(hex[4:6])
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	if lum < 60 {
		clarear := func(c int) int {
			v := int(float64(c)*0.5) + 110
			if v > 255 {
				return 255
			}
			return v
		}
		r, g, b = clarear(r), clarear(g), clarear(b)
	}
	return lipgloss.Color("#" + hx(r) + hx(g) + hx(b))
}

func hexPair(s string) int {
	n, err := strconv.ParseInt(s, 16, 0)
	if err != nil {
		return 255
	}
	return int(n)
}

func hx(n int) string {
	s := strconv.FormatInt(int64(n), 16)
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

func Pad(s string, width int) string {
	w := ansi.StringWidth(s)
	if w >= width {
		return ansi.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-w)
}
