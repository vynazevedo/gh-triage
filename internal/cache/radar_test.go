package cache

import (
	"testing"
	"time"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func useTempCache(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	orig := userCacheDir
	userCacheDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userCacheDir = orig })
}

func TestRadarRoundTrip(t *testing.T) {
	useTempCache(t)
	items := []gh.RadarItem{
		{Item: gh.Item{Number: 7, Repo: "a/b", Title: "x", CI: gh.CIFail}, Reason: "ci falhou", Class: 2},
		{Item: gh.Item{Number: 9, Repo: "a/b", Title: "y"}, Reason: "atribuída a você", Class: 4},
	}
	if err := SaveRadar(items, 4200); err != nil {
		t.Fatalf("SaveRadar: %v", err)
	}
	snap, ok := LoadRadar()
	if !ok {
		t.Fatal("esperava encontrar o cache recém-salvo")
	}
	if len(snap.Items) != 2 {
		t.Fatalf("itens=%d, quer 2", len(snap.Items))
	}
	if snap.Items[0].Number != 7 || snap.Items[0].Class != 2 || snap.Items[0].CI != gh.CIFail {
		t.Errorf("primeiro item não sobreviveu ao round-trip: %+v", snap.Items[0])
	}
	if snap.Rate != 4200 {
		t.Errorf("rate=%d, quer 4200", snap.Rate)
	}
	if snap.Age() > time.Minute {
		t.Errorf("idade do cache recém-salvo grande demais: %v", snap.Age())
	}
}

func TestLoadMissing(t *testing.T) {
	useTempCache(t)
	if _, ok := LoadRadar(); ok {
		t.Fatal("não deveria haver cache num diretório vazio")
	}
}

func TestAgeZeroValue(t *testing.T) {
	if (RadarSnapshot{}).Age() < time.Hour*24*365 {
		t.Error("snapshot sem SavedAt deveria ser tratado como muito antigo")
	}
}
