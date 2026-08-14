package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

func WriteBlocking(w io.Writer, b gh.Blocking, format string) error {
	if format == "json" {
		return blockingJSON(w, b)
	}
	return blockingPlain(w, b)
}

func waitText(iso string) string {
	d, ok := gh.AgeSince(iso)
	if !ok {
		return "?"
	}
	dias := int(d.Hours()) / 24
	if dias < 1 {
		return "hoje"
	}
	if dias == 1 {
		return "há 1 dia"
	}
	return "há " + strconv.Itoa(dias) + " dias"
}

func blockingPlain(w io.Writer, b gh.Blocking) error {
	if _, err := fmt.Fprintln(w, "Bloqueio de review"); err != nil {
		return err
	}
	secoes := []struct {
		titulo string
		itens  []gh.Item
		quem   func(gh.Item) string
	}{
		{"Esperando pelo seu review", b.OnYou, func(it gh.Item) string { return "@" + orDash(it.Author) }},
		{"Você espera por review", b.OnOthers, func(it gh.Item) string { return "sem review ainda" }},
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
			linha := "  - " + it.Repo + "#" + strconv.Itoa(it.Number) + " " + it.Title +
				" · " + s.quem(it) + " · " + waitText(it.UpdatedAt)
			if _, err := fmt.Fprintln(w, linha); err != nil {
				return err
			}
		}
	}
	return nil
}

func orDash(s string) string {
	if s == "" {
		return "desconhecido"
	}
	return s
}

func blockingJSON(w io.Writer, b gh.Blocking) error {
	conv := func(items []gh.Item) []record {
		out := make([]record, 0, len(items))
		for _, it := range items {
			out = append(out, toRecord(it))
		}
		return out
	}
	obj := struct {
		OnYou    []record `json:"on_you"`
		OnOthers []record `json:"on_others"`
	}{OnYou: conv(b.OnYou), OnOthers: conv(b.OnOthers)}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		return fmt.Errorf("falha ao serializar bloqueio em JSON: %w", err)
	}
	return nil
}
