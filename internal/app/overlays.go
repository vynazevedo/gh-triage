package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vynazevedo/gh-triage/internal/gh"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

func (m Model) viewDetail() string {
	cabecalho := m.headerBar(m.width)
	rodape := ui.StyleMuted.Render(" ") +
		ui.StyleKey.Render("j k") + ui.StyleMuted.Render(" rola ") +
		ui.StyleKey.Render("n") + ui.StyleMuted.Render(" comenta ") +
		ui.StyleKey.Render("o") + ui.StyleMuted.Render(" browser ") +
		ui.StyleKey.Render("r") + ui.StyleMuted.Render(" recarrega ") +
		ui.StyleKey.Render("Esc") + ui.StyleMuted.Render(" volta ")

	corpo := ui.StyleBorderActive.Width(m.width - 2).Render(m.detail.View())
	return lipgloss.JoinVertical(lipgloss.Left, cabecalho, corpo, rodape)
}

func (m Model) viewModal() string {
	larg := 60
	if larg > m.width-4 {
		larg = m.width - 4
	}
	var corpo []string
	if m.modal.kind == modalConfirm {
		for i, r := range m.modal.summary {
			if i == 0 {
				corpo = append(corpo, ui.StyleText.Bold(true).Render(r))
			} else {
				corpo = append(corpo, ui.StyleMuted.Render(r))
			}
		}
		if len(m.modal.mutations) > 1 {
			corpo = append(corpo, ui.StyleMuted.Render("erros parciais são reportados por item"))
		}
		corpo = append(corpo, "", ui.StyleKey.Render("y")+ui.StyleMuted.Render(" sim  ")+
			ui.StyleKey.Render("n")+ui.StyleMuted.Render(" não"))
	} else {
		for i, o := range m.modal.options {
			prefixo := "  "
			est := ui.StyleText
			if i == m.modal.cursor {
				prefixo = ui.StyleTitle.Render("▸ ")
				est = ui.StyleText.Bold(true)
			}
			corpo = append(corpo, prefixo+est.Render(o))
		}
		corpo = append(corpo, "", ui.StyleKey.Render("Enter")+ui.StyleMuted.Render(" escolhe  ")+
			ui.StyleKey.Render("Esc")+ui.StyleMuted.Render(" cancela"))
	}

	return modalBox(m.modal.title, corpo, larg)
}

func (m Model) viewModalText() string {
	larg := 60
	if larg > m.width-4 {
		larg = m.width - 4
	}
	title := map[textField]string{
		fieldAssignee:   "assignee",
		fieldNewIssue:   "nova issue (título)",
		fieldSearchName: "salvar busca como",
		fieldRepo:       "repositório (owner/repo) ou org (owner)",
	}[m.activeField]
	corpo := []string{
		m.input.View(),
		"",
		ui.StyleKey.Render("Enter") + ui.StyleMuted.Render(" confirma  ") +
			ui.StyleKey.Render("Esc") + ui.StyleMuted.Render(" cancela"),
	}
	return modalBox(title, corpo, larg)
}

func modalBox(title string, corpo []string, larg int) string {
	content := ui.StyleTitle.Render(title) + "\n\n" + strings.Join(corpo, "\n")
	return ui.StyleBorderActive.Width(larg).Padding(0, 1).Render(content)
}

func (m Model) viewHelp() string {
	secoes := [][2]string{
		{"NAVEGAÇÃO", ""},
		{"Tab / Shift+Tab", "troca de aba"},
		{"1 2 3 4", "Radar, Triagem, PRs, Commits"},
		{"ctrl+d / ctrl+u", "meia página"},
		{"j k / setas", "move o cursor"},
		{"g / G", "topo / fim"},
		{"Enter", "abre a view completa (comentários, reviews)"},
		{"/", "busca incremental"},
		{"f", "painel de filtros"},
		{"", ""},
		{"SELEÇÃO E BUSCAS", ""},
		{"espaço", "marca / desmarca"},
		{"A", "marca todos / limpa"},
		{"S", "buscas salvas"},
		{"w", "salvar a busca atual"},
		{"", ""},
		{"ISSUES E PRs", ""},
		{"L", "aplica label"},
		{"a", "atribui assignee"},
		{"x / X", "fecha / reabre"},
		{"n", "comenta ($EDITOR)"},
		{"c / e", "cria / edita issue ($EDITOR)"},
		{"", ""},
		{"PULL REQUESTS", ""},
		{"v", "review: aprovar, comentar, pedir mudanças"},
		{"M", "merge: merge, squash, rebase"},
		{"C", "checkout local (gh pr checkout)"},
		{"", ""},
		{"GERAL", ""},
		{".", "picker de org/repo (lista das suas orgs)"},
		{">", "trocar por digitação (owner ou owner/repo)"},
		{"o", "abre no browser"},
		{"r", "recarrega"},
		{"q / Esc", "volta ou sai"},
	}

	var linhas []string
	for _, s := range secoes {
		switch {
		case s[0] == "" && s[1] == "":
			linhas = append(linhas, "")
		case s[1] == "":
			linhas = append(linhas, ui.StyleTitle.Render(s[0]))
		default:
			linhas = append(linhas, "  "+ui.StyleKey.Render(ui.Pad(s[0], 18))+ui.StyleText.Render(s[1]))
		}
	}
	content := ui.StyleTitle.Bold(true).Render("atalhos") + "\n\n" +
		strings.Join(linhas, "\n") + "\n\n" +
		ui.StyleMuted.Render("qualquer tecla fecha")

	larg := 62
	if larg > m.width-2 {
		larg = m.width - 2
	}
	return ui.StyleBorderActive.Width(larg).Padding(0, 1).Render(content)
}

func (m Model) viewSavedSearches() string {
	var linhas []string
	if len(m.savedSearches) == 0 {
		linhas = append(linhas, ui.StyleMuted.Render("nenhuma busca salva ainda"))
		linhas = append(linhas, ui.StyleMuted.Render("use w na tabela para salvar a busca atual"))
	} else {
		for i, b := range m.savedSearches {
			prefixo := "  "
			est := ui.StyleText
			if i == m.savedCursor {
				prefixo = ui.StyleTitle.Render("▸ ")
				est = ui.StyleText.Bold(true)
			}
			summary := ui.StyleMuted.Render("  " + filtersSummary(b.Filters))
			linhas = append(linhas, prefixo+est.Render(b.Name)+summary)
		}
	}
	content := ui.StyleTitle.Bold(true).Render("buscas salvas") + "\n\n" +
		strings.Join(linhas, "\n") + "\n\n" +
		ui.StyleKey.Render("Enter") + ui.StyleMuted.Render(" aplica  ") +
		ui.StyleKey.Render("d") + ui.StyleMuted.Render(" remove  ") +
		ui.StyleKey.Render("Esc") + ui.StyleMuted.Render(" volta")

	larg := 64
	if larg > m.width-2 {
		larg = m.width - 2
	}
	return ui.StyleBorderActive.Width(larg).Padding(0, 1).Render(content)
}

func filtersSummary(f gh.Filters) string {
	var p []string
	if f.State != "" && f.State != "open" {
		p = append(p, f.State)
	}
	if f.Kind != "" {
		p = append(p, f.Kind)
	}
	for _, l := range f.Labels {
		p = append(p, "label:"+l)
	}
	if f.Assignee != "" {
		p = append(p, "@"+f.Assignee)
	}
	if f.Term != "" {
		p = append(p, "\""+f.Term+"\"")
	}
	if len(p) == 0 {
		return "abertos"
	}
	return strings.Join(p, " ")
}
