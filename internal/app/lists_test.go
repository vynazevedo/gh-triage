package app

import (
	"testing"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func item(number int, title string) itemRow {
	return itemRow{Item: gh.Item{Number: number, Kind: gh.KindIssue, Title: title}}
}

func batch(n int) []itemRow {
	out := make([]itemRow, n)
	for i := range out {
		out[i] = item(i+1, "x")
	}
	return out
}

func TestRefilterFuzzyRanking(t *testing.T) {
	l := newList[itemRow]()
	l.replace([]itemRow{
		item(1, "documenta o release"),
		item(2, "corrige avaliação de risco"),
		item(3, "avaliação"),
	}, 3, "", "")

	l.refilter("avaliação")
	if len(l.visible) != 2 {
		t.Fatalf("visiveis=%d, quer 2", len(l.visible))
	}
	if l.items[l.visible[0]].Number != 3 {
		t.Errorf("melhor match deveria vir primeiro, veio #%d", l.items[l.visible[0]].Number)
	}

	l.refilter("zzz")
	if len(l.visible) != 0 {
		t.Fatalf("termo sem match deveria zerar, veio %d", len(l.visible))
	}

	l.refilter("")
	if len(l.visible) != 3 {
		t.Fatalf("termo vazio deveria mostrar tudo, veio %d", len(l.visible))
	}
}

func TestReplacePopulatesVisible(t *testing.T) {
	l := newList[itemRow]()
	l.replace(batch(50), 1000, "CURSOR-2", "")
	if len(l.visible) != 50 {
		t.Fatalf("visiveis=%d, quer 50", len(l.visible))
	}
	if l.total != 1000 || l.nextCursor != "CURSOR-2" {
		t.Fatalf("total=%d cursor=%q", l.total, l.nextCursor)
	}
}

func TestShouldPaginateAtThreshold(t *testing.T) {
	l := newList[itemRow]()
	l.replace(batch(50), 1000, "CURSOR-2", "")

	l.cursor = 0
	if l.shouldPaginate() {
		t.Error("no topo não deveria paginar")
	}
	l.cursor = 40
	if !l.shouldPaginate() {
		t.Error("em 80% deveria paginar")
	}
	l.paginating = true
	if l.shouldPaginate() {
		t.Error("não deveria paginar enquanto já pagina")
	}
}

func TestShouldPaginateNoCursorFalse(t *testing.T) {
	l := newList[itemRow]()
	l.replace(batch(50), 50, "", "")
	l.cursor = 49
	if l.shouldPaginate() {
		t.Error("sem próximo cursor não deveria paginar")
	}
}

func TestAppendDedupPreservesCursor(t *testing.T) {
	l := newList[itemRow]()
	l.replace(batch(50), 1000, "C2", "")
	l.cursor = 45
	chaveAntes := l.items[l.visible[l.cursor]].Key()

	segunda := make([]itemRow, 0, 51)
	segunda = append(segunda, item(50, "x"))
	for i := 51; i <= 100; i++ {
		segunda = append(segunda, item(i, "y"))
	}
	l.appendItems(segunda, "C3", "")

	if len(l.items) != 100 {
		t.Fatalf("itens=%d, quer 100 (item 50 repetido não duplica)", len(l.items))
	}
	if l.nextCursor != "C3" {
		t.Fatalf("cursor=%q", l.nextCursor)
	}
	if got := l.items[l.visible[l.cursor]].Key(); got != chaveAntes {
		t.Errorf("cursor pulou de %q para %q", chaveAntes, got)
	}
}

func TestSelectionAndSelected(t *testing.T) {
	l := newList[itemRow]()
	l.replace(batch(3), 3, "", "")
	l.cursor = 1
	l.toggleSelect()
	if len(l.markedItems()) != 1 || l.markedItems()[0].Number != 2 {
		t.Fatalf("marcados=%v", l.markedItems())
	}
	l.toggleAll()
	if len(l.markedItems()) != 3 {
		t.Fatalf("alternarTodos deveria marcar 3, veio %d", len(l.markedItems()))
	}
	l.toggleAll()
	if len(l.selected) != 0 {
		t.Fatal("segundo alternarTodos deveria limpar")
	}
}
