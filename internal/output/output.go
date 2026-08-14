package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/vynazevedo/gh-triage/internal/gh"
)

type record struct {
	Number         int      `json:"number"`
	Type           string   `json:"type"`
	State          string   `json:"state"`
	Title          string   `json:"title"`
	Repo           string   `json:"repo"`
	Author         string   `json:"author"`
	Labels         []string `json:"labels"`
	Assignees      []string `json:"assignees"`
	Milestone      string   `json:"milestone"`
	URL            string   `json:"url"`
	Draft          bool     `json:"draft"`
	CI             string   `json:"ci"`
	ReviewDecision string   `json:"review_decision"`
	Comments       int      `json:"comments"`
	Linked         []string `json:"linked"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	Reason         string   `json:"reason,omitempty"`
}

var columns = []string{
	"number", "type", "state", "title", "repo", "author",
	"labels", "assignees", "milestone", "url", "draft", "ci",
	"review_decision", "comments", "linked", "created_at", "updated_at", "reason",
}

func toRecord(item gh.Item) record {
	labels := make([]string, 0, len(item.Labels))
	for _, l := range item.Labels {
		labels = append(labels, l.Name)
	}
	assignees := item.Assignees
	if assignees == nil {
		assignees = []string{}
	}
	linked := make([]string, 0, len(item.Linked))
	for _, v := range item.Linked {
		linked = append(linked, v.Reference())
	}
	return record{
		Number:         item.Number,
		Type:           item.Kind.String(),
		State:          item.State.String(),
		Title:          item.Title,
		Repo:           item.Repo,
		Author:         item.Author,
		Labels:         labels,
		Assignees:      assignees,
		Milestone:      item.Milestone,
		URL:            item.URL,
		Draft:          item.Draft,
		CI:             item.CI.String(),
		ReviewDecision: item.ReviewDecision,
		Comments:       item.Comments,
		Linked:         linked,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func WriteRadar(w io.Writer, items []gh.RadarItem, format string) error {
	regs := make([]record, 0, len(items))
	for _, it := range items {
		r := toRecord(it.Item)
		r.Reason = it.Reason
		regs = append(regs, r)
	}
	switch format {
	case "tsv":
		return writeTSV(w, regs)
	case "plain":
		return writePlain(w, regs)
	default:
		return writeJSON(w, regs)
	}
}

func Write(w io.Writer, items []gh.Item, format string) error {
	regs := make([]record, 0, len(items))
	for _, i := range items {
		regs = append(regs, toRecord(i))
	}
	switch format {
	case "tsv":
		return writeTSV(w, regs)
	case "plain":
		return writePlain(w, regs)
	default:
		return writeJSON(w, regs)
	}
}

func writePlain(w io.Writer, regs []record) error {
	for _, r := range regs {
		kind := "issue"
		if r.Type == "pr" {
			kind = "PR"
		}
		linha := fmt.Sprintf("%-5s #%-6d [%s] %s", kind, r.Number, r.State, r.Title)
		if r.Reason != "" {
			linha = fmt.Sprintf("%-5s #%-6d [%s] %-38s %s  (%s)", kind, r.Number, r.State, r.Repo, r.Title, r.Reason)
		}
		if _, err := fmt.Fprintln(w, linha); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w io.Writer, regs []record) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(regs); err != nil {
		return fmt.Errorf("falha ao serializar em JSON: %w", err)
	}
	return nil
}

func writeTSV(w io.Writer, regs []record) error {
	if _, err := fmt.Fprintln(w, strings.Join(columns, "\t")); err != nil {
		return err
	}
	for _, r := range regs {
		linha := []string{
			strconv.Itoa(r.Number),
			r.Type,
			r.State,
			sanitize(r.Title),
			sanitize(r.Repo),
			sanitize(r.Author),
			sanitize(strings.Join(r.Labels, ",")),
			sanitize(strings.Join(r.Assignees, ",")),
			sanitize(r.Milestone),
			r.URL,
			strconv.FormatBool(r.Draft),
			r.CI,
			r.ReviewDecision,
			strconv.Itoa(r.Comments),
			sanitize(strings.Join(r.Linked, ",")),
			r.CreatedAt,
			r.UpdatedAt,
			sanitize(r.Reason),
		}
		if _, err := fmt.Fprintln(w, strings.Join(linha, "\t")); err != nil {
			return err
		}
	}
	return nil
}

func sanitize(v string) string {
	return strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(v)
}
