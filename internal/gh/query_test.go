package gh

import "testing"

func TestBuildQueryMinimal(t *testing.T) {
	got := BuildQuery(Filters{Repo: "cli/cli", State: "open"})
	want := "repo:cli/cli is:open"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildQueryStateAllNoRestrict(t *testing.T) {
	got := BuildQuery(Filters{Repo: "cli/cli", State: "all"})
	if got != "repo:cli/cli" {
		t.Fatalf("estado all não deveria emitir cláusula: %q", got)
	}
}

func TestBuildQueryKindBothNoRestrict(t *testing.T) {
	got := BuildQuery(Filters{Repo: "cli/cli", State: "open"})
	if contains(got, "is:issue") || contains(got, "is:pr") {
		t.Fatalf("tipo padrão deveria trazer issues e PRs juntos: %q", got)
	}
}

func TestBuildQueryLabelsAndQuotes(t *testing.T) {
	got := BuildQuery(Filters{
		Repo:      "cli/cli",
		State:     "open",
		Labels:    []string{"bug", "boa primeira"},
		Milestone: `v2 "beta"`,
		Term:      "  avaliação  ",
	})
	want := `repo:cli/cli is:open label:bug label:"boa primeira" milestone:"v2 \"beta\"" avaliação`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestOrgScopeVsRepo(t *testing.T) {
	if !OrgScope("cli") {
		t.Error("owner sem barra deveria ser escopo de org")
	}
	if OrgScope("cli/cli") {
		t.Error("owner/repo não é escopo de org")
	}
	if got := BuildQuery(Filters{Repo: "cli", State: "open"}); got != "org:cli is:open" {
		t.Errorf("org query=%q, quer org:cli is:open", got)
	}
	if got := BuildQuery(Filters{Repo: "cli/cli", State: "open"}); got != "repo:cli/cli is:open" {
		t.Errorf("repo query=%q", got)
	}
}

func TestValidateRepoAcceptsOrg(t *testing.T) {
	for _, ok := range []string{"cli", "cli/cli"} {
		if err := ValidateRepo(ok); err != nil {
			t.Errorf("ValidarRepo(%q) deveria aceitar: %v", ok, err)
		}
	}
	for _, ruim := range []string{"", "a/b/c", "/x"} {
		if err := ValidateRepo(ruim); err == nil {
			t.Errorf("ValidarRepo(%q) deveria rejeitar", ruim)
		}
	}
}
