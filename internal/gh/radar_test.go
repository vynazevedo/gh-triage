package gh

import "testing"

func TestClassifyRadar(t *testing.T) {
	casos := []struct {
		nome   string
		origem RadarSource
		it     Item
		classe int
	}{
		{"review pedido", SourceReviewRequested, Item{}, 0},
		{"mudanças pedidas", SourceMyPR, Item{ReviewDecision: "CHANGES_REQUESTED"}, 1},
		{"ci falhou", SourceMyPR, Item{CI: CIFail}, 2},
		{"pronto para merge", SourceMyPR, Item{ReviewDecision: "APPROVED"}, 3},
		{"atribuída", SourceAssigned, Item{}, 4},
		{"menção", SourceMention, Item{}, 5},
		{"aguardando", SourceMyPR, Item{}, 6},
		{"rascunho", SourceMyPR, Item{Draft: true}, 7},
	}
	for _, c := range casos {
		classe, motivo := ClassifyRadar(c.origem, c.it)
		if classe != c.classe {
			t.Errorf("%s: classe=%d, quer %d", c.nome, classe, c.classe)
		}
		if motivo == "" {
			t.Errorf("%s: motivo vazio", c.nome)
		}
	}
}

func TestMergeRadarDedupAndOrder(t *testing.T) {
	pr := Item{Number: 10, Repo: "a/x", Kind: KindPR, UpdatedAt: "2026-07-20T10:00:00Z"}
	issueVelha := Item{Number: 5, Repo: "a/x", Kind: KindIssue, UpdatedAt: "2026-07-01T10:00:00Z"}
	issueNova := Item{Number: 6, Repo: "b/y", Kind: KindIssue, UpdatedAt: "2026-07-22T10:00:00Z"}

	out := MergeRadar([]SourceResult{
		{SourceMention, []Item{pr, issueVelha, issueNova}},
		{SourceReviewRequested, []Item{pr}},
		{SourceAssigned, []Item{issueVelha}},
	})

	if len(out) != 3 {
		t.Fatalf("len=%d, quer 3 (dedupe)", len(out))
	}
	if out[0].Number != 10 || out[0].Class != 0 {
		t.Errorf("primeiro deveria ser o PR com review pedido (classe 0), veio #%d classe %d", out[0].Number, out[0].Class)
	}
	if out[1].Number != 5 || out[1].Class != 4 {
		t.Errorf("segundo deveria ser a issue atribuída (classe 4), veio #%d classe %d", out[1].Number, out[1].Class)
	}
	if out[2].Class != 5 {
		t.Errorf("terceiro deveria ser a menção, veio classe %d", out[2].Class)
	}
}

func TestMergeRadarOldestFirstWithinClass(t *testing.T) {
	nova := Item{Number: 1, Repo: "a/x", Kind: KindIssue, UpdatedAt: "2026-07-22T10:00:00Z"}
	velha := Item{Number: 2, Repo: "a/x", Kind: KindIssue, UpdatedAt: "2026-07-01T10:00:00Z"}

	out := MergeRadar([]SourceResult{
		{SourceAssigned, []Item{nova, velha}},
	})

	if len(out) != 2 {
		t.Fatalf("len=%d, quer 2", len(out))
	}
	if out[0].Number != 2 {
		t.Errorf("a que espera há mais tempo (#2) deveria vir primeiro, veio #%d", out[0].Number)
	}
}

func TestKeyNoCollisionAcrossRepos(t *testing.T) {
	a := Item{Number: 7, Repo: "org/a", Kind: KindIssue}
	b := Item{Number: 7, Repo: "org/b", Kind: KindIssue}
	if a.Key() == b.Key() {
		t.Fatalf("chaves iguais para repos distintos: %q", a.Key())
	}
	semRepo := Item{Number: 7, Kind: KindIssue}
	if semRepo.Key() != "issue#7" {
		t.Fatalf("chave sem repo=%q", semRepo.Key())
	}
}
