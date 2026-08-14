package gh

import "strings"

const searchQuery = `
query($q: String!, $cursor: String) {
  rateLimit { remaining }
  search(query: $q, type: ISSUE, first: 50, after: $cursor) {
    issueCount
    pageInfo { hasNextPage endCursor }
    nodes {
      __typename
      ... on Issue {
        number title state createdAt updatedAt body url
        repository { nameWithOwner }
        author { login }
        labels(first: 10) { nodes { name color } }
        assignees(first: 5) { nodes { login } }
        milestone { title }
        comments { totalCount }
        timelineItems(itemTypes: [CROSS_REFERENCED_EVENT, CONNECTED_EVENT], first: 10) {
          nodes {
            __typename
            ... on CrossReferencedEvent {
              source { __typename ... on PullRequest { number state isDraft repository { nameWithOwner } } }
            }
            ... on ConnectedEvent {
              subject { __typename ... on PullRequest { number state isDraft repository { nameWithOwner } } }
            }
          }
        }
      }
      ... on PullRequest {
        number title state isDraft createdAt updatedAt body url
        repository { nameWithOwner }
        author { login }
        labels(first: 10) { nodes { name color } }
        assignees(first: 5) { nodes { login } }
        milestone { title }
        comments { totalCount }
        reviewDecision
        statusCheckRollup: commits(last: 1) {
          nodes { commit { statusCheckRollup { state } } }
        }
      }
    }
  }
}`

type Filters struct {
	Repo      string
	State     string
	Kind      string
	Labels    []string
	Assignee  string
	Milestone string
	Term      string
}

func OrgScope(repo string) bool {
	return repo != "" && !strings.Contains(repo, "/")
}

func BuildQuery(f Filters) string {
	qualificador := "repo:"
	if OrgScope(f.Repo) {
		qualificador = "org:"
	}
	partes := []string{qualificador + f.Repo}

	switch f.State {
	case "closed":
		partes = append(partes, "is:closed")
	case "all":
	default:
		partes = append(partes, "is:open")
	}

	switch f.Kind {
	case "issue":
		partes = append(partes, "is:issue")
	case "pr":
		partes = append(partes, "is:pr")
	}

	for _, l := range f.Labels {
		partes = append(partes, "label:"+quote(l))
	}
	if f.Assignee != "" {
		partes = append(partes, "assignee:"+quote(f.Assignee))
	}
	if f.Milestone != "" {
		partes = append(partes, "milestone:"+quote(f.Milestone))
	}
	if term := strings.TrimSpace(f.Term); term != "" {
		partes = append(partes, term)
	}

	return strings.Join(partes, " ")
}

func quote(v string) string {
	if strings.ContainsAny(v, " \"") {
		return "\"" + strings.ReplaceAll(v, "\"", "\\\"") + "\""
	}
	return v
}

type searchResponse struct {
	RateLimit *struct {
		Remaining int `json:"remaining"`
	} `json:"rateLimit"`
	Search struct {
		IssueCount int `json:"issueCount"`
		PageInfo   struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []searchNode `json:"nodes"`
	} `json:"search"`
}

type searchNode struct {
	Typename  string `json:"__typename"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	IsDraft   bool   `json:"isDraft"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Body      string `json:"body"`
	URL       string `json:"url"`
	Author    *actor `json:"author"`
	Milestone *struct {
		Title string `json:"title"`
	} `json:"milestone"`
	Repository *struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Labels struct {
		Nodes []struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"nodes"`
	} `json:"labels"`
	Assignees struct {
		Nodes []actor `json:"nodes"`
	} `json:"assignees"`
	Comments struct {
		TotalCount int `json:"totalCount"`
	} `json:"comments"`
	ReviewDecision    string `json:"reviewDecision"`
	StatusCheckRollup struct {
		Nodes []struct {
			Commit *struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"statusCheckRollup"`
	TimelineItems struct {
		Nodes []timelineNode `json:"nodes"`
	} `json:"timelineItems"`
}

type actor struct {
	Login string `json:"login"`
}

type timelineNode struct {
	Typename string     `json:"__typename"`
	Source   *reference `json:"source"`
	Subject  *reference `json:"subject"`
}

type reference struct {
	Typename   string `json:"__typename"`
	Number     int    `json:"number"`
	State      string `json:"state"`
	IsDraft    bool   `json:"isDraft"`
	Repository *struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}
