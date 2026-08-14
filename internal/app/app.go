package app

import (
	blist "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/vynazevedo/gh-triage/internal/cache"
	"github.com/vynazevedo/gh-triage/internal/gh"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

type Tab int

const (
	TabRadar Tab = iota
	TabTriage
	TabPRs
	TabCommits
)

func (a Tab) title() string {
	switch a {
	case TabRadar:
		return "Radar"
	case TabPRs:
		return "PRs"
	case TabCommits:
		return "Commits"
	default:
		return "Triagem"
	}
}

var tabs = []Tab{TabRadar, TabTriage, TabPRs, TabCommits}

type detailRef struct {
	Repo   string
	Number int
}

func (r detailRef) key() string {
	return r.Repo + "#" + itoa(r.Number)
}

type mode int

const (
	modeTable mode = iota
	modeSearch
	modeDetail
	modeHelp
	modeModal
	modeModalText
	modeSavedSearches
	modeFilter
	modePicker
)

type toast struct {
	text  string
	isErr bool
}

type Model struct {
	client  *gh.Client
	repo    string
	filters gh.Filters

	tab     Tab
	mode    mode
	radar   list[itemRow]
	triage  list[itemRow]
	prs     list[itemRow]
	commits list[commitRow]

	width, height int
	search        textinput.Model
	preview       viewport.Model
	detail        viewport.Model
	spin          spinner.Model
	input         textinput.Model

	modal        *modal
	activeField  textField
	details      map[string]gh.Detail
	openedDetail detailRef

	currentToast *toast
	loading      bool
	rateLimit    int

	savedSearches []SavedSearch
	savedCursor   int
	filterCursor  int

	picker      blist.Model
	pickerStage int
	myLogin     string

	watch        bool
	watchSeconds int

	quit bool
}

func (m Model) WithWatch(seconds int) Model {
	m.watch = true
	m.watchSeconds = seconds
	return m
}

func New(client *gh.Client, repo string, filters gh.Filters, abaInicial Tab) Model {
	search := textinput.New()
	search.Prompt = "/"
	search.Placeholder = "buscar…"

	input := textinput.New()
	input.Prompt = "> "

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = ui.StyleTitle

	m := Model{
		client:        client,
		repo:          repo,
		filters:       filters,
		tab:           abaInicial,
		mode:          modeTable,
		radar:         newList[itemRow](),
		triage:        newList[itemRow](),
		prs:           newList[itemRow](),
		commits:       newList[commitRow](),
		search:        search,
		input:         input,
		spin:          sp,
		details:       map[string]gh.Detail{},
		savedSearches: loadSearches(),
	}
	if abaInicial == TabRadar {
		m.seedRadarFromCache()
	}
	return m
}

func (m *Model) seedRadarFromCache() {
	snap, ok := cache.LoadRadar()
	if !ok || len(snap.Items) == 0 {
		return
	}
	linhas := make([]itemRow, len(snap.Items))
	for i, it := range snap.Items {
		linhas[i] = itemRow{Item: it.Item, Reason: it.Reason, Class: it.Class}
	}
	m.radar.replace(linhas, len(linhas), "", "")
}

func (m Model) Init() tea.Cmd {
	m.loading = true
	cmds := []tea.Cmd{m.spin.Tick, m.loadCmd()}
	if m.watch {
		cmds = append(cmds, radarTickCmd(m.watchSeconds))
	}
	return tea.Batch(cmds...)
}

type itemsMsg struct {
	tab         Tab
	appendItems bool
	page        gh.Page
	err         error
}

type commitsMsg struct {
	appendItems bool
	page        gh.CommitsPage
	err         error
}

type detailMsg struct {
	ref    detailRef
	detail gh.Detail
	err    error
}

type radarMsg struct {
	items []gh.RadarItem
	rate  int
	err   error
}

type itemResult struct {
	target string
	err    string
}

type batchResult struct {
	results []itemResult
}

type editorDone struct {
	target  editorTarget
	content string
	err     error
}

type checkoutDone struct {
	number int
	err    error
}

func (m *Model) tabQuery() string {
	f := m.filters
	if m.tab == TabPRs {
		f.Kind = "pr"
	}
	return gh.BuildQuery(f)
}

func (m Model) loadCmd() tea.Cmd {
	switch m.tab {
	case TabRadar:
		return fetchRadarCmd(m.client)
	case TabCommits:
		if m.repo == "" || gh.OrgScope(m.repo) {
			return func() tea.Msg {
				return commitsMsg{err: errOrgScope}
			}
		}
		return fetchCommitsCmd(m.client, m.repo, "", false)
	default:
		if m.repo == "" {
			tab := m.tab
			return func() tea.Msg {
				return itemsMsg{tab: tab, err: errNoRepo}
			}
		}
		return fetchItemsCmd(m.client, m.tab, m.tabQuery(), "", false)
	}
}

func fetchRadarCmd(c *gh.Client) tea.Cmd {
	return func() tea.Msg {
		items, rate, err := c.FetchRadar()
		return radarMsg{items: items, rate: rate, err: err}
	}
}

var (
	errOrgScope = scopeErr("a aba Commits precisa de um repositório específico; use . para trocar")
	errNoRepo   = scopeErr("nenhum repositório definido — use . para escolher")
)

type scopeErr string

func (e scopeErr) Error() string { return string(e) }

func fetchItemsCmd(c *gh.Client, tab Tab, search, cursor string, appendItems bool) tea.Cmd {
	return func() tea.Msg {
		page, err := c.Search(search, cursor)
		return itemsMsg{tab: tab, appendItems: appendItems, page: page, err: err}
	}
}

func fetchCommitsCmd(c *gh.Client, repo, cursor string, appendItems bool) tea.Cmd {
	return func() tea.Msg {
		page, err := c.FetchCommits(repo, cursor)
		return commitsMsg{appendItems: appendItems, page: page, err: err}
	}
}

func fetchDetailCmd(c *gh.Client, ref detailRef) tea.Cmd {
	return func() tea.Msg {
		det, err := c.FetchDetail(ref.Repo, ref.Number)
		return detailMsg{ref: ref, detail: det, err: err}
	}
}

func applyCmd(c *gh.Client, repo string, muts []gh.Mutation) tea.Cmd {
	return func() tea.Msg {
		res := applyBatch(c, repo, muts)
		return batchResult{results: res}
	}
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		b := browser.New("", nil, nil)
		_ = b.Browse(url)
		return nil
	}
}

func applyBatch(c *gh.Client, repo string, muts []gh.Mutation) []itemResult {
	out := make([]itemResult, 0, len(muts))
	sem := make(chan struct{}, 5)
	res := make([]itemResult, len(muts))
	done := make(chan int, len(muts))
	for i, mut := range muts {
		go func(i int, mut gh.Mutation) {
			sem <- struct{}{}
			defer func() { <-sem }()
			target := "novo"
			if mut.Number > 0 {
				target = "#" + itoa(mut.Number)
			}
			r := itemResult{target: target}
			if err := c.Apply(repo, mut); err != nil {
				r.err = err.Error()
			}
			res[i] = r
			done <- i
		}(i, mut)
	}
	for range muts {
		<-done
	}
	out = append(out, res...)
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(b[pos:])
}
