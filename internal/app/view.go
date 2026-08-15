package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

func (m Model) View() string {
	if m.width == 0 {
		return "iniciando…"
	}
	if m.mode == modeHelp {
		return m.viewHelp()
	}
	if m.mode == modeDetail {
		return m.viewDetail()
	}
	if m.mode == modeSavedSearches {
		return m.viewSavedSearches()
	}
	if m.mode == modePicker {
		return m.viewPicker()
	}

	header := m.headerBar(m.width)
	body := m.viewBody()
	footer := m.viewStatus()

	base := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	if m.mode == modeModal && m.modal != nil {
		return m.overlay(base, m.viewModal())
	}
	if m.mode == modeModalText {
		return m.overlay(base, m.viewModalText())
	}
	if m.mode == modeFilter {
		return m.overlay(base, m.viewFilter())
	}
	if m.mode == modeQuitConfirm {
		return m.overlay(base, m.viewQuitConfirm())
	}
	return base
}

func (m Model) viewQuitConfirm() string {
	corpo := []string{
		ui.StyleText.Bold(true).Render("Sair do gh-triage?"),
		"",
		ui.StyleKey.Render("Esc") + ui.StyleMuted.Render(" ou ") + ui.StyleKey.Render("q") +
			ui.StyleMuted.Render(" confirma  ") +
			ui.StyleKey.Render("qualquer tecla") + ui.StyleMuted.Render(" cancela"),
	}
	larg := 42
	if larg > m.width-4 {
		larg = m.width - 4
	}
	return ui.StyleBorderActive.Width(larg).Padding(0, 1).Render(strings.Join(corpo, "\n"))
}

func (m Model) viewBody() string {
	bodyH := m.bodyHeight()
	title := m.tab.title()
	if m.tab == TabRadar && m.showMap {
		title = "Radar · mapa"
	}
	if m.twoPane() {
		listW := m.listWidth()
		sideW := m.sideWidth()
		list := m.panel(title, m.tabCount(m.tab), m.tabKeys(), listW, bodyH, true,
			m.viewTable(listW-2, bodyH-2))
		side := m.panel("Preview", -1, nil, sideW, bodyH, false, m.preview.View())
		return lipgloss.JoinHorizontal(lipgloss.Top, list, side)
	}
	return m.panel(title, m.tabCount(m.tab), m.tabKeys(), m.width, bodyH, true,
		m.viewTable(m.width-2, bodyH-2))
}

// panel renders a btop-style titled box: a hand-composed top border carrying
// the title (+ count) on the left and hotkeys on the right, over a body framed
// on left/right/bottom.
func (m Model) panel(title string, count int, keys [][2]string, w, h int, active bool, body string) string {
	if w < 4 || h < 3 {
		return body
	}
	bc := ui.Border
	if active {
		bc = ui.BorderActive
	}
	line := lipgloss.NewStyle().Foreground(bc)

	left := line.Render("╭─┤ ") + ui.StyleTitle.Render(title)
	if count >= 0 {
		left += ui.StyleMuted.Render(" · ") + ui.StyleNumber.Render(itoa(count))
	}
	left += line.Render(" ├")

	right := line.Render("─╮")
	if len(keys) > 0 {
		right = line.Render("┤ ") + hotkeys(keys) + line.Render(" ├─╮")
	}

	fillN := w - ansi.StringWidth(left) - ansi.StringWidth(right)
	if fillN < 0 && len(keys) > 0 {
		right = line.Render("─╮")
		fillN = w - ansi.StringWidth(left) - ansi.StringWidth(right)
	}
	if fillN < 0 {
		fillN = 0
	}
	top := left + line.Render(strings.Repeat("─", fillN)) + right

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(bc).
		Width(w - 2).
		Height(h - 2).
		Render(body)

	return top + "\n" + box
}

func hotkeys(keys [][2]string) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, ui.StyleKey.Render(k[0])+" "+ui.StyleMuted.Render(k[1]))
	}
	return strings.Join(parts, ui.StyleMuted.Render(" · "))
}

func (m Model) headerBar(w int) string {
	parts := []string{ui.StyleTitle.Render(" " + m.repo + " ")}
	for _, a := range tabs {
		style := ui.StyleTabInactive
		mark := "  "
		if a == m.tab {
			style = ui.StyleTabActive
			mark = " ▸"
		}
		label := a.title()
		if n := m.tabCount(a); n >= 0 {
			label += " " + itoa(n)
		}
		parts = append(parts, style.Render(mark+label))
	}
	left := strings.Join(parts, " ")

	if m.tab == TabRadar && m.radar.loaded && len(m.radar.items) > 0 {
		left += "   " + m.urgencyMeter()
	} else if m.loading {
		left += ui.StyleMuted.Render("   " + m.spin.View() + " carregando")
	}

	right := hotkeys([][2]string{{"Tab", "abas"}, {"f", "filtro"}, {"S", "salvas"}, {"?", "ajuda"}})
	return m.spread(left, right)
}

// urgencyMeter is the one hero-colored element: a segmented bar of Radar items
// grouped into critical / medium / low, plus the total count.
func (m Model) urgencyMeter() string {
	var crit, med, low float64
	for _, it := range m.radar.items {
		switch {
		case it.Class <= 2:
			crit++
		case it.Class <= 4:
			med++
		default:
			low++
		}
	}
	bar := ui.SegmentedMeter(
		[]float64{crit, med, low},
		[]string{string(ui.Danger.Dark), string(ui.Warn.Dark), string(ui.Muted.Dark)},
		12,
	)
	return ui.StyleMuted.Render("urgência ") + bar + ui.StyleNumber.Render(" "+itoa(len(m.radar.items)))
}

func (m Model) spread(left, right string) string {
	lw := ansi.StringWidth(left)
	rw := ansi.StringWidth(right)
	gap := m.width - lw - rw
	if gap < 1 {
		return ui.Truncate(left, m.width)
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) viewStatus() string {
	if m.currentToast != nil {
		style := ui.StyleToastOk
		if m.currentToast.isErr {
			style = ui.StyleToastErr
		}
		return style.Render(" " + ui.Truncate(m.currentToast.text, m.width-1))
	}
	if m.mode == modeSearch {
		return ui.StyleKey.Render(" ") + m.search.View() +
			ui.StyleMuted.Render("  "+itoa(m.totalVisible())+" de "+itoa(m.totalResults()))
	}

	var b strings.Builder
	b.WriteString(ui.StyleMuted.Render(" " + m.selectionSummary() + " "))
	for _, k := range m.tabKeys() {
		b.WriteString(ui.StyleKey.Render(k[0]))
		b.WriteString(ui.StyleMuted.Render(" " + k[1] + " "))
	}
	if m.rateLimitLow() {
		b.WriteString(ui.StyleWarn.Render(" rate limit: " + itoa(m.rateLimit)))
	}
	return ui.Truncate(b.String(), m.width)
}

func (m Model) selectionSummary() string {
	n := m.totalSelected()
	switch n {
	case 0:
		return itoa(m.totalVisible()) + " itens ·"
	case 1:
		return "1 marcado ·"
	default:
		return itoa(n) + " marcados ·"
	}
}

func (m Model) tabCount(a Tab) int {
	switch a {
	case TabRadar:
		if m.radar.loaded {
			return len(m.radar.items)
		}
	case TabTriage:
		if m.triage.loaded {
			return m.triage.total
		}
	case TabPRs:
		if m.prs.loaded {
			return m.prs.total
		}
	case TabCommits:
		if m.commits.loaded {
			return len(m.commits.items)
		}
	}
	return -1
}

func (m Model) tabKeys() [][2]string {
	switch m.tab {
	case TabRadar:
		if m.showMap {
			return [][2]string{{"m", "lista"}, {"r", "refresh"}, {"?", "ajuda"}}
		}
		return [][2]string{{"enter", "detalhe"}, {"m", "mapa"}, {"o", "abrir"}, {"x", "fechar"}, {"/", "busca"}}
	case TabCommits:
		return [][2]string{{"o", "abrir"}, {"/", "busca"}, {"r", "refresh"}}
	case TabPRs:
		return [][2]string{{"v", "review"}, {"M", "merge"}, {"C", "checkout"}, {"x", "fechar"}, {"o", "abrir"}}
	default:
		return [][2]string{{"L", "label"}, {"a", "assign"}, {"x", "fechar"}, {"n", "comentar"}, {"e", "editar"}}
	}
}

func (m Model) overlay(base, panel string) string {
	baseLines := strings.Split(base, "\n")
	panelLines := strings.Split(panel, "\n")
	panelH := len(panelLines)
	panelW := lipgloss.Width(panel)

	top := (len(baseLines) - panelH) / 2
	if top < 0 {
		top = 0
	}
	left := (m.width - panelW) / 2
	if left < 0 {
		left = 0
	}

	for i, pl := range panelLines {
		idx := top + i
		if idx < 0 || idx >= len(baseLines) {
			continue
		}
		baseLines[idx] = paste(baseLines[idx], pl, left, m.width)
	}
	return strings.Join(baseLines, "\n")
}

func paste(bg, fg string, col, width int) string {
	bg = padRight(bg, width)
	pref := ansi.Truncate(bg, col, "")
	fgW := ansi.StringWidth(fg)
	suf := ansi.TruncateLeft(bg, col+fgW, "")
	return pref + "\x1b[0m" + fg + "\x1b[0m" + suf
}

func padRight(s string, width int) string {
	l := ansi.StringWidth(s)
	if l >= width {
		return s
	}
	return s + strings.Repeat(" ", width-l)
}
