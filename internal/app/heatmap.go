package app

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

var heatColumns = []struct {
	code  string
	class int
}{
	{"rev", 0}, {"chg", 1}, {"ci", 2}, {"mrg", 3}, {"atr", 4}, {"men", 5}, {"agd", 6},
}

const (
	heatLabelW = 22
	heatCellW  = 4
)

func heatColumnIndex(class int) int {
	for i, c := range heatColumns {
		if c.class == class {
			return i
		}
	}
	return -1
}

func (m Model) radarHeatmap(w, h int) string {
	if len(m.radar.items) == 0 {
		return m.emptyTable(w, h)
	}

	counts := map[string][]int{}
	totals := map[string]int{}
	maxCell := 1
	for _, it := range m.radar.items {
		col := heatColumnIndex(it.Class)
		if col < 0 {
			continue
		}
		repo := it.Repo
		if repo == "" {
			repo = "—"
		}
		if counts[repo] == nil {
			counts[repo] = make([]int, len(heatColumns))
		}
		counts[repo][col]++
		totals[repo]++
		if counts[repo][col] > maxCell {
			maxCell = counts[repo][col]
		}
	}

	repos := make([]string, 0, len(counts))
	for r := range counts {
		repos = append(repos, r)
	}
	sort.Slice(repos, func(a, b int) bool {
		if totals[repos[a]] != totals[repos[b]] {
			return totals[repos[a]] > totals[repos[b]]
		}
		return repos[a] < repos[b]
	})

	ramp := ui.Gradient("#3A5170", "#E0A23C", "#F0616B", 4)

	var linhas []string
	linhas = append(linhas, heatHeader())

	limite := h - 3
	if limite < 1 {
		limite = 1
	}
	mostrados := repos
	restantes := 0
	if len(repos) > limite {
		mostrados = repos[:limite]
		restantes = len(repos) - limite
	}
	for _, r := range mostrados {
		linhas = append(linhas, heatRow(r, counts[r], totals[r], maxCell, ramp))
	}
	if restantes > 0 {
		linhas = append(linhas, ui.StyleMuted.Render("  … +"+itoa(restantes)+" repositórios"))
	}
	linhas = append(linhas, "", heatLegend(ramp))
	return strings.Join(linhas, "\n")
}

func heatHeader() string {
	var b strings.Builder
	b.WriteString(ui.StyleHeader.Render(ui.Pad("REPOSITÓRIO", heatLabelW)))
	for _, c := range heatColumns {
		b.WriteString(" " + ui.StyleHeader.Render(padLeft(c.code, heatCellW)))
	}
	b.WriteString("  " + ui.StyleHeader.Render("total"))
	return b.String()
}

func heatRow(repo string, cells []int, total, maxCell int, ramp []string) string {
	var b strings.Builder
	b.WriteString(ui.StyleText.Render(ui.Pad(ui.Truncate(repo, heatLabelW), heatLabelW)))
	for _, c := range cells {
		b.WriteString(" " + heatCell(c, maxCell, ramp))
	}
	b.WriteString("  " + ui.StyleNumber.Render(padLeft(itoa(total), 3)))
	return b.String()
}

func heatCell(count, maxCell int, ramp []string) string {
	if count == 0 {
		return ui.StyleMuted.Render(padLeft("·", heatCellW))
	}
	glyphs := []string{"░", "▒", "▓", "█"}
	lvl := (count*len(glyphs) - 1) / maxCell
	if lvl < 0 {
		lvl = 0
	}
	if lvl >= len(glyphs) {
		lvl = len(glyphs) - 1
	}
	cor := lipgloss.Color(ramp[lvl])
	texto := glyphs[lvl] + " " + padLeft(itoa(count), 2)
	return lipgloss.NewStyle().Foreground(cor).Render(texto)
}

func heatLegend(ramp []string) string {
	escala := ""
	glyphs := []string{"░", "▒", "▓", "█"}
	for i, g := range glyphs {
		escala += lipgloss.NewStyle().Foreground(lipgloss.Color(ramp[i])).Render(g)
	}
	legenda := ui.StyleMuted.Render("menos ") + escala + ui.StyleMuted.Render(" mais    ") +
		ui.StyleMuted.Render("rev review · chg mudanças · ci CI · mrg merge · atr atribuída · men menção · agd aguardando")
	return legenda
}
