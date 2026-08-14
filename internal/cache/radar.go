package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

type RadarSnapshot struct {
	SavedAt time.Time      `json:"saved_at"`
	Items   []gh.RadarItem `json:"items"`
	Rate    int            `json:"rate,omitempty"`
}

func (s RadarSnapshot) Age() time.Duration {
	if s.SavedAt.IsZero() {
		return time.Duration(1<<62 - 1)
	}
	return time.Since(s.SavedAt)
}

func radarPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gh-triage", "radar.json"), nil
}

func LoadRadar() (RadarSnapshot, bool) {
	caminho, err := radarPath()
	if err != nil {
		return RadarSnapshot{}, false
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return RadarSnapshot{}, false
	}
	var snap RadarSnapshot
	if json.Unmarshal(dados, &snap) != nil {
		return RadarSnapshot{}, false
	}
	return snap, true
}

func SaveRadar(items []gh.RadarItem, rate int) error {
	caminho, err := radarPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		return err
	}
	dados, err := json.Marshal(RadarSnapshot{SavedAt: time.Now(), Items: items, Rate: rate})
	if err != nil {
		return err
	}
	tmp := caminho + ".tmp"
	if err := os.WriteFile(tmp, dados, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, caminho)
}
