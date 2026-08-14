package gh

import (
	"regexp"
	"strconv"
)

func numberRegexp() *regexp.Regexp { return regexp.MustCompile(`#(\d+)`) }

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

type Kind int

const (
	KindIssue Kind = iota
	KindPR
)

func (t Kind) String() string {
	if t == KindPR {
		return "pr"
	}
	return "issue"
}

type State int

const (
	StateOpen State = iota
	StateClosed
	StateMerged
)

func (e State) String() string {
	switch e {
	case StateMerged:
		return "merged"
	case StateClosed:
		return "closed"
	default:
		return "open"
	}
}

func parseState(bruto string) State {
	switch bruto {
	case "MERGED":
		return StateMerged
	case "CLOSED":
		return StateClosed
	default:
		return StateOpen
	}
}

type CIStatus int

const (
	CINone CIStatus = iota
	CISuccess
	CIFail
	CIPending
)

func parseCI(bruto string) CIStatus {
	switch bruto {
	case "SUCCESS":
		return CISuccess
	case "FAILURE", "ERROR":
		return CIFail
	case "PENDING", "EXPECTED":
		return CIPending
	default:
		return CINone
	}
}

func (c CIStatus) String() string {
	switch c {
	case CISuccess:
		return "success"
	case CIFail:
		return "failure"
	case CIPending:
		return "pending"
	default:
		return ""
	}
}

type Label struct {
	Name  string
	Color string
}

type LinkedPR struct {
	Number       int
	State        State
	Draft        bool
	ExternalRepo string
}

func (p LinkedPR) Reference() string {
	if p.ExternalRepo != "" {
		return p.ExternalRepo + "#" + strconv.Itoa(p.Number)
	}
	return "#" + strconv.Itoa(p.Number)
}

type Item struct {
	Number         int
	Repo           string
	Kind           Kind
	State          State
	Draft          bool
	Title          string
	Body           string
	URL            string
	Author         string
	Labels         []Label
	Assignees      []string
	Milestone      string
	Comments       int
	CreatedAt      string
	UpdatedAt      string
	ReviewDecision string
	CI             CIStatus
	Linked         []LinkedPR
}

func (i Item) Key() string {
	base := i.Kind.String() + "#" + strconv.Itoa(i.Number)
	if i.Repo != "" {
		return i.Repo + " " + base
	}
	return base
}

type Page struct {
	Total              int
	Items              []Item
	NextCursor         string
	RateLimitRemaining int
}

type Commit struct {
	OID         string
	Title       string
	Body        string
	Author      string
	Date        string
	URL         string
	PRs         []int
	CitedIssues []int
	CI          CIStatus
	Additions   int
	Deletions   int
}

func (c Commit) ShortOID() string {
	if len(c.OID) < 7 {
		return c.OID
	}
	return c.OID[:7]
}

func (c Commit) Key() string { return c.OID }

type CommitsPage struct {
	Commits            []Commit
	NextCursor         string
	RateLimitRemaining int
}

type Comment struct {
	Author string
	Body   string
	Date   string
}

type Review struct {
	Author string
	State  string
	Body   string
	Date   string
}

type Detail struct {
	Number        int
	Title         string
	Body          string
	Author        string
	TotalComments int
	Comments      []Comment
	Reviews       []Review
}

var numberPattern = numberRegexp()

func CitedNumbers(msg string) []int {
	achados := numberPattern.FindAllStringSubmatch(msg, -1)
	var out []int
	visto := map[int]bool{}
	for _, m := range achados {
		n := atoi(m[1])
		if n > 0 && !visto[n] {
			visto[n] = true
			out = append(out, n)
		}
	}
	return out
}
