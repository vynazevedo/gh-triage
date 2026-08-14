package app

import (
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

type keyable interface {
	Key() string
	searchText() string
}

type itemRow struct {
	gh.Item
	Reason string
	Class  int
}

func (i itemRow) searchText() string {
	var b strings.Builder
	b.WriteString("#")
	b.WriteString(i.Item.Key())
	b.WriteByte(' ')
	b.WriteString(i.Reason)
	b.WriteByte(' ')
	b.WriteString(i.Repo)
	b.WriteByte(' ')
	b.WriteString(i.Title)
	b.WriteByte(' ')
	b.WriteString(i.Author)
	b.WriteByte(' ')
	for _, l := range i.Labels {
		b.WriteString(l.Name)
		b.WriteByte(' ')
	}
	for _, a := range i.Assignees {
		b.WriteString(a)
		b.WriteByte(' ')
	}
	return b.String()
}

type commitRow struct{ gh.Commit }

func (c commitRow) searchText() string {
	var b strings.Builder
	b.WriteString(c.ShortOID())
	b.WriteByte(' ')
	b.WriteString(c.Title)
	b.WriteByte(' ')
	b.WriteString(c.Author)
	return b.String()
}

type list[T keyable] struct {
	items      []T
	visible    []int
	cursor     int
	selected   map[string]bool
	total      int
	loaded     bool
	nextCursor string
	paginating bool
	scroll     int
}

func newList[T keyable]() list[T] {
	return list[T]{selected: map[string]bool{}}
}

func (l *list[T]) current() (T, bool) {
	var zero T
	if l.cursor < 0 || l.cursor >= len(l.visible) {
		return zero, false
	}
	return l.items[l.visible[l.cursor]], true
}

func (l *list[T]) markedItems() []T {
	if len(l.selected) == 0 {
		if it, ok := l.current(); ok {
			return []T{it}
		}
		return nil
	}
	var out []T
	for _, it := range l.items {
		if l.selected[it.Key()] {
			out = append(out, it)
		}
	}
	return out
}

func (l *list[T]) refilter(term string) {
	var chaveAnterior string
	if it, ok := l.current(); ok {
		chaveAnterior = it.Key()
	}
	l.visible = l.visible[:0]
	if strings.TrimSpace(term) == "" {
		for i := range l.items {
			l.visible = append(l.visible, i)
		}
	} else {
		fontes := make([]string, len(l.items))
		for i := range l.items {
			fontes[i] = l.items[i].searchText()
		}
		for _, m := range fuzzy.Find(strings.TrimSpace(term), fontes) {
			l.visible = append(l.visible, m.Index)
		}
	}
	l.cursor = 0
	if chaveAnterior != "" {
		for pos, idx := range l.visible {
			if l.items[idx].Key() == chaveAnterior {
				l.cursor = pos
				break
			}
		}
	}
	if l.cursor >= len(l.visible) && len(l.visible) > 0 {
		l.cursor = len(l.visible) - 1
	}
	l.scroll = 0
}

func (l *list[T]) move(delta int) {
	if len(l.visible) == 0 {
		return
	}
	l.cursor += delta
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor >= len(l.visible) {
		l.cursor = len(l.visible) - 1
	}
	l.scroll = 0
}

func (l *list[T]) toggleSelect() {
	if it, ok := l.current(); ok {
		k := it.Key()
		if l.selected[k] {
			delete(l.selected, k)
		} else {
			l.selected[k] = true
		}
	}
}

func (l *list[T]) toggleAll() {
	todosMarcados := len(l.visible) > 0
	for _, idx := range l.visible {
		if !l.selected[l.items[idx].Key()] {
			todosMarcados = false
			break
		}
	}
	if todosMarcados {
		l.selected = map[string]bool{}
		return
	}
	for _, idx := range l.visible {
		l.selected[l.items[idx].Key()] = true
	}
}

func (l *list[T]) replace(items []T, total int, cursor, term string) {
	l.items = items
	l.total = total
	l.nextCursor = cursor
	l.paginating = false
	l.loaded = true
	novas := map[string]bool{}
	for _, it := range items {
		if l.selected[it.Key()] {
			novas[it.Key()] = true
		}
	}
	l.selected = novas
	l.refilter(term)
}

func (l *list[T]) appendItems(items []T, cursor, term string) {
	existentes := map[string]bool{}
	for _, it := range l.items {
		existentes[it.Key()] = true
	}
	for _, it := range items {
		if !existentes[it.Key()] {
			l.items = append(l.items, it)
		}
	}
	l.nextCursor = cursor
	l.paginating = false
	l.refilter(term)
}

func (l *list[T]) shouldPaginate() bool {
	if l.paginating || l.nextCursor == "" || len(l.items) == 0 {
		return false
	}
	if l.cursor >= len(l.visible) {
		return false
	}
	limiar := len(l.items) * 4 / 5
	return l.visible[l.cursor] >= limiar
}
