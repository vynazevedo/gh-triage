package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

type SavedSearch struct {
	Name    string     `json:"nome"`
	Filters gh.Filters `json:"filtros"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gh-triage", "buscas.json"), nil
}

func loadSearches() []SavedSearch {
	caminho, err := configPath()
	if err != nil {
		return nil
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil
	}
	var buscas []SavedSearch
	if json.Unmarshal(dados, &buscas) != nil {
		return nil
	}
	return buscas
}

func saveSearches(buscas []SavedSearch) error {
	caminho, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		return err
	}
	dados, err := json.MarshalIndent(buscas, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(caminho, dados, 0o644)
}
