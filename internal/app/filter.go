package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

type filterGroup int

const (
	groupKind filterGroup = iota
	groupState
	groupAssignee
	groupLabel
	groupMilestone
)

type filterRow struct {
	grupo  filterGroup
	rotulo string
	valor  string
}

func (m *Model) linhasFiltro() []filterRow {
	linhas := []filterRow{
		{groupKind, "issues e PRs", ""},
		{groupKind, "apenas issues", "issue"},
		{groupKind, "apenas PRs", "pr"},
		{groupState, "abertos", "open"},
		{groupState, "fechados", "closed"},
		{groupState, "todos", "all"},
		{groupAssignee, "atribuídas a mim", "@me"},
	}
	for _, l := range m.knownLabels() {
		linhas = append(linhas, filterRow{groupLabel, l, l})
	}
	for _, ms := range m.knownMilestones() {
		linhas = append(linhas, filterRow{groupMilestone, ms, ms})
	}
	return linhas
}

func (m *Model) filterChecked(l filterRow) bool {
	switch l.grupo {
	case groupKind:
		return m.filters.Kind == l.valor
	case groupState:
		return m.filters.State == l.valor
	case groupAssignee:
		return m.filters.Assignee == "@me"
	case groupLabel:
		return containsStr(m.filters.Labels, l.valor)
	case groupMilestone:
		return m.filters.Milestone == l.valor
	}
	return false
}

func (m *Model) toggleFilter() {
	linhas := m.linhasFiltro()
	if m.filterCursor >= len(linhas) {
		return
	}
	l := linhas[m.filterCursor]
	switch l.grupo {
	case groupKind:
		m.filters.Kind = l.valor
	case groupState:
		m.filters.State = l.valor
	case groupAssignee:
		if m.filters.Assignee == "@me" {
			m.filters.Assignee = ""
		} else {
			m.filters.Assignee = "@me"
		}
	case groupLabel:
		if containsStr(m.filters.Labels, l.valor) {
			m.filters.Labels = removeStr(m.filters.Labels, l.valor)
		} else {
			m.filters.Labels = append(m.filters.Labels, l.valor)
		}
	case groupMilestone:
		if m.filters.Milestone == l.valor {
			m.filters.Milestone = ""
		} else {
			m.filters.Milestone = l.valor
		}
	}
}

func (m Model) filterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	total := len(m.linhasFiltro())
	switch msg.String() {
	case "esc", "f":
		m.mode = modeTable
		return m, nil
	case "j", "down":
		if m.filterCursor < total-1 {
			m.filterCursor++
		}
		return m, nil
	case "k", "up":
		if m.filterCursor > 0 {
			m.filterCursor--
		}
		return m, nil
	case " ":
		m.toggleFilter()
		return m, nil
	case "enter":
		m.mode = modeTable
		m.triage.loaded = false
		m.prs.loaded = false
		m.loading = true
		return m, tea.Batch(m.spin.Tick, m.loadCmd())
	}
	return m, nil
}

func (m Model) viewFilter() string {
	linhas := m.linhasFiltro()
	var corpo []string
	grupoAnterior := filterGroup(-1)
	nomeGrupo := map[filterGroup]string{
		groupKind: "tipo", groupState: "estado", groupAssignee: "assignee",
		groupLabel: "labels", groupMilestone: "milestone",
	}
	for i, l := range linhas {
		if l.grupo != grupoAnterior {
			if i > 0 {
				corpo = append(corpo, "")
			}
			corpo = append(corpo, ui.StyleMuted.Render(nomeGrupo[l.grupo]))
			grupoAnterior = l.grupo
		}
		marca := ui.StyleMuted.Render(" ▢ ")
		if m.filterChecked(l) {
			marca = ui.StyleTitle.Render(" ▣ ")
		}
		text := ui.StyleText.Render(l.rotulo)
		if i == m.filterCursor {
			text = ui.StyleText.Bold(true).Render("‣ " + l.rotulo)
		} else {
			text = "  " + text
		}
		corpo = append(corpo, marca+text)
	}

	content := ui.StyleTitle.Bold(true).Render("filtros") + "\n" +
		strings.Join(corpo, "\n") + "\n\n" +
		ui.StyleKey.Render("espaço") + ui.StyleMuted.Render(" alterna  ") +
		ui.StyleKey.Render("Enter") + ui.StyleMuted.Render(" aplica  ") +
		ui.StyleKey.Render("Esc") + ui.StyleMuted.Render(" fecha")

	larg := 46
	if larg > m.width-2 {
		larg = m.width - 2
	}
	return ui.StyleBorderActive.Width(larg).Padding(0, 1).Render(content)
}

func containsStr(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func removeStr(xs []string, v string) []string {
	out := xs[:0]
	for _, x := range xs {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}
