package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type MutationKind int

const (
	MutLabel MutationKind = iota
	MutAssignee
	MutClose
	MutReopen
	MutComment
	MutEdit
	MutReview
	MutMerge
	MutCreateIssue
)

type Mutation struct {
	Kind   MutationKind
	Repo   string
	Number int
	Labels []string
	Logins []string
	Body   string
	Title  string
	Event  string // APPROVE | COMMENT | REQUEST_CHANGES
	Method string // merge | squash | rebase
}

func (m Mutation) Description() string {
	switch m.Kind {
	case MutLabel:
		return "aplicar label " + strings.Join(m.Labels, ", ")
	case MutAssignee:
		if len(m.Logins) == 0 {
			return "limpar assignees"
		}
		return "atribuir a " + strings.Join(m.Logins, ", ")
	case MutClose:
		return "fechar"
	case MutReopen:
		return "reabrir"
	case MutComment:
		return "comentar"
	case MutEdit:
		return "editar"
	case MutReview:
		switch m.Event {
		case "APPROVE":
			return "aprovar"
		case "REQUEST_CHANGES":
			return "pedir mudanças"
		default:
			return "comentar no review"
		}
	case MutMerge:
		return "merge (" + m.Method + ")"
	case MutCreateIssue:
		return "criar issue"
	}
	return "?"
}

func (m Mutation) Destructive() bool {
	return m.Kind == MutClose || m.Kind == MutMerge
}

func (c *Client) Apply(repo string, m Mutation) error {
	if m.Repo != "" {
		repo = m.Repo
	}
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	base := fmt.Sprintf("repos/%s/%s", owner, name)

	switch m.Kind {
	case MutLabel:
		return c.post(base+"/issues/"+strconv.Itoa(m.Number)+"/labels",
			map[string]any{"labels": m.Labels}, m.Number)
	case MutAssignee:
		return c.post(base+"/issues/"+strconv.Itoa(m.Number)+"/assignees",
			map[string]any{"assignees": m.Logins}, m.Number)
	case MutClose:
		return c.patch(base+"/issues/"+strconv.Itoa(m.Number),
			map[string]any{"state": "closed"}, m.Number)
	case MutReopen:
		return c.patch(base+"/issues/"+strconv.Itoa(m.Number),
			map[string]any{"state": "open"}, m.Number)
	case MutComment:
		return c.post(base+"/issues/"+strconv.Itoa(m.Number)+"/comments",
			map[string]any{"body": m.Body}, m.Number)
	case MutEdit:
		campos := map[string]any{}
		if m.Title != "" {
			campos["title"] = m.Title
		}
		if m.Body != "" {
			campos["body"] = m.Body
		}
		return c.patch(base+"/issues/"+strconv.Itoa(m.Number), campos, m.Number)
	case MutReview:
		return c.post(base+"/pulls/"+strconv.Itoa(m.Number)+"/reviews",
			map[string]any{"event": m.Event, "body": m.Body}, m.Number)
	case MutMerge:
		return c.put(base+"/pulls/"+strconv.Itoa(m.Number)+"/merge",
			map[string]any{"merge_method": m.Method}, m.Number)
	case MutCreateIssue:
		return c.post(base+"/issues",
			map[string]any{"title": m.Title, "body": m.Body, "labels": m.Labels}, 0)
	}
	return fmt.Errorf("mutação não suportada")
}

func (c *Client) post(path string, corpo map[string]any, number int) error {
	return c.call("POST", path, corpo, number)
}

func (c *Client) patch(path string, corpo map[string]any, number int) error {
	return c.call("PATCH", path, corpo, number)
}

func (c *Client) put(path string, corpo map[string]any, number int) error {
	return c.call("PUT", path, corpo, number)
}

func (c *Client) call(metodo, path string, corpo map[string]any, number int) error {
	dados, err := json.Marshal(corpo)
	if err != nil {
		return err
	}
	var resp any
	if err := c.rest.Do(metodo, path, bytes.NewReader(dados), &resp); err != nil {
		if number > 0 {
			return fmt.Errorf("#%d: %w", number, err)
		}
		return err
	}
	return nil
}
