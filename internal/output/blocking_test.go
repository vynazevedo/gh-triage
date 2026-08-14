package output

import (
	"strings"
	"testing"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func TestBlockingPlain(t *testing.T) {
	b := gh.Blocking{
		OnYou:    []gh.Item{{Number: 5, Repo: "a/b", Title: "Fix", Author: "ana"}},
		OnOthers: nil,
	}
	var buf strings.Builder
	if err := WriteBlocking(&buf, b, "plain"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, quer := range []string{
		"Bloqueio de review",
		"Esperando pelo seu review:",
		"a/b#5 Fix · @ana ·",
		"Você espera por review:",
		"  (nada)",
	} {
		if !strings.Contains(got, quer) {
			t.Errorf("blocking plain não contém %q\n---\n%s", quer, got)
		}
	}
}

func TestWaitText(t *testing.T) {
	if got := waitText("não é data"); got != "?" {
		t.Errorf("iso inválido=%q, quer ?", got)
	}
}
