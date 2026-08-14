package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

type Counts struct {
	Review   int `json:"review"`
	Changes  int `json:"changes"`
	CI       int `json:"ci"`
	Merge    int `json:"merge"`
	Assigned int `json:"assigned"`
	Mentions int `json:"mentions"`
	Waiting  int `json:"waiting"`
	Drafts   int `json:"drafts"`
	Total    int `json:"total"`
	Urgent   int `json:"urgent"`
}

func Summarize(items []gh.RadarItem) Counts {
	var c Counts
	for _, it := range items {
		switch it.Class {
		case 0:
			c.Review++
		case 1:
			c.Changes++
		case 2:
			c.CI++
		case 3:
			c.Merge++
		case 4:
			c.Assigned++
		case 5:
			c.Mentions++
		case 6:
			c.Waiting++
		case 7:
			c.Drafts++
		}
	}
	c.Total = len(items)
	c.Urgent = c.Review + c.Changes + c.CI
	return c
}

func WriteCount(w io.Writer, items []gh.RadarItem, stale bool) error {
	c := Summarize(items)
	var partes []string
	add := func(n int, rotulo string) {
		if n > 0 {
			partes = append(partes, fmt.Sprintf("%d %s", n, rotulo))
		}
	}
	add(c.Review, "review")
	add(c.Changes, "mudanças")
	add(c.CI, "ci")
	add(c.Merge, "merge")
	add(c.Assigned, "atribuídas")
	add(c.Mentions, "menções")
	add(c.Waiting, "aguardando")

	var linha string
	if len(partes) == 0 {
		linha = "radar limpo"
	} else {
		linha = strings.Join(partes, " · ") + fmt.Sprintf(" · %d total", c.Total)
	}
	if stale {
		linha += " *"
	}
	_, err := fmt.Fprintln(w, linha)
	return err
}

func WriteCountJSON(w io.Writer, items []gh.RadarItem, ageSeconds int, stale bool) error {
	obj := struct {
		Counts
		AgeSeconds int  `json:"age_seconds"`
		Stale      bool `json:"stale"`
	}{Counts: Summarize(items), AgeSeconds: ageSeconds, Stale: stale}
	enc := json.NewEncoder(w)
	if err := enc.Encode(obj); err != nil {
		return fmt.Errorf("falha ao serializar contagens em JSON: %w", err)
	}
	return nil
}
