package app

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vynazevedo/gh-triage/internal/gh"
)

type modalKind int

const (
	modalConfirm modalKind = iota
	modalSelect
)

type selectTarget int

const (
	selLabel selectTarget = iota
	selMerge
	selReview
)

type textField int

const (
	fieldAssignee textField = iota
	fieldNewIssue
	fieldSearchName
	fieldRepo
)

type modal struct {
	kind      modalKind
	title     string
	summary   []string
	mutations []gh.Mutation
	target    selectTarget
	options   []string
	cursor    int
}

type destKind int

const (
	destBody destKind = iota
	destComment
	destNewIssue
	destReview
)

type editorTarget struct {
	kind   destKind
	repo   string
	number int
	title  string
	event  string
}

func (m *Model) confirmBatch(t gh.MutationKind) Model {
	items := m.targetItems()
	if len(items) == 0 {
		m.currentToast = &toast{text: "nada selecionado"}
		return *m
	}
	muts := make([]gh.Mutation, 0, len(items))
	alvos := make([]string, 0, len(items))
	for _, it := range items {
		muts = append(muts, gh.Mutation{Kind: t, Number: it.Number, Repo: m.itemRepo(it)})
		alvos = append(alvos, "#"+itoa(it.Number))
	}
	m.openConfirm(muts, alvos)
	return *m
}

func (m *Model) openConfirm(muts []gh.Mutation, alvos []string) {
	if len(muts) == 0 {
		return
	}
	m.modal = &modal{
		kind:  modalConfirm,
		title: "confirmar",
		summary: []string{
			muts[0].Description() + " em " + itoa(len(muts)) + " item(ns)?",
			strings.Join(alvos, " "),
		},
		mutations: muts,
	}
	m.mode = modeModal
}

func (m *Model) openSelect(target selectTarget, options []string) {
	title := map[selectTarget]string{selLabel: "label", selMerge: "método de merge", selReview: "review"}[target]
	m.modal = &modal{kind: modalSelect, title: title, target: target, options: options}
	m.mode = modeModal
}

func (m *Model) openText(campo textField) {
	m.input.SetValue("")
	m.input.Focus()
	m.activeField = campo
	m.mode = modeModalText
}

func (m Model) modalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.modal = nil
		m.mode = modeTable
		return m, nil
	}
	if m.modal == nil {
		m.mode = modeTable
		return m, nil
	}
	switch m.modal.kind {
	case modalConfirm:
		switch msg.String() {
		case "y", "s", "enter":
			muts := m.modal.mutations
			m.modal = nil
			m.mode = modeTable
			m.loading = true
			return m, applyCmd(m.client, m.repo, muts)
		case "n":
			m.modal = nil
			m.mode = modeTable
		}
		return m, nil
	case modalSelect:
		switch msg.String() {
		case "j", "down":
			if m.modal.cursor < len(m.modal.options)-1 {
				m.modal.cursor++
			}
		case "k", "up":
			if m.modal.cursor > 0 {
				m.modal.cursor--
			}
		case "enter", " ":
			return m.finishSelect()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) finishSelect() (tea.Model, tea.Cmd) {
	escolha := ""
	if m.modal.cursor < len(m.modal.options) {
		escolha = m.modal.options[m.modal.cursor]
	}
	target := m.modal.target
	m.modal = nil
	m.mode = modeTable

	switch target {
	case selLabel:
		items := m.targetItems()
		muts := make([]gh.Mutation, 0, len(items))
		alvos := make([]string, 0, len(items))
		for _, it := range items {
			muts = append(muts, gh.Mutation{Kind: gh.MutLabel, Number: it.Number, Repo: m.itemRepo(it), Labels: []string{escolha}})
			alvos = append(alvos, "#"+itoa(it.Number))
		}
		m.openConfirm(muts, alvos)
		return m, nil
	case selMerge:
		it, ok := m.currentItem()
		if !ok {
			return m, nil
		}
		m.openConfirm(
			[]gh.Mutation{{Kind: gh.MutMerge, Number: it.Number, Repo: m.itemRepo(it), Method: escolha}},
			[]string{"#" + itoa(it.Number)},
		)
		return m, nil
	case selReview:
		it, ok := m.currentItem()
		if !ok {
			return m, nil
		}
		event := map[string]string{"aprovar": "APPROVE", "pedir mudanças": "REQUEST_CHANGES"}[escolha]
		if event == "" {
			event = "COMMENT"
		}
		return m, editorCmd(editorTarget{kind: destReview, number: it.Number, repo: m.itemRepo(it), event: event}, "")
	}
	return m, nil
}

func (m Model) textKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.input.Blur()
		m.mode = modeTable
		return m, nil
	case "enter":
		valor := strings.TrimSpace(m.input.Value())
		m.input.Blur()
		m.mode = modeTable
		return m.finishText(valor)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) finishText(valor string) (tea.Model, tea.Cmd) {
	if valor == "" {
		m.currentToast = &toast{text: "valor vazio, nada foi feito"}
		return m, nil
	}
	switch m.activeField {
	case fieldAssignee:
		items := m.targetItems()
		muts := make([]gh.Mutation, 0, len(items))
		alvos := make([]string, 0, len(items))
		for _, it := range items {
			muts = append(muts, gh.Mutation{Kind: gh.MutAssignee, Number: it.Number, Repo: m.itemRepo(it), Logins: []string{valor}})
			alvos = append(alvos, "#"+itoa(it.Number))
		}
		m.openConfirm(muts, alvos)
		return m, nil
	case fieldRepo:
		m.filters.Repo = valor
		m.repo = valor
		m.triage = newList[itemRow]()
		m.prs = newList[itemRow]()
		m.commits = newList[commitRow]()
		m.details = map[string]gh.Detail{}
		m.tab = TabTriage
		m.loading = true
		return m, tea.Batch(m.spin.Tick, m.loadCmd())
	case fieldNewIssue:
		return m, editorCmd(editorTarget{kind: destNewIssue, title: valor}, "")
	case fieldSearchName:
		m.savedSearches = append(m.savedSearches, SavedSearch{Name: valor, Filters: m.filters})
		if err := saveSearches(m.savedSearches); err != nil {
			m.currentToast = &toast{text: "falha ao salvar: " + err.Error(), isErr: true}
		} else {
			m.currentToast = &toast{text: "busca salva: " + valor}
		}
		return m, nil
	}
	return m, nil
}

func editorMutation(d editorTarget, content string) (gh.Mutation, bool) {
	switch d.kind {
	case destBody:
		return gh.Mutation{Kind: gh.MutEdit, Repo: d.repo, Number: d.number, Title: d.title, Body: content}, true
	case destComment:
		return gh.Mutation{Kind: gh.MutComment, Repo: d.repo, Number: d.number, Body: content}, true
	case destNewIssue:
		return gh.Mutation{Kind: gh.MutCreateIssue, Title: d.title, Body: content}, true
	case destReview:
		return gh.Mutation{Kind: gh.MutReview, Repo: d.repo, Number: d.number, Event: d.event, Body: content}, true
	}
	return gh.Mutation{}, false
}

func editorCmd(target editorTarget, inicial string) tea.Cmd {
	arquivo, err := os.CreateTemp("", "gh-triage-*.md")
	if err != nil {
		return func() tea.Msg { return editorDone{target: target, err: err} }
	}
	nome := arquivo.Name()
	_, _ = arquivo.WriteString(inicial)
	_ = arquivo.Close()

	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	partes := strings.Fields(editor)
	partes = append(partes, nome)
	cmd := exec.Command(partes[0], partes[1:]...)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer os.Remove(nome)
		if err != nil {
			return editorDone{target: target, err: err}
		}
		dados, lerErr := os.ReadFile(nome)
		if lerErr != nil {
			return editorDone{target: target, err: lerErr}
		}
		content := strings.TrimSpace(string(dados))
		if content == strings.TrimSpace(inicial) {
			content = ""
		}
		return editorDone{target: target, content: content}
	})
}

func checkoutCmd(number int) tea.Cmd {
	cmd := exec.Command("gh", "pr", "checkout", itoa(number))
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return checkoutDone{number: number, err: err}
	})
}
