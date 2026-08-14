package gh

import (
	"fmt"
	"strings"
)

const commitsQuery = `
query($owner: String!, $name: String!, $cursor: String) {
  rateLimit { remaining }
  repository(owner: $owner, name: $name) {
    defaultBranchRef {
      target {
        ... on Commit {
          history(first: 50, after: $cursor) {
            pageInfo { hasNextPage endCursor }
            nodes {
              oid messageHeadline messageBody committedDate url
              additions deletions
              author { user { login } name }
              associatedPullRequests(first: 5) { nodes { number } }
              statusCheckRollup { state }
            }
          }
        }
      }
    }
  }
}`

type commitsResponse struct {
	RateLimit *struct {
		Remaining int `json:"remaining"`
	} `json:"rateLimit"`
	Repository struct {
		DefaultBranchRef struct {
			Target struct {
				History struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []commitNode `json:"nodes"`
				} `json:"history"`
			} `json:"target"`
		} `json:"defaultBranchRef"`
	} `json:"repository"`
}

type commitNode struct {
	OID             string `json:"oid"`
	MessageHeadline string `json:"messageHeadline"`
	MessageBody     string `json:"messageBody"`
	CommittedDate   string `json:"committedDate"`
	URL             string `json:"url"`
	Additions       int    `json:"additions"`
	Deletions       int    `json:"deletions"`
	Author          *struct {
		User *actor  `json:"user"`
		Name *string `json:"name"`
	} `json:"author"`
	AssociatedPullRequests struct {
		Nodes []struct {
			Number int `json:"number"`
		} `json:"nodes"`
	} `json:"associatedPullRequests"`
	StatusCheckRollup *struct {
		State string `json:"state"`
	} `json:"statusCheckRollup"`
}

func (c *Client) FetchCommits(repo, cursor string) (CommitsPage, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return CommitsPage{}, err
	}
	vars := map[string]any{"owner": owner, "name": name, "cursor": (*string)(nil)}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	var resp commitsResponse
	if err := c.gql.Do(commitsQuery, vars, &resp); err != nil {
		return CommitsPage{}, fmt.Errorf("falha ao listar commits de %q: %w", repo, err)
	}
	return commitsFromResponse(resp), nil
}

func commitsFromResponse(resp commitsResponse) CommitsPage {
	hist := resp.Repository.DefaultBranchRef.Target.History
	commits := make([]Commit, 0, len(hist.Nodes))
	for _, n := range hist.Nodes {
		ci := CINone
		if n.StatusCheckRollup != nil {
			ci = parseCI(n.StatusCheckRollup.State)
		}
		prs := make([]int, 0, len(n.AssociatedPullRequests.Nodes))
		for _, p := range n.AssociatedPullRequests.Nodes {
			prs = append(prs, p.Number)
		}
		commits = append(commits, Commit{
			OID:         n.OID,
			Title:       n.MessageHeadline,
			Body:        n.MessageBody,
			Author:      commitAuthor(n),
			Date:        n.CommittedDate,
			URL:         n.URL,
			PRs:         prs,
			CitedIssues: CitedNumbers(n.MessageHeadline + " " + n.MessageBody),
			CI:          ci,
			Additions:   n.Additions,
			Deletions:   n.Deletions,
		})
	}

	var proximo string
	if hist.PageInfo.HasNextPage {
		proximo = hist.PageInfo.EndCursor
	}
	rl := 0
	if resp.RateLimit != nil {
		rl = resp.RateLimit.Remaining
	}
	return CommitsPage{Commits: commits, NextCursor: proximo, RateLimitRemaining: rl}
}

func commitAuthor(n commitNode) string {
	if n.Author == nil {
		return ""
	}
	if n.Author.User != nil && n.Author.User.Login != "" {
		return n.Author.User.Login
	}
	if n.Author.Name != nil {
		return *n.Author.Name
	}
	return ""
}

func splitRepo(repo string) (string, string, error) {
	partes := strings.SplitN(repo, "/", 2)
	if len(partes) != 2 || partes[0] == "" || partes[1] == "" {
		return "", "", fmt.Errorf("repositório inválido: %q", repo)
	}
	return partes[0], partes[1], nil
}
