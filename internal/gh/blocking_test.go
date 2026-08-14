package gh

import "testing"

func TestOldestFirst(t *testing.T) {
	items := []Item{
		{Number: 1, UpdatedAt: "2026-08-10T10:00:00Z"},
		{Number: 2, UpdatedAt: "2026-07-01T10:00:00Z"},
		{Number: 3, UpdatedAt: "2026-08-01T10:00:00Z"},
	}
	oldestFirst(items)
	if items[0].Number != 2 || items[1].Number != 3 || items[2].Number != 1 {
		t.Fatalf("ordem=%d,%d,%d; queria 2,3,1 (mais antigo primeiro)", items[0].Number, items[1].Number, items[2].Number)
	}
}
