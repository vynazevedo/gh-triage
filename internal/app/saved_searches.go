package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) openSaveSearch() (tea.Model, tea.Cmd) {
	m.input.SetValue("")
	m.input.Focus()
	m.activeField = fieldSearchName
	m.mode = modeModalText
	return m, nil
}

func (m Model) savedSearchesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "S":
		m.mode = modeTable
		return m, nil
	case "j", "down":
		if m.savedCursor < len(m.savedSearches)-1 {
			m.savedCursor++
		}
		return m, nil
	case "k", "up":
		if m.savedCursor > 0 {
			m.savedCursor--
		}
		return m, nil
	case "d":
		if m.savedCursor < len(m.savedSearches) {
			m.savedSearches = append(m.savedSearches[:m.savedCursor], m.savedSearches[m.savedCursor+1:]...)
			_ = saveSearches(m.savedSearches)
			if m.savedCursor >= len(m.savedSearches) && m.savedCursor > 0 {
				m.savedCursor--
			}
		}
		return m, nil
	case "enter":
		if m.savedCursor < len(m.savedSearches) {
			m.filters = m.savedSearches[m.savedCursor].Filters
			m.mode = modeTable
			m.triage.loaded = false
			m.prs.loaded = false
			m.loading = true
			return m, tea.Batch(m.spin.Tick, m.loadCmd())
		}
		return m, nil
	}
	return m, nil
}
