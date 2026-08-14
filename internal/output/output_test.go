package output

import (
	"strings"
	"testing"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func TestWriteTSVSanitizesAllFields(t *testing.T) {
	items := []gh.Item{{
		Number: 1,
		Repo:   "a/b",
		Title:  "linha1\nlinha2",
		Author: "ana\tzero",
		Labels: []gh.Label{{Name: "bug\tp1"}},
	}}
	var b strings.Builder
	if err := Write(&b, items, "tsv"); err != nil {
		t.Fatal(err)
	}
	linhas := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	if len(linhas) != 2 {
		t.Fatalf("esperava cabeçalho + 1 linha de dados, veio %d:\n%s", len(linhas), b.String())
	}
	cols := strings.Split(linhas[1], "\t")
	if len(cols) != len(columns) {
		t.Fatalf("linha tem %d colunas, esperava %d (tab/newline vazaram)", len(cols), len(columns))
	}
}
