package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	if s == "esc" {
		return tea.KeyMsg{Type: tea.KeyEscape}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestQuitConfirmFlow(t *testing.T) {
	base := Model{mode: modeTable}

	afterEsc, _ := base.onKey(key("esc"))
	m1 := afterEsc.(Model)
	if m1.mode != modeQuitConfirm {
		t.Fatalf("primeiro esc deveria abrir a confirmação, modo=%d", m1.mode)
	}
	if m1.quit {
		t.Fatal("não deveria sair no primeiro esc")
	}

	afterCancel, _ := m1.onKey(key("j"))
	if afterCancel.(Model).mode != modeTable {
		t.Fatal("qualquer outra tecla deveria cancelar e voltar à tabela")
	}

	afterConfirm, _ := m1.onKey(key("esc"))
	if !afterConfirm.(Model).quit {
		t.Fatal("segundo esc deveria confirmar a saída")
	}
}

func TestMapToggleOnlyRadar(t *testing.T) {
	radar := Model{mode: modeTable, tab: TabRadar}
	after, _ := radar.onKey(key("m"))
	if !after.(Model).showMap {
		t.Fatal("m no Radar deveria ligar o mapa")
	}

	triagem := Model{mode: modeTable, tab: TabTriage}
	after2, _ := triagem.onKey(key("m"))
	if after2.(Model).showMap {
		t.Fatal("m fora do Radar não deveria ligar o mapa")
	}
}
