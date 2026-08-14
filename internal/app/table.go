package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vynazevedo/gh-triage/internal/gh"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

func (m Model) viewTable(w, h int) string {
	switch m.tab {
	case TabCommits:
		return m.commitsTable(w, h)
	case TabRadar:
		return m.radarTable(w, h)
	default:
		return m.itemsTable(w, h)
	}
}

func window(cursor, total, height int) (int, int) {
	if total <= height {
		return 0, total
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	if start+height > total {
		start = total - height
	}
	return start, start + height
}

func titleWidth(w, reserved int) int {
	tw := w - reserved
	if tw < 12 {
		tw = 12
	}
	return tw
}

func (m Model) radarTable(w, h int) string {
	titleW := titleWidth(w, 64)
	if len(m.radar.visible) == 0 {
		return m.radarHeader(w, titleW) + "\n" + m.emptyTable(w, h-1)
	}
	rows := []string{m.radarHeader(w, titleW)}
	start, end := window(m.radar.cursor, len(m.radar.visible), h-1)
	for pos := start; pos < end; pos++ {
		it := m.radar.items[m.radar.visible[pos]]
		rows = append(rows, m.renderRadarRow(it, w, titleW, pos == m.radar.cursor, m.radar.selected[it.Key()]))
	}
	return strings.Join(rows, "\n")
}

func (m Model) radarHeader(w, titleW int) string {
	return header([]string{
		ui.Pad("   #", 8),
		ui.Pad("", 1),
		ui.Pad("TÍTULO", titleW),
		ui.Pad("REPO", 20),
		ui.Pad("MOTIVO", 22),
		"QUANDO",
	}, w)
}

func (m Model) renderRadarRow(it itemRow, w, titleW int, sel, marked bool) string {
	icon, iconColor := ui.ItemIcon(it.Item)
	mark := "  "
	if marked {
		mark = ui.StyleTitle.Render("▣ ")
	}
	num := ui.StyleNumber.Render(ui.Pad(itoa(it.Number), 6))
	ic := lipgloss.NewStyle().Foreground(iconColor).Render(icon)
	title := ui.StyleText.Render(ui.Pad(ui.Truncate(it.Title, titleW), titleW))
	repo := ui.StyleMuted.Render(ui.Pad(ui.Truncate(it.Repo, 20), 20))
	reasonIcon, reasonColor := ui.ReasonIcon(it.Class)
	reason := lipgloss.NewStyle().Foreground(reasonColor).Render(ui.Pad(ui.Truncate(reasonIcon+" "+it.Reason, 22), 22))
	when := agedWhen(it.UpdatedAt)

	cols := []string{mark + num, ic, title, repo, reason, when}
	return m.renderRow(cols, w, sel)
}

func (m Model) itemsTable(w, h int) string {
	l := m.currentItemList()
	reserved := 50
	if m.tab == TabPRs {
		reserved = 58
	}
	titleW := titleWidth(w, reserved)

	if len(l.visible) == 0 {
		return m.itemsHeader(w, titleW, m.tab == TabPRs) + "\n" + m.emptyTable(w, h-1)
	}

	rows := []string{m.itemsHeader(w, titleW, m.tab == TabPRs)}
	start, end := window(l.cursor, len(l.visible), h-1)
	for pos := start; pos < end; pos++ {
		it := l.items[l.visible[pos]]
		rows = append(rows, m.renderItemRow(it.Item, w, titleW, pos == l.cursor, l.selected[it.Key()]))
	}
	return strings.Join(rows, "\n")
}

func (m Model) itemsHeader(w, titleW int, withReview bool) string {
	cols := []string{
		ui.Pad("   #", 8),
		ui.Pad("", 1),
		ui.Pad("TÍTULO", titleW),
		ui.Pad("LABELS", 16),
		ui.Pad("ASSIGNEE", 10),
	}
	if withReview {
		cols = append(cols, ui.Pad("REVIEW", 8))
	}
	cols = append(cols, ui.Pad("", 1), "QUANDO")
	return header(cols, w)
}

func (m Model) renderItemRow(it gh.Item, w, titleW int, sel, marked bool) string {
	icon, iconColor := ui.ItemIcon(it)
	mark := "  "
	if marked {
		mark = ui.StyleTitle.Render("▣ ")
	}
	num := ui.StyleNumber.Render(ui.Pad(itoa(it.Number), 6))
	ic := lipgloss.NewStyle().Foreground(iconColor).Render(icon)
	title := ui.StyleText.Render(ui.Pad(ui.Truncate(it.Title, titleW), titleW))
	labels := m.labelsCell(it, 16)
	assignee := m.assigneeCell(it, 10)

	cols := []string{mark + num, ic, title, labels, assignee}
	if m.tab == TabPRs {
		txt, color := ui.ReviewText(it.ReviewDecision)
		cols = append(cols, lipgloss.NewStyle().Foreground(color).Render(ui.Pad(txt, 8)))
	}
	ciIcon, ciColor := ui.CIIcon(it.CI)
	cols = append(cols, lipgloss.NewStyle().Foreground(ciColor).Render(ciIcon))
	cols = append(cols, agedWhen(it.UpdatedAt))

	return m.renderRow(cols, w, sel)
}

func (m Model) renderRow(cols []string, w int, sel bool) string {
	line := strings.Join(cols, " ")
	if sel {
		body := ui.StyleSelected.Width(w - 1).Render(ui.Truncate(line, w-1))
		return ui.StyleAccent.Render("▌") + body
	}
	return ui.Truncate(" "+line, w)
}

func header(cols []string, w int) string {
	return ui.Truncate(" "+ui.StyleHeader.Render(strings.Join(cols, " ")), w)
}

func (m Model) labelsCell(it gh.Item, width int) string {
	if len(it.Labels) == 0 {
		return ui.Pad("", width)
	}
	var parts []string
	used := 0
	for _, l := range it.Labels {
		cost := len([]rune(l.Name)) + 1
		if used+cost > width {
			parts = append(parts, ui.StyleMuted.Render("…"))
			break
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(ui.LabelColor(l.Color)).Render(l.Name))
		used += cost
	}
	txt := strings.Join(parts, ui.StyleMuted.Render(","))
	pad := width - lipgloss.Width(txt)
	if pad > 0 {
		txt += strings.Repeat(" ", pad)
	}
	return txt
}

func (m Model) assigneeCell(it gh.Item, width int) string {
	if len(it.Assignees) == 0 {
		return ui.Pad("", width)
	}
	txt := it.Assignees[0]
	if len(it.Assignees) > 1 {
		txt += " +" + itoa(len(it.Assignees)-1)
	}
	return lipgloss.NewStyle().Foreground(ui.Assignee).Render(ui.Pad(ui.Truncate(txt, width), width))
}

func (m Model) commitsTable(w, h int) string {
	titleW := titleWidth(w, 56)
	if len(m.commits.visible) == 0 {
		return m.commitsHeader(w, titleW) + "\n" + m.emptyTable(w, h-1)
	}

	rows := []string{m.commitsHeader(w, titleW)}
	start, end := window(m.commits.cursor, len(m.commits.visible), h-1)
	for pos := start; pos < end; pos++ {
		c := m.commits.items[m.commits.visible[pos]]
		rows = append(rows, m.renderCommitRow(c.Commit, w, titleW, pos == m.commits.cursor, m.commits.selected[c.Key()]))
	}
	return strings.Join(rows, "\n")
}

func (m Model) commitsHeader(w, titleW int) string {
	return header([]string{
		ui.Pad("   SHA", 9),
		ui.Pad("MENSAGEM", titleW),
		ui.Pad("AUTOR", 10),
		ui.Pad("VÍNCULOS", 14),
		ui.Pad("DIFF", 12),
		ui.Pad("", 1),
		"QUANDO",
	}, w)
}

func (m Model) renderCommitRow(c gh.Commit, w, titleW int, sel, marked bool) string {
	mark := "  "
	if marked {
		mark = ui.StyleTitle.Render("▣ ")
	}
	sha := ui.StyleNumber.Render(ui.Pad(c.ShortOID(), 7))
	title := ui.StyleText.Render(ui.Pad(ui.Truncate(c.Title, titleW), titleW))
	author := lipgloss.NewStyle().Foreground(ui.Assignee).Render(ui.Pad(ui.Truncate(orDash(c.Author), 10), 10))
	links := m.linksCell(c, 14)
	diff := lipgloss.NewStyle().Foreground(ui.Success).Render("+"+itoa(c.Additions)) + " " +
		lipgloss.NewStyle().Foreground(ui.Danger).Render("-"+itoa(c.Deletions))
	ciIcon, ciColor := ui.CIIcon(c.CI)
	ci := lipgloss.NewStyle().Foreground(ciColor).Render(ciIcon)

	cols := []string{mark + sha, title, author, links, ui.Pad(diff, 12), ci, ui.StyleMuted.Render(relativeAge(c.Date))}
	return m.renderRow(cols, w, sel)
}

func (m Model) linksCell(c gh.Commit, width int) string {
	var parts []string
	for _, pr := range c.PRs {
		if len(parts) >= 2 {
			break
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(ui.Merged).Render("PR#"+itoa(pr)))
	}
	for _, is := range c.CitedIssues {
		if len(parts) >= 3 || containsInt(c.PRs, is) {
			continue
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(ui.Success).Render("#"+itoa(is)))
	}
	if len(parts) == 0 {
		return ui.StyleMuted.Render(ui.Pad("·", width))
	}
	txt := strings.Join(parts, " ")
	pad := width - lipgloss.Width(txt)
	if pad > 0 {
		txt += strings.Repeat(" ", pad)
	}
	return txt
}

func (m Model) emptyTable(w, height int) string {
	txt := "nenhum resultado para os filtros atuais"
	if m.tab == TabRadar {
		txt = "tudo em dia: nada esperando por você"
	}
	if m.loading {
		txt = "buscando…"
	} else if m.search.Value() != "" {
		txt = "nada casa com a busca — Esc para limpar"
	}
	center := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Foreground(ui.Muted).Render(txt)
	rows := []string{"", center}
	for len(rows) < height {
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
}

func orDash(s string) string {
	if s == "" {
		return "·"
	}
	return s
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
