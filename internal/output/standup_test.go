package output

import (
	"strings"
	"testing"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func TestStandupPlain(t *testing.T) {
	su := gh.Standup{
		SinceDays: 2,
		Shipped:   []gh.Item{{Number: 10, Repo: "a/b", Title: "Feature X"}},
		Blocked:   []gh.Item{{Number: 20, Repo: "a/b", Title: "Fix Y", CI: gh.CIFail}},
	}
	var b strings.Builder
	if err := WriteStandup(&b, su, "plain"); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, quer := range []string{
		"últimos 2 dias",
		"Entregue",
		"a/b#10 Feature X",
		"Em andamento:",
		"  (nada)",
		"a/b#20 Fix Y (CI vermelho)",
	} {
		if !strings.Contains(got, quer) {
			t.Errorf("standup plain não contém %q\n---\n%s", quer, got)
		}
	}
}

func TestStandupJSON(t *testing.T) {
	su := gh.Standup{SinceDays: 1, Reviewing: []gh.Item{{Number: 5, Repo: "a/b", Title: "z"}}}
	var b strings.Builder
	if err := WriteStandup(&b, su, "json"); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, quer := range []string{`"since_days": 1`, `"reviewing"`, `"number": 5`, `"shipped": []`} {
		if !strings.Contains(got, quer) {
			t.Errorf("standup json não contém %q\n---\n%s", quer, got)
		}
	}
}
