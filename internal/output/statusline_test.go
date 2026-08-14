package output

import (
	"strings"
	"testing"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func TestSummarize(t *testing.T) {
	items := []gh.RadarItem{
		{Class: 0}, {Class: 0}, {Class: 1}, {Class: 2}, {Class: 6}, {Class: 7},
	}
	c := Summarize(items)
	if c.Review != 2 || c.Changes != 1 || c.CI != 1 || c.Waiting != 1 || c.Drafts != 1 {
		t.Fatalf("contagens erradas: %+v", c)
	}
	if c.Total != 6 {
		t.Errorf("total=%d, quer 6", c.Total)
	}
	if c.Urgent != 4 {
		t.Errorf("urgent=%d, quer 4 (review+changes+ci)", c.Urgent)
	}
}

func TestWriteCountEmpty(t *testing.T) {
	var b strings.Builder
	if err := WriteCount(&b, nil, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "radar limpo") {
		t.Errorf("radar vazio deveria dizer 'radar limpo', veio %q", b.String())
	}
}

func TestWriteCountLine(t *testing.T) {
	items := []gh.RadarItem{{Class: 0}, {Class: 6}, {Class: 6}}
	var b strings.Builder
	if err := WriteCount(&b, items, false); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, quer := range []string{"1 review", "2 aguardando", "3 total"} {
		if !strings.Contains(got, quer) {
			t.Errorf("linha %q não contém %q", strings.TrimSpace(got), quer)
		}
	}
}

func TestWriteCountStaleMarker(t *testing.T) {
	var b strings.Builder
	if err := WriteCount(&b, []gh.RadarItem{{Class: 0}}, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "*") {
		t.Errorf("dado velho deveria marcar com *, veio %q", strings.TrimSpace(b.String()))
	}
}

func TestWriteCountJSON(t *testing.T) {
	var b strings.Builder
	items := []gh.RadarItem{{Class: 0}, {Class: 3}}
	if err := WriteCountJSON(&b, items, 12, false); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, quer := range []string{`"review":1`, `"merge":1`, `"total":2`, `"age_seconds":12`, `"stale":false`} {
		if !strings.Contains(got, quer) {
			t.Errorf("json %q não contém %q", strings.TrimSpace(got), quer)
		}
	}
}
