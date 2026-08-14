package app

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

func (m *Model) currentItemList() *list[itemRow] {
	switch m.tab {
	case TabPRs:
		return &m.prs
	case TabRadar:
		return &m.radar
	default:
		return &m.triage
	}
}

func (m *Model) itemRepo(it gh.Item) string {
	if it.Repo != "" {
		return it.Repo
	}
	return m.repo
}

func (m *Model) targetItems() []gh.Item {
	var out []gh.Item
	for _, it := range m.currentItemList().markedItems() {
		out = append(out, it.Item)
	}
	return out
}

func (m *Model) currentItem() (gh.Item, bool) {
	it, ok := m.currentItemList().current()
	return it.Item, ok
}

func (m Model) twoPane() bool {
	return m.width >= 90 && m.height > 14
}

func (m Model) bodyHeight() int {
	h := m.height - 2
	if h < 3 {
		h = 3
	}
	return h
}

func (m Model) listWidth() int {
	if m.twoPane() {
		w := m.width * 62 / 100
		if w < 40 {
			w = 40
		}
		return w
	}
	return m.width
}

func (m Model) sideWidth() int {
	if m.twoPane() {
		return m.width - m.listWidth()
	}
	return 0
}

func (m *Model) resize() {
	if m.twoPane() {
		m.preview.Width = m.sideWidth() - 2
		m.preview.Height = m.bodyHeight() - 2
	} else {
		m.preview.Width = m.width - 2
		m.preview.Height = 3
	}
	if m.preview.Width < 10 {
		m.preview.Width = 10
	}
	if m.preview.Height < 1 {
		m.preview.Height = 1
	}
	m.detail.Width = m.width - 2
	altDet := m.height - 4
	if altDet < 3 {
		altDet = 3
	}
	m.detail.Height = altDet
	m.search.Width = m.width - 6
	m.input.Width = 50
	m.syncPreview()
	if m.mode == modeDetail {
		m.fillDetail()
	}
}

func (m *Model) syncPreview() {
	m.preview.SetContent(m.previewContent())
}

func (m *Model) fillDetail() {
	m.detail.SetContent(m.detailContent())
	m.detail.GotoTop()
}

func (m *Model) switchTab(a Tab) tea.Cmd {
	if m.tab == a {
		return nil
	}
	m.tab = a
	m.search.SetValue("")
	m.search.Blur()
	m.mode = modeTable
	loaded := false
	switch a {
	case TabRadar:
		loaded = m.radar.loaded
	case TabCommits:
		loaded = m.commits.loaded
	case TabPRs:
		loaded = m.prs.loaded
	default:
		loaded = m.triage.loaded
	}
	m.syncPreview()
	if loaded {
		return nil
	}
	m.loading = true
	return tea.Batch(m.spin.Tick, m.loadCmd())
}

func (m *Model) checkPagination() tea.Cmd {
	switch m.tab {
	case TabCommits:
		if !m.commits.shouldPaginate() {
			return nil
		}
		m.commits.paginating = true
		return tea.Batch(m.spin.Tick, fetchCommitsCmd(m.client, m.repo, m.commits.nextCursor, true))
	default:
		l := m.currentItemList()
		if !l.shouldPaginate() {
			return nil
		}
		cursor := l.nextCursor
		l.paginating = true
		return tea.Batch(m.spin.Tick, fetchItemsCmd(m.client, m.tab, m.tabQuery(), cursor, true))
	}
}

func (m *Model) totalVisible() int {
	if m.tab == TabCommits {
		return len(m.commits.visible)
	}
	return len(m.currentItemList().visible)
}

func (m *Model) totalSelected() int {
	if m.tab == TabCommits {
		return len(m.commits.selected)
	}
	return len(m.currentItemList().selected)
}

func (m *Model) totalResults() int {
	if m.tab == TabCommits {
		return len(m.commits.items)
	}
	return m.currentItemList().total
}

func (m *Model) currentURL() string {
	if m.tab == TabCommits {
		if c, ok := m.commits.current(); ok {
			return c.URL
		}
		return ""
	}
	if it, ok := m.currentItem(); ok {
		return it.URL
	}
	return ""
}

func (m *Model) knownLabels() []string {
	visto := map[string]bool{}
	for _, it := range m.triage.items {
		for _, l := range it.Labels {
			visto[l.Name] = true
		}
	}
	for _, it := range m.prs.items {
		for _, l := range it.Labels {
			visto[l.Name] = true
		}
	}
	return sorted(visto)
}

func (m *Model) knownMilestones() []string {
	visto := map[string]bool{}
	for _, it := range m.triage.items {
		if it.Milestone != "" {
			visto[it.Milestone] = true
		}
	}
	for _, it := range m.prs.items {
		if it.Milestone != "" {
			visto[it.Milestone] = true
		}
	}
	return sorted(visto)
}

func sorted(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (m *Model) rateLimitLow() bool {
	return m.rateLimit > 0 && m.rateLimit < 100
}
