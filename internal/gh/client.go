package gh

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

type Client struct {
	gql  *api.GraphQLClient
	rest *api.RESTClient
}

func NewClient() (*Client, error) {
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o cliente GraphQL do gh: %w", err)
	}
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o cliente REST do gh: %w", err)
	}
	return &Client{gql: gql, rest: rest}, nil
}

func CurrentRepo() (string, error) {
	r, err := repository.Current()
	if err != nil {
		return "", fmt.Errorf("falha ao descobrir o repositório atual (use -R owner/repo): %w", err)
	}
	return r.Owner + "/" + r.Name, nil
}

func ValidateRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf("repositório vazio. Use owner/repo ou owner (org)")
	}
	partes := strings.Split(repo, "/")
	if len(partes) > 2 || partes[0] == "" || (len(partes) == 2 && partes[1] == "") {
		return fmt.Errorf("inválido: %q. Use owner/repo ou owner (org)", repo)
	}
	return nil
}

func (c *Client) Search(search, cursor string) (Page, error) {
	vars := map[string]any{"q": search, "cursor": (*string)(nil)}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	var resp searchResponse
	if err := c.gql.Do(searchQuery, vars, &resp); err != nil {
		return Page{}, fmt.Errorf("falha ao buscar %q na API do GitHub: %w", search, err)
	}
	return pageFromResponse(resp), nil
}

func pageFromResponse(resp searchResponse) Page {
	items := make([]Item, 0, len(resp.Search.Nodes))
	for _, no := range resp.Search.Nodes {
		switch no.Typename {
		case "Issue":
			items = append(items, itemFromIssue(no))
		case "PullRequest":
			items = append(items, itemFromPR(no))
		}
	}

	var proximo string
	if resp.Search.PageInfo.HasNextPage {
		proximo = resp.Search.PageInfo.EndCursor
	}
	rl := 0
	if resp.RateLimit != nil {
		rl = resp.RateLimit.Remaining
	}

	return Page{
		Total:              resp.Search.IssueCount,
		Items:              items,
		NextCursor:         proximo,
		RateLimitRemaining: rl,
	}
}

func itemFromIssue(no searchNode) Item {
	repoIssue := ""
	if no.Repository != nil {
		repoIssue = no.Repository.NameWithOwner
	}
	return Item{
		Number:    no.Number,
		Repo:      repoIssue,
		Kind:      KindIssue,
		State:     parseState(no.State),
		Title:     no.Title,
		Body:      no.Body,
		URL:       no.URL,
		Author:    login(no.Author),
		Labels:    labelsOf(no),
		Assignees: assigneesOf(no),
		Milestone: milestoneOf(no),
		Comments:  no.Comments.TotalCount,
		CreatedAt: no.CreatedAt,
		UpdatedAt: no.UpdatedAt,
		Linked:    linkedOf(no.TimelineItems.Nodes, repoIssue),
	}
}

func itemFromPR(no searchNode) Item {
	repoPR := ""
	if no.Repository != nil {
		repoPR = no.Repository.NameWithOwner
	}
	ci := CINone
	for _, n := range no.StatusCheckRollup.Nodes {
		if n.Commit != nil && n.Commit.StatusCheckRollup != nil {
			ci = parseCI(n.Commit.StatusCheckRollup.State)
			break
		}
	}
	return Item{
		Number:         no.Number,
		Repo:           repoPR,
		Kind:           KindPR,
		State:          parseState(no.State),
		Draft:          no.IsDraft,
		Title:          no.Title,
		Body:           no.Body,
		URL:            no.URL,
		Author:         login(no.Author),
		Labels:         labelsOf(no),
		Assignees:      assigneesOf(no),
		Milestone:      milestoneOf(no),
		Comments:       no.Comments.TotalCount,
		CreatedAt:      no.CreatedAt,
		UpdatedAt:      no.UpdatedAt,
		ReviewDecision: no.ReviewDecision,
		CI:             ci,
	}
}

func login(a *actor) string {
	if a == nil {
		return ""
	}
	return a.Login
}

func labelsOf(no searchNode) []Label {
	out := make([]Label, 0, len(no.Labels.Nodes))
	for _, l := range no.Labels.Nodes {
		out = append(out, Label{Name: l.Name, Color: l.Color})
	}
	return out
}

func assigneesOf(no searchNode) []string {
	out := make([]string, 0, len(no.Assignees.Nodes))
	for _, a := range no.Assignees.Nodes {
		out = append(out, a.Login)
	}
	return out
}

func milestoneOf(no searchNode) string {
	if no.Milestone != nil {
		return no.Milestone.Title
	}
	return ""
}

func linkedOf(nodes []timelineNode, repoDaIssue string) []LinkedPR {
	var out []LinkedPR
	for _, ev := range nodes {
		ref := ev.Source
		if ref == nil {
			ref = ev.Subject
		}
		if ref == nil || ref.Typename != "PullRequest" {
			continue
		}
		repoDoPR := ""
		if ref.Repository != nil {
			repoDoPR = ref.Repository.NameWithOwner
		}
		externo := ""
		if repoDoPR != "" && repoDoPR != repoDaIssue {
			externo = repoDoPR
		}
		duplicado := false
		for _, v := range out {
			if v.Number == ref.Number && v.ExternalRepo == externo {
				duplicado = true
				break
			}
		}
		if duplicado {
			continue
		}
		out = append(out, LinkedPR{
			Number:       ref.Number,
			State:        parseState(ref.State),
			Draft:        ref.IsDraft,
			ExternalRepo: externo,
		})
	}
	return out
}
