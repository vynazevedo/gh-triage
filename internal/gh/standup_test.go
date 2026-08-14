package gh

import "testing"

func TestBlockedFrom(t *testing.T) {
	items := []Item{
		{Number: 1, ReviewDecision: "APPROVED"},
		{Number: 2, ReviewDecision: "CHANGES_REQUESTED"},
		{Number: 3, CI: CIFail},
		{Number: 4, CI: CISuccess},
		{Number: 5},
	}
	got := BlockedFrom(items)
	if len(got) != 2 {
		t.Fatalf("bloqueados=%d, quer 2", len(got))
	}
	if got[0].Number != 2 || got[1].Number != 3 {
		t.Errorf("bloqueados errados: %d, %d", got[0].Number, got[1].Number)
	}
}
