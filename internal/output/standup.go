package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func WriteStandup(w io.Writer, su gh.Standup, format string) error {
	if format == "json" {
		return standupJSON(w, su)
	}
	return standupPlain(w, su)
}

func standupPlain(w io.Writer, su gh.Standup) error {
	dias := strconv.Itoa(su.SinceDays)
	if _, err := fmt.Fprintln(w, "Standup dos últimos "+dias+" dias"); err != nil {
		return err
	}
	secoes := []struct {
		titulo string
		itens  []gh.Item
		motivo func(gh.Item) string
	}{
		{"Entregue (merged)", su.Shipped, nil},
		{"Em andamento", su.Ongoing, nil},
		{"Revisando (pedido a você)", su.Reviewing, nil},
		{"Bloqueado", su.Blocked, blockReason},
	}
	for _, s := range secoes {
		if _, err := fmt.Fprintf(w, "\n%s:\n", s.titulo); err != nil {
			return err
		}
		if len(s.itens) == 0 {
			if _, err := fmt.Fprintln(w, "  (nada)"); err != nil {
				return err
			}
			continue
		}
		for _, it := range s.itens {
			linha := "  - " + it.Repo + "#" + strconv.Itoa(it.Number) + " " + it.Title
			if s.motivo != nil {
				if r := s.motivo(it); r != "" {
					linha += " (" + r + ")"
				}
			}
			if _, err := fmt.Fprintln(w, linha); err != nil {
				return err
			}
		}
	}
	return nil
}

func blockReason(it gh.Item) string {
	switch {
	case it.ReviewDecision == "CHANGES_REQUESTED":
		return "mudanças pedidas"
	case it.CI == gh.CIFail:
		return "CI vermelho"
	default:
		return ""
	}
}

func standupJSON(w io.Writer, su gh.Standup) error {
	conv := func(items []gh.Item) []record {
		out := make([]record, 0, len(items))
		for _, it := range items {
			out = append(out, toRecord(it))
		}
		return out
	}
	obj := struct {
		SinceDays int      `json:"since_days"`
		Shipped   []record `json:"shipped"`
		Ongoing   []record `json:"ongoing"`
		Reviewing []record `json:"reviewing"`
		Blocked   []record `json:"blocked"`
	}{
		SinceDays: su.SinceDays,
		Shipped:   conv(su.Shipped),
		Ongoing:   conv(su.Ongoing),
		Reviewing: conv(su.Reviewing),
		Blocked:   conv(su.Blocked),
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		return fmt.Errorf("falha ao serializar standup em JSON: %w", err)
	}
	return nil
}
