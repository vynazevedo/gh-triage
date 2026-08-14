package app

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/vynazevedo/gh-triage/internal/gh"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

func (m Model) previewContent() string {
	larg := m.preview.Width - 1
	if larg < 20 {
		larg = 20
	}
	if m.tab == TabCommits {
		if c, ok := m.commits.current(); ok {
			return m.commitPreview(c.Commit, larg)
		}
		return ""
	}
	if it, ok := m.currentItem(); ok {
		body := m.itemPreview(it, larg)
		if m.tab == TabRadar {
			if row, ok := m.radar.current(); ok && row.Reason != "" {
				ri, rc := ui.ReasonIcon(row.Class)
				head := lipgloss.NewStyle().Foreground(rc).Bold(true).Render(ri + " " + row.Reason)
				body = head + "\n" + body
			}
		}
		return body
	}
	return ""
}

func (m Model) itemPreview(it gh.Item, larg int) string {
	icone, cor := ui.ItemIcon(it)
	cabecalho := lipgloss.NewStyle().Foreground(cor).Render(icone) + " " +
		ui.StyleNumber.Render("#"+itoa(it.Number)) + " " +
		ui.StyleText.Bold(true).Render(it.Title)

	meta := ui.StyleMuted.Render("por " + orUnknown(it.Author) +
		" · " + itoa(it.Comments) + " comentários")
	if it.Milestone != "" {
		meta += ui.StyleMuted.Render(" · milestone: " + it.Milestone)
	}
	if len(it.Assignees) > 0 {
		meta += lipgloss.NewStyle().Foreground(ui.Assignee).Render(" · " + strings.Join(it.Assignees, ", "))
	}

	linhas := []string{cabecalho, meta, urgencyLine(it)}

	if it.Kind == gh.KindPR {
		var state []string
		if txt, cor := ciText(it.CI); txt != "" {
			state = append(state, lipgloss.NewStyle().Foreground(cor).Render(txt))
		}
		if it.ReviewDecision != "" {
			state = append(state, ui.StyleMuted.Render("review: "+translateReview(it.ReviewDecision)))
		}
		if len(state) > 0 {
			linhas = append(linhas, strings.Join(state, ui.StyleMuted.Render(" · ")))
		}
	}

	for _, v := range it.Linked {
		linhas = append(linhas, ui.StyleTitle.Render("↳ vinculada: PR "+v.Reference()+" ("+situation(v)+")"))
	}

	linhas = append(linhas, "")
	if strings.TrimSpace(it.Body) == "" {
		linhas = append(linhas, ui.StyleMuted.Render("(sem descrição)"))
	} else {
		linhas = append(linhas, ui.Markdown(it.Body, larg))
	}
	return strings.Join(linhas, "\n")
}

func (m Model) commitPreview(c gh.Commit, larg int) string {
	cabecalho := ui.StyleNumber.Render(c.ShortOID()) + " " + ui.StyleText.Bold(true).Render(c.Title)
	meta := ui.StyleMuted.Render("por " + orUnknown(c.Author) +
		" · +" + itoa(c.Additions) + " -" + itoa(c.Deletions))
	linhas := []string{cabecalho, meta}
	for _, pr := range c.PRs {
		linhas = append(linhas, lipgloss.NewStyle().Foreground(ui.Merged).Render("↳ entrou pelo PR #"+itoa(pr)))
	}
	for _, is := range c.CitedIssues {
		if containsInt(c.PRs, is) {
			continue
		}
		linhas = append(linhas, lipgloss.NewStyle().Foreground(ui.Success).Render("↳ cita a issue #"+itoa(is)))
	}
	linhas = append(linhas, "")
	if strings.TrimSpace(c.Body) == "" {
		linhas = append(linhas, ui.StyleMuted.Render("(sem corpo de mensagem)"))
	} else {
		linhas = append(linhas, ui.Markdown(c.Body, larg))
	}
	return strings.Join(linhas, "\n")
}

func (m Model) detailContent() string {
	larg := m.width - 4
	if larg < 20 {
		larg = 20
	}
	det, ok := m.details[m.openedDetail.key()]
	if !ok {
		if m.loading {
			return ui.StyleMuted.Render("carregando comentários…")
		}
		return ui.StyleMuted.Render("sem detalhe carregado")
	}

	linhas := []string{
		ui.StyleNumber.Bold(true).Render("#"+itoa(det.Number)) + " " + ui.StyleText.Bold(true).Render(det.Title),
		ui.StyleMuted.Render("por " + orUnknown(det.Author)),
		"",
	}
	if strings.TrimSpace(det.Body) == "" {
		linhas = append(linhas, ui.StyleMuted.Render("(sem descrição)"))
	} else {
		linhas = append(linhas, ui.Markdown(det.Body, larg))
	}

	if len(det.Reviews) > 0 {
		linhas = append(linhas, "", separator("reviews"))
		for _, r := range det.Reviews {
			linhas = append(linhas, m.reviewBlock(r, larg))
		}
	}

	linhas = append(linhas, "", separator(commentsLabel(det)))
	if len(det.Comments) == 0 {
		linhas = append(linhas, ui.StyleMuted.Render("(nenhum comentário)"))
	} else {
		for _, c := range det.Comments {
			linhas = append(linhas, m.commentBlock(c, larg))
		}
	}
	return strings.Join(linhas, "\n")
}

func (m Model) commentBlock(c gh.Comment, larg int) string {
	cab := lipgloss.NewStyle().Foreground(ui.Assignee).Bold(true).Render("  "+orUnknown(c.Author)) +
		ui.StyleMuted.Render("  "+relativeAge(c.Date))
	corpo := indent(ui.Markdown(orEmpty(c.Body), larg-2))
	return cab + "\n" + corpo + "\n"
}

func (m Model) reviewBlock(r gh.Review, larg int) string {
	rotulo, cor := reviewLabel(r.State)
	cab := lipgloss.NewStyle().Foreground(ui.Assignee).Bold(true).Render("  "+orUnknown(r.Author)) +
		lipgloss.NewStyle().Foreground(cor).Render(" "+rotulo) +
		ui.StyleMuted.Render("  "+relativeAge(r.Date))
	if strings.TrimSpace(r.Body) == "" {
		return cab + "\n"
	}
	return cab + "\n" + indent(ui.Markdown(r.Body, larg-2)) + "\n"
}

func separator(rotulo string) string {
	return lipgloss.NewStyle().Foreground(ui.Border).Render("── ") +
		ui.StyleTitle.Render(rotulo)
}

func commentsLabel(det gh.Detail) string {
	if det.TotalComments > len(det.Comments) {
		return "comentários (" + itoa(len(det.Comments)) + " de " + itoa(det.TotalComments) + ")"
	}
	return "comentários (" + itoa(det.TotalComments) + ")"
}

func reviewLabel(state string) (string, lipgloss.AdaptiveColor) {
	switch state {
	case "APPROVED":
		return "aprovou", ui.Success
	case "CHANGES_REQUESTED":
		return "pediu mudanças", ui.Danger
	case "DISMISSED":
		return "review descartado", ui.Muted
	default:
		return "comentou", ui.Warn
	}
}

func indent(s string) string {
	linhas := strings.Split(s, "\n")
	for i := range linhas {
		linhas[i] = "  " + linhas[i]
	}
	return strings.Join(linhas, "\n")
}

func situation(v gh.LinkedPR) string {
	if v.Draft {
		return "rascunho"
	}
	return v.State.String()
}

func ciText(ci gh.CIStatus) (string, lipgloss.AdaptiveColor) {
	switch ci {
	case gh.CISuccess:
		return "CI: passou", ui.Success
	case gh.CIFail:
		return "CI: falhou", ui.Danger
	case gh.CIPending:
		return "CI: rodando", ui.Warn
	default:
		return "", ui.Muted
	}
}

func translateReview(d string) string {
	switch d {
	case "APPROVED":
		return "aprovado"
	case "CHANGES_REQUESTED":
		return "mudanças pedidas"
	case "REVIEW_REQUIRED":
		return "pendente"
	default:
		return d
	}
}

func orUnknown(s string) string {
	if s == "" {
		return "desconhecido"
	}
	return s
}

func orEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(vazio)"
	}
	return s
}

func stalenessColor(iso string) (lipgloss.AdaptiveColor, bool) {
	d, ok := gh.AgeSince(iso)
	if !ok {
		return ui.Muted, false
	}
	switch {
	case d > 7*24*time.Hour:
		return ui.Danger, true
	case d > 3*24*time.Hour:
		return ui.Warn, true
	default:
		return ui.Muted, false
	}
}

func agedWhen(iso string) string {
	cor, _ := stalenessColor(iso)
	return lipgloss.NewStyle().Foreground(cor).Render(relativeAge(iso))
}

func urgencyLine(it gh.Item) string {
	cor, forte := stalenessColor(it.UpdatedAt)
	linha := lipgloss.NewStyle().Foreground(cor).Bold(forte).Render("há " + relativeAge(it.UpdatedAt))
	if hint := gh.UrgencyHint(it); hint != "" {
		linha += ui.StyleMuted.Render(" · ") + lipgloss.NewStyle().Foreground(cor).Render(hint)
	}
	return linha
}

func relativeAge(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return "?"
	}
	min := int(time.Since(t).Minutes())
	if min < 0 {
		min = 0
	}
	switch {
	case min < 60:
		return itoa(min) + "m"
	case min < 1440:
		return itoa(min/60) + "h"
	case min < 10080:
		return itoa(min/1440) + "d"
	case min < 86400:
		return itoa(min/10080) + "sem"
	case min < 525600:
		return itoa(min/43200) + "mes"
	default:
		return itoa(min/525600) + "a"
	}
}
