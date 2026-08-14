package gh

import "testing"

func TestCitedNumbers(t *testing.T) {
	got := CitedNumbers("corrige #13 e também #4567, e #13 de novo")
	if len(got) != 2 || got[0] != 13 || got[1] != 4567 {
		t.Fatalf("got %v, quer [13 4567] sem duplicar", got)
	}
	if len(CitedNumbers("sem referência")) != 0 {
		t.Fatal("texto sem # deveria dar vazio")
	}
}

func TestMutationDescription(t *testing.T) {
	casos := []struct {
		m    Mutation
		quer string
	}{
		{Mutation{Kind: MutLabel, Labels: []string{"bug", "p1"}}, "aplicar label bug, p1"},
		{Mutation{Kind: MutMerge, Method: "squash"}, "merge (squash)"},
		{Mutation{Kind: MutAssignee}, "limpar assignees"},
		{Mutation{Kind: MutReview, Event: "APPROVE"}, "aprovar"},
	}
	for _, c := range casos {
		if got := c.m.Description(); got != c.quer {
			t.Errorf("Descricao()=%q, quer %q", got, c.quer)
		}
	}
}

func TestDestructive(t *testing.T) {
	if !(Mutation{Kind: MutClose}).Destructive() {
		t.Error("fechar deveria ser destrutiva")
	}
	if !(Mutation{Kind: MutMerge}).Destructive() {
		t.Error("merge deveria ser destrutiva")
	}
	if (Mutation{Kind: MutLabel}).Destructive() {
		t.Error("label não é destrutiva")
	}
}

func TestShortOID(t *testing.T) {
	if got := (Commit{OID: "a1b2c3d4e5f6"}).ShortOID(); got != "a1b2c3d" {
		t.Errorf("OIDCurto=%q, quer a1b2c3d", got)
	}
	if got := (Commit{OID: "abc"}).ShortOID(); got != "abc" {
		t.Errorf("OIDCurto curto=%q", got)
	}
}
