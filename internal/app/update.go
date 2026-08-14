package app

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vynazevedo/gh-triage/internal/cache"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

type radarTickMsg struct{}

func radarTickCmd(seconds int) tea.Cmd {
	return tea.Tick(time.Duration(seconds)*time.Second, func(time.Time) tea.Msg {
		return radarTickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		if m.loading || m.currentPaginating() {
			return m, cmd
		}
		return m, nil

	case itemsMsg:
		return m.onItems(msg)

	case commitsMsg:
		return m.onCommits(msg)

	case detailMsg:
		m.loading = false
		if msg.err != nil {
			m.currentToast = &toast{text: msg.err.Error(), isErr: true}
			return m, nil
		}
		m.details[msg.ref.key()] = msg.detail
		m.currentToast = nil
		if m.mode == modeDetail && m.openedDetail == msg.ref {
			m.fillDetail()
		}
		return m, nil

	case radarMsg:
		return m.onRadar(msg)

	case radarTickMsg:
		var cmd tea.Cmd
		if m.tab == TabRadar && m.mode == modeTable && !m.loading {
			cmd = fetchRadarCmd(m.client)
		}
		return m, tea.Batch(cmd, radarTickCmd(m.watchSeconds))

	case orgsMsg:
		return m.onOrgs(msg)

	case reposMsg:
		return m.onRepos(msg)

	case batchResult:
		return m.onBatch(msg)

	case editorDone:
		return m.onEditor(msg)

	case checkoutDone:
		m.loading = false
		if msg.err != nil {
			m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		} else {
			m.currentToast = &toast{text: "checkout do PR #" + itoa(msg.number) + " concluído"}
		}
		return m, nil

	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return m, nil
}

func (m *Model) currentPaginating() bool {
	switch m.tab {
	case TabCommits:
		return m.commits.paginating
	case TabPRs:
		return m.prs.paginating
	case TabRadar:
		return m.radar.paginating
	default:
		return m.triage.paginating
	}
}

func (m Model) onRadar(msg radarMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil && len(msg.items) == 0 {
		m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		return m, nil
	}
	linhas := make([]itemRow, len(msg.items))
	for i, it := range msg.items {
		linhas[i] = itemRow{Item: it.Item, Reason: it.Reason, Class: it.Class}
	}
	m.radar.replace(linhas, len(linhas), "", m.search.Value())
	if msg.rate > 0 {
		m.rateLimit = msg.rate
	}
	if msg.err != nil {
		m.currentToast = &toast{text: "radar parcial: " + msg.err.Error(), isErr: true}
	} else {
		m.currentToast = nil
		_ = cache.SaveRadar(msg.items, msg.rate)
	}
	m.syncPreview()
	return m, nil
}

func (m Model) onItems(msg itemsMsg) (tea.Model, tea.Cmd) {
	if !msg.appendItems {
		m.loading = false
	}
	l := &m.triage
	if msg.tab == TabPRs {
		l = &m.prs
	}
	if msg.err != nil {
		l.paginating = false
		m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		return m, nil
	}
	items := wrapItems(msg.page.Items)
	if msg.appendItems {
		l.appendItems(items, msg.page.NextCursor, m.search.Value())
	} else {
		l.replace(items, msg.page.Total, msg.page.NextCursor, m.search.Value())
	}
	m.rateLimit = msg.page.RateLimitRemaining
	m.currentToast = nil
	m.syncPreview()
	return m, nil
}

func (m Model) onCommits(msg commitsMsg) (tea.Model, tea.Cmd) {
	if !msg.appendItems {
		m.loading = false
	}
	if msg.err != nil {
		m.commits.paginating = false
		m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		return m, nil
	}
	cs := wrapCommits(msg.page.Commits)
	if msg.appendItems {
		m.commits.appendItems(cs, msg.page.NextCursor, m.search.Value())
	} else {
		m.commits.replace(cs, len(cs), msg.page.NextCursor, m.search.Value())
	}
	m.rateLimit = msg.page.RateLimitRemaining
	m.currentToast = nil
	m.syncPreview()
	return m, nil
}

func (m Model) onBatch(msg batchResult) (tea.Model, tea.Cmd) {
	m.loading = false
	var falhas []itemResult
	for _, r := range msg.results {
		if r.err != "" {
			falhas = append(falhas, r)
		}
	}
	ok := len(msg.results) - len(falhas)
	if len(falhas) == 0 {
		m.currentToast = &toast{text: itoa(ok) + " ok"}
	} else {
		det := ""
		for i, f := range falhas {
			if i > 0 {
				det += " · "
			}
			det += f.target + ": " + f.err
		}
		m.currentToast = &toast{text: itoa(ok) + " ok, " + itoa(len(falhas)) + " falhou: " + det, isErr: true}
	}
	m.clearCurrentSelect()
	m.loading = true
	cmds := []tea.Cmd{m.loadCmd()}
	if m.mode == modeDetail && m.openedDetail.Number > 0 {
		delete(m.details, m.openedDetail.key())
		cmds = append(cmds, fetchDetailCmd(m.client, m.openedDetail))
	}
	return m, tea.Batch(cmds...)
}

func (m Model) onEditor(msg editorDone) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		return m, nil
	}
	if msg.content == "" {
		m.currentToast = &toast{text: "edição cancelada"}
		return m, nil
	}
	mut, ok := editorMutation(msg.target, msg.content)
	if !ok {
		return m, nil
	}
	m.loading = true
	return m, applyCmd(m.client, m.repo, []gh.Mutation{mut})
}

func wrapItems(items []gh.Item) []itemRow {
	out := make([]itemRow, len(items))
	for i, it := range items {
		out[i] = itemRow{Item: it}
	}
	return out
}

func wrapCommits(cs []gh.Commit) []commitRow {
	out := make([]commitRow, len(cs))
	for i, c := range cs {
		out[i] = commitRow{c}
	}
	return out
}

func (m *Model) clearCurrentSelect() {
	switch m.tab {
	case TabRadar:
		m.radar.selected = map[string]bool{}
	case TabCommits:
		m.commits.selected = map[string]bool{}
	case TabPRs:
		m.prs.selected = map[string]bool{}
	default:
		m.triage.selected = map[string]bool{}
	}
}
