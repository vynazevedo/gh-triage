package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

func (m Model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.quit = true
		return m, tea.Quit
	}

	switch m.mode {
	case modeSearch:
		return m.searchKey(msg)
	case modeDetail:
		return m.detailKey(msg)
	case modeHelp:
		m.mode = modeTable
		return m, nil
	case modeModal:
		return m.modalKey(msg)
	case modeModalText:
		return m.textKey(msg)
	case modeSavedSearches:
		return m.savedSearchesKey(msg)
	case modeFilter:
		return m.filterKey(msg)
	case modePicker:
		return m.pickerKey(msg)
	}

	switch msg.String() {
	case "tab", "l", "right":
		return m, m.switchTab(nextTab(m.tab, 1))
	case "shift+tab", "h", "left":
		return m, m.switchTab(nextTab(m.tab, -1))
	case "1":
		return m, m.switchTab(TabRadar)
	case "2":
		return m, m.switchTab(TabTriage)
	case "3":
		return m, m.switchTab(TabPRs)
	case "4":
		return m, m.switchTab(TabCommits)
	}

	return m.tableKey(msg)
}

func nextTab(a Tab, delta int) Tab {
	i := (int(a) + delta + len(tabs)) % len(tabs)
	return tabs[i]
}

func (m Model) tableKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.moveCurrent(1)
		return m, m.checkPagination()
	case "k", "up":
		m.moveCurrent(-1)
		return m, m.checkPagination()
	case "ctrl+d":
		m.moveCurrent(10)
		return m, m.checkPagination()
	case "ctrl+u":
		m.moveCurrent(-10)
		return m, nil
	case "g":
		m.gotoTop()
		return m, nil
	case "G":
		m.gotoEnd()
		return m, m.checkPagination()
	case " ":
		m.toggleCurrentSelect()
		return m, nil
	case "A":
		m.toggleCurrentAll()
		return m, nil
	case "enter":
		return m.openDetail()
	case "/":
		m.mode = modeSearch
		m.search.Focus()
		return m, nil
	case "f":
		m.mode = modeFilter
		m.filterCursor = 0
		return m, nil
	case ".":
		return m.openPicker()
	case ">":
		m.input.SetValue(m.repo)
		m.input.Focus()
		m.activeField = fieldRepo
		m.mode = modeModalText
		return m, nil
	case "S":
		m.mode = modeSavedSearches
		m.savedCursor = 0
		return m, nil
	case "w":
		return m.openSaveSearch()
	case "?":
		m.mode = modeHelp
		return m, nil
	case "o":
		if u := m.currentURL(); u != "" {
			return m, openURLCmd(u)
		}
		return m, nil
	case "r":
		m.loading = true
		return m, tea.Batch(m.spin.Tick, m.loadCmd())
	case "q", "esc":
		m.quit = true
		return m, tea.Quit
	}
	return m.actionKey(msg)
}

func (m Model) actionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "L":
		options := m.knownLabels()
		if len(options) == 0 {
			m.currentToast = &toast{text: "nenhuma label conhecida nos itens carregados"}
			return m, nil
		}
		m.openSelect(selLabel, options)
		return m, nil
	case "a":
		m.openText(fieldAssignee)
		return m, nil
	case "x":
		return m.confirmBatch(gh.MutClose), nil
	case "X":
		return m.confirmBatch(gh.MutReopen), nil
	case "n":
		if it, ok := m.currentItem(); ok {
			return m, editorCmd(editorTarget{kind: destComment, number: it.Number, repo: m.itemRepo(it)}, "")
		}
		return m, nil
	case "c":
		if m.repo == "" {
			m.currentToast = &toast{text: "defina um repositório com . antes de criar issue"}
			return m, nil
		}
		m.openText(fieldNewIssue)
		return m, nil
	case "e":
		if it, ok := m.currentItem(); ok {
			return m, editorCmd(editorTarget{kind: destBody, number: it.Number, title: it.Title, repo: m.itemRepo(it)}, it.Body)
		}
		return m, nil
	case "v":
		if m.currentIsPR() {
			m.openSelect(selReview, []string{"aprovar", "comentar", "pedir mudanças"})
		} else {
			m.currentToast = &toast{text: "review só se aplica a pull requests"}
		}
		return m, nil
	case "M":
		if m.currentIsPR() {
			m.openSelect(selMerge, []string{"merge", "squash", "rebase"})
		} else {
			m.currentToast = &toast{text: "merge só se aplica a pull requests"}
		}
		return m, nil
	case "C":
		if it, ok := m.currentItem(); ok && it.Kind == gh.KindPR {
			if it.Repo != "" && it.Repo != m.repo {
				m.currentToast = &toast{text: "checkout só funciona no repositório atual (" + orDash(m.repo) + ")"}
				return m, nil
			}
			m.loading = true
			return m, checkoutCmd(it.Number)
		}
		m.currentToast = &toast{text: "checkout só se aplica a pull requests"}
		return m, nil
	}
	return m, nil
}

func (m Model) searchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.search.SetValue("")
		m.search.Blur()
		m.mode = modeTable
		m.refilterCurrent()
		return m, nil
	case "enter":
		m.search.Blur()
		m.mode = modeTable
		return m, nil
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.refilterCurrent()
	return m, cmd
}

func (m Model) detailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = modeTable
		m.openedDetail = detailRef{}
		return m, nil
	case "n":
		if m.openedDetail.Number > 0 {
			return m, editorCmd(editorTarget{kind: destComment, number: m.openedDetail.Number, repo: m.openedDetail.Repo}, "")
		}
		return m, nil
	case "o":
		if u := m.currentURL(); u != "" {
			return m, openURLCmd(u)
		}
		return m, nil
	case "r":
		if m.openedDetail.Number > 0 {
			delete(m.details, m.openedDetail.key())
			m.loading = true
			return m, tea.Batch(m.spin.Tick, fetchDetailCmd(m.client, m.openedDetail))
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

func (m Model) openDetail() (tea.Model, tea.Cmd) {
	it, ok := m.currentItem()
	if !ok {
		return m, nil
	}
	m.mode = modeDetail
	ref := detailRef{Repo: m.itemRepo(it), Number: it.Number}
	m.openedDetail = ref
	if _, existe := m.details[ref.key()]; existe {
		m.fillDetail()
		return m, nil
	}
	m.loading = true
	m.fillDetail()
	return m, tea.Batch(m.spin.Tick, fetchDetailCmd(m.client, ref))
}

func (m *Model) moveCurrent(d int) {
	if m.tab == TabCommits {
		m.commits.move(d)
	} else {
		m.currentItemList().move(d)
	}
	m.syncPreview()
}

func (m *Model) gotoTop() {
	if m.tab == TabCommits {
		m.commits.cursor = 0
	} else {
		m.currentItemList().cursor = 0
	}
	m.syncPreview()
}

func (m *Model) gotoEnd() {
	if m.tab == TabCommits {
		if n := len(m.commits.visible); n > 0 {
			m.commits.cursor = n - 1
		}
	} else {
		l := m.currentItemList()
		if n := len(l.visible); n > 0 {
			l.cursor = n - 1
		}
	}
	m.syncPreview()
}

func (m *Model) toggleCurrentSelect() {
	if m.tab == TabCommits {
		m.commits.toggleSelect()
	} else {
		m.currentItemList().toggleSelect()
	}
}

func (m *Model) toggleCurrentAll() {
	if m.tab == TabCommits {
		m.commits.toggleAll()
	} else {
		m.currentItemList().toggleAll()
	}
}

func (m *Model) refilterCurrent() {
	if m.tab == TabCommits {
		m.commits.refilter(m.search.Value())
	} else {
		m.currentItemList().refilter(m.search.Value())
	}
	m.syncPreview()
}

func (m *Model) currentIsPR() bool {
	it, ok := m.currentItem()
	return ok && it.Kind == gh.KindPR
}
