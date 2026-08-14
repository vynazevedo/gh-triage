package gh

import (
	"testing"
	"time"
)

func TestUrgencyHint(t *testing.T) {
	casos := []struct {
		nome string
		it   Item
		quer string
	}{
		{"rascunho tem prioridade", Item{Kind: KindPR, Draft: true, CI: CIFail}, "rascunho"},
		{"ci vermelho", Item{Kind: KindPR, CI: CIFail}, "CI vermelho"},
		{"mudanças pedidas", Item{Kind: KindPR, ReviewDecision: "CHANGES_REQUESTED"}, "mudanças pedidas"},
		{"aprovado", Item{Kind: KindPR, ReviewDecision: "APPROVED"}, "aprovado, pronto para merge"},
		{"ci rodando", Item{Kind: KindPR, CI: CIPending}, "CI rodando"},
		{"sem review", Item{Kind: KindPR}, "sem review ainda"},
		{"issue simples", Item{Kind: KindIssue}, ""},
		{"issue com PR", Item{Kind: KindIssue, Linked: []LinkedPR{{Number: 1}}}, "tem PR vinculado"},
	}
	for _, c := range casos {
		if got := UrgencyHint(c.it); got != c.quer {
			t.Errorf("%s: UrgencyHint=%q, quer %q", c.nome, got, c.quer)
		}
	}
}

func TestAgeSince(t *testing.T) {
	if _, ok := AgeSince("não é data"); ok {
		t.Error("iso inválido não deveria parsear")
	}
	d, ok := AgeSince(time.Now().Add(-2 * time.Hour).Format(time.RFC3339))
	if !ok {
		t.Fatal("iso válido deveria parsear")
	}
	if d < time.Hour {
		t.Errorf("idade=%v, esperava ao menos 1h", d)
	}
}
