package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

func renderModel(items []gh.Item) Model {
	m := New(nil, "cli/cli", filters("open"), TabTriage)
	m.width, m.height = 100, 28
	m.resize()
	linhas := make([]itemRow, len(items))
	for i, it := range items {
		linhas[i] = itemRow{Item: it}
	}
	m.triage.replace(linhas, len(items), "", "")
	m.syncPreview()
	return m
}

func stripANSI(s string) string { return ansi.Strip(s) }

func TestViewShowsTabsAndRepo(t *testing.T) {
	out := stripANSI(renderModel(nil).View())
	for _, esperado := range []string{"cli/cli", "Triagem", "PRs", "Commits"} {
		if !strings.Contains(out, esperado) {
			t.Errorf("View deveria conter %q", esperado)
		}
	}
}

func TestViewEmptyTable(t *testing.T) {
	out := stripANSI(renderModel(nil).View())
	if !strings.Contains(out, "nenhum resultado") {
		t.Error("tabela vazia deveria mostrar aviso")
	}
}

func TestViewPopulated(t *testing.T) {
	items := []gh.Item{
		{Number: 9214, Kind: gh.KindIssue, State: gh.StateOpen, Title: "trava no fork",
			Labels: []gh.Label{{Name: "bug", Color: "d73a4a"}}, Author: "vilmib",
			UpdatedAt: "2026-07-20T12:00:00Z"},
		{Number: 9201, Kind: gh.KindPR, State: gh.StateOpen, Title: "fix json",
			CI: gh.CISuccess, UpdatedAt: "2026-07-19T12:00:00Z"},
	}
	out := stripANSI(renderModel(items).View())
	for _, esperado := range []string{"9214", "trava no fork", "bug", "9201", "fix json", "2 itens"} {
		if !strings.Contains(out, esperado) {
			t.Errorf("View populada deveria conter %q\n---\n%s", esperado, out)
		}
	}
}

func TestViewNoANSILeak(t *testing.T) {
	items := []gh.Item{{Number: 1, Title: "x", Labels: []gh.Label{{Name: "bug", Color: "d73a4a"}}}}
	m := renderModel(items)
	m.mode = modeFilter
	saida := m.View()
	limpo := stripANSI(saida)
	if strings.Contains(limpo, "[38;") || strings.Contains(limpo, "[0m") || strings.Contains(limpo, "[1;") {
		t.Error("overlay do filtro não deveria vazar sequências ANSI cruas")
	}
	if !strings.Contains(limpo, "filtros") {
		t.Error("modo filtro deveria renderizar o painel")
	}
}

func TestViewPreviewWithCrossref(t *testing.T) {
	items := []gh.Item{{
		Number: 9214, Kind: gh.KindIssue, State: gh.StateOpen, Title: "trava",
		Linked: []gh.LinkedPR{{Number: 9201, State: gh.StateOpen}},
	}}
	out := stripANSI(renderModel(items).View())
	if !strings.Contains(out, "vinculada: PR #9201") {
		t.Errorf("preview deveria mostrar o cruzamento issue->PR\n---\n%s", out)
	}
}
