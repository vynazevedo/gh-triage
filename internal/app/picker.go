package app

import (
	blist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vynazevedo/gh-triage/internal/gh"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

type pickerKind int

const (
	pickOrg pickerKind = iota
	pickRepo
	pickMyRepos
	pickWholeOrg
)

type pickerItem struct {
	title string
	desc  string
	value string
	kind  pickerKind
}

func (p pickerItem) Title() string       { return p.title }
func (p pickerItem) Description() string { return p.desc }
func (p pickerItem) FilterValue() string { return p.title + " " + p.desc }

const (
	pickerStageOrgs = iota
	pickerStageRepos
)

type orgsMsg struct {
	login  string
	owners []gh.Owner
	err    error
}

type reposMsg struct {
	org   string
	repos []gh.RepoRef
	err   error
}

func fetchOrgsCmd(c *gh.Client) tea.Cmd {
	return func() tea.Msg {
		login, _ := c.MyLogin()
		owners, err := c.MyOrgs()
		return orgsMsg{login: login, owners: owners, err: err}
	}
}

func fetchReposCmd(c *gh.Client, org, myLogin string) tea.Cmd {
	return func() tea.Msg {
		var repos []gh.RepoRef
		var err error
		if org == myLogin || org == "" {
			repos, err = c.MyRepos()
		} else {
			repos, err = c.OrgRepos(org)
		}
		return reposMsg{org: org, repos: repos, err: err}
	}
}

func newPickerList(title string, w, h int) blist.Model {
	d := blist.NewDefaultDelegate()
	l := blist.New(nil, d, w, h)
	l.Title = title
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.Styles.Title = ui.StyleTitle
	return l
}

func (m Model) openPicker() (tea.Model, tea.Cmd) {
	w, h := m.pickerSize()
	m.picker = newPickerList("organizações", w, h)
	m.picker.SetItems([]blist.Item{pickerItem{title: "carregando…", kind: pickOrg}})
	m.pickerStage = pickerStageOrgs
	m.mode = modePicker
	m.loading = true
	return m, tea.Batch(m.spin.Tick, fetchOrgsCmd(m.client))
}

func (m Model) pickerSize() (int, int) {
	w := m.width * 70 / 100
	if w < 40 {
		w = 40
	}
	if w > m.width-4 {
		w = m.width - 4
	}
	h := m.height - 6
	if h < 6 {
		h = 6
	}
	return w, h
}

func (m Model) onOrgs(msg orgsMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		m.mode = modeTable
		return m, nil
	}
	m.myLogin = msg.login
	items := []blist.Item{
		pickerItem{title: "Meus repositórios", desc: "@" + orDash(msg.login), value: msg.login, kind: pickMyRepos},
	}
	for _, o := range msg.owners {
		items = append(items, pickerItem{title: o.Login, desc: "organização", value: o.Login, kind: pickOrg})
	}
	m.picker.SetItems(items)
	m.picker.Title = "escolha org ou @me"
	return m, nil
}

func (m Model) onRepos(msg reposMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.currentToast = &toast{text: msg.err.Error(), isErr: true}
		return m, nil
	}
	var items []blist.Item
	if msg.org != "" && msg.org != m.myLogin {
		items = append(items, pickerItem{
			title: "▸ Toda a organização " + msg.org,
			desc:  "org:" + msg.org + " — todos os repos de uma vez",
			value: msg.org,
			kind:  pickWholeOrg,
		})
	}
	for _, r := range msg.repos {
		desc := r.Description
		if desc == "" {
			desc = orDash(r.Owner)
		}
		items = append(items, pickerItem{title: r.FullName(), desc: desc, value: r.FullName(), kind: pickRepo})
	}
	w, h := m.pickerSize()
	m.picker = newPickerList("repositórios de "+msg.org, w, h)
	m.picker.SetItems(items)
	m.pickerStage = pickerStageRepos
	return m, nil
}

func (m Model) pickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.pickerStage == pickerStageRepos {
			m.loading = true
			return m.openPicker()
		}
		m.mode = modeTable
		return m, nil
	case "enter":
		sel, ok := m.picker.SelectedItem().(pickerItem)
		if !ok {
			return m, nil
		}
		switch sel.kind {
		case pickMyRepos, pickOrg:
			m.loading = true
			m.picker.SetItems([]blist.Item{pickerItem{title: "carregando…"}})
			return m, tea.Batch(m.spin.Tick, fetchReposCmd(m.client, sel.value, m.myLogin))
		case pickRepo, pickWholeOrg:
			return m.applyScope(sel.value)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

func (m Model) applyScope(repo string) (tea.Model, tea.Cmd) {
	m.filters.Repo = repo
	m.repo = repo
	m.triage = newList[itemRow]()
	m.prs = newList[itemRow]()
	m.commits = newList[commitRow]()
	m.radar = newList[itemRow]()
	m.details = map[string]gh.Detail{}
	m.mode = modeTable
	if gh.OrgScope(repo) {
		m.tab = TabTriage
	}
	m.loading = true
	return m, tea.Batch(m.spin.Tick, m.loadCmd())
}

func (m Model) viewPicker() string {
	hint := hotkeys([][2]string{
		{"/", "filtrar"}, {"↑↓", "navegar"}, {"enter", "escolher"}, {"esc", "voltar"},
	})
	return m.picker.View() + "\n " + hint
}
