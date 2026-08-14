package app

import (
	"testing"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func filters(state string) gh.Filters {
	return gh.Filters{Repo: "cli/cli", State: state}
}

func groupPos(m Model, g filterGroup, valor string) int {
	for i, l := range m.linhasFiltro() {
		if l.grupo == g && l.valor == valor {
			return i
		}
	}
	return -1
}

func TestToggleKindRadio(t *testing.T) {
	m := New(nil, "cli/cli", filters("open"), TabTriage)
	m.filterCursor = groupPos(m, groupKind, "pr")
	m.toggleFilter()

	if m.filters.Kind != "pr" {
		t.Fatalf("tipo=%q, quer pr", m.filters.Kind)
	}
	if !m.filterChecked(filterRow{grupo: groupKind, valor: "pr"}) {
		t.Error("pr deveria estar marcado")
	}
	if m.filterChecked(filterRow{grupo: groupKind, valor: ""}) {
		t.Error("ambos não deveria estar marcado após escolher pr")
	}
}

func TestAssigneeToggle(t *testing.T) {
	m := New(nil, "cli/cli", filters("open"), TabTriage)
	m.filterCursor = groupPos(m, groupAssignee, "@me")

	m.toggleFilter()
	if m.filters.Assignee != "@me" {
		t.Fatalf("assignee=%q, quer @me", m.filters.Assignee)
	}
	m.toggleFilter()
	if m.filters.Assignee != "" {
		t.Fatalf("segundo toggle deveria limpar, veio %q", m.filters.Assignee)
	}
}

func TestLabelMultiToggle(t *testing.T) {
	m := New(nil, "cli/cli", filters("open"), TabTriage)
	m.triage.replace([]itemRow{
		{Item: gh.Item{Number: 1, Title: "x", Labels: []gh.Label{{Name: "bug"}, {Name: "p1"}}}},
	}, 1, "", "")

	p := groupPos(m, groupLabel, "bug")
	if p < 0 {
		t.Fatal("label bug deveria aparecer no painel a partir dos itens carregados")
	}
	m.filterCursor = p
	m.toggleFilter()
	if !containsStr(m.filters.Labels, "bug") {
		t.Fatalf("labels=%v, quer conter bug", m.filters.Labels)
	}
	m.toggleFilter()
	if containsStr(m.filters.Labels, "bug") {
		t.Error("segundo toggle deveria remover a label")
	}
}
