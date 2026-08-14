package gh

import "fmt"

const detalheQuery = `
query($owner: String!, $name: String!, $numero: Int!) {
  repository(owner: $owner, name: $name) {
    issueOrPullRequest(number: $numero) {
      __typename
      ... on Issue {
        number title body
        author { login }
        comments(last: 30) { totalCount nodes { author { login } body createdAt } }
      }
      ... on PullRequest {
        number title body
        author { login }
        comments(last: 30) { totalCount nodes { author { login } body createdAt } }
        reviews(last: 20) { nodes { author { login } state body submittedAt } }
      }
    }
  }
}`

type detalheResponse struct {
	Repository struct {
		IssueOrPullRequest struct {
			Typename string `json:"__typename"`
			Number   int    `json:"number"`
			Title    string `json:"title"`
			Body     string `json:"body"`
			Author   *actor `json:"author"`
			Comments struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					Author    *actor `json:"author"`
					Body      string `json:"body"`
					CreatedAt string `json:"createdAt"`
				} `json:"nodes"`
			} `json:"comments"`
			Reviews struct {
				Nodes []struct {
					Author      *actor `json:"author"`
					State       string `json:"state"`
					Body        string `json:"body"`
					SubmittedAt string `json:"submittedAt"`
				} `json:"nodes"`
			} `json:"reviews"`
		} `json:"issueOrPullRequest"`
	} `json:"repository"`
}

func (c *Client) FetchDetail(repo string, number int) (Detail, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return Detail{}, err
	}
	vars := map[string]any{"owner": owner, "name": name, "numero": number}

	var resp detalheResponse
	if err := c.gql.Do(detalheQuery, vars, &resp); err != nil {
		return Detail{}, fmt.Errorf("falha ao carregar #%d de %q: %w", number, repo, err)
	}

	no := resp.Repository.IssueOrPullRequest
	if no.Number == 0 {
		return Detail{}, fmt.Errorf("#%d não encontrado em %q", number, repo)
	}

	comentarios := make([]Comment, 0, len(no.Comments.Nodes))
	for _, cm := range no.Comments.Nodes {
		comentarios = append(comentarios, Comment{
			Author: login(cm.Author),
			Body:   cm.Body,
			Date:   cm.CreatedAt,
		})
	}
	var reviews []Review
	for _, r := range no.Reviews.Nodes {
		if r.Body == "" && r.State == "COMMENTED" {
			continue
		}
		reviews = append(reviews, Review{
			Author: login(r.Author),
			State:  r.State,
			Body:   r.Body,
			Date:   r.SubmittedAt,
		})
	}

	return Detail{
		Number:        no.Number,
		Title:         no.Title,
		Body:          no.Body,
		Author:        login(no.Author),
		TotalComments: no.Comments.TotalCount,
		Comments:      comentarios,
		Reviews:       reviews,
	}, nil
}
