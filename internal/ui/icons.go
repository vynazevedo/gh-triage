package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

var nerd = true

func UseNerdFonts(ativo bool) { nerd = ativo }

func pick(nf, plano string) string {
	if nerd {
		return nf
	}
	return plano
}

func ItemIcon(it gh.Item) (string, lipgloss.AdaptiveColor) {
	switch {
	case it.Kind == gh.KindPR && it.State == gh.StateOpen && it.Draft:
		return pick("", "◌"), Draft
	case it.Kind == gh.KindIssue && it.State == gh.StateOpen:
		return pick("", "○"), Success
	case it.Kind == gh.KindIssue:
		return pick("", "●"), Merged
	case it.Kind == gh.KindPR && it.State == gh.StateOpen:
		return pick("", "⇄"), Success
	case it.Kind == gh.KindPR && it.State == gh.StateMerged:
		return pick("", "✓"), Merged
	default:
		return pick("", "✗"), Danger
	}
}

func CIIcon(ci gh.CIStatus) (string, lipgloss.AdaptiveColor) {
	switch ci {
	case gh.CISuccess:
		return pick("", "✓"), Success
	case gh.CIFail:
		return pick("\U000f0159", "✗"), Danger
	case gh.CIPending:
		return pick("", "•"), Warn
	default:
		return "·", Muted
	}
}

func ReviewText(decisao string) (string, lipgloss.AdaptiveColor) {
	switch decisao {
	case "APPROVED":
		return pick("\U000f012c ", "") + "aprovado", Success
	case "CHANGES_REQUESTED":
		return pick(" ", "") + "mudanças", Danger
	case "REVIEW_REQUIRED":
		return pick(" ", "") + "pendente", Warn
	default:
		return "·", Muted
	}
}

func ReasonIcon(class int) (string, lipgloss.AdaptiveColor) {
	switch class {
	case 0:
		return pick("", "»"), Warn
	case 1:
		return pick("", "!"), Danger
	case 2:
		return pick("\U000f0159", "!"), Danger
	case 3:
		return pick("", "✓"), Success
	case 4:
		return pick("", "•"), Muted
	case 5:
		return pick("", "@"), Muted
	default:
		return pick("", "·"), Muted
	}
}

func RepoIcon() string    { return pick(" ", "") }
func PersonIcon() string  { return pick(" ", "") }
func CommentIcon() string { return pick(" ", "") }
func CommitIcon() string  { return pick(" ", "") }
