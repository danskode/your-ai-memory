package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicolaieilstrup/your-ai-memory/tui/wiki"
)

type opItem struct {
	title string
	desc  string
	cmd   []string // claude invocation args
}

func (i opItem) Title() string       { return i.title }
func (i opItem) Description() string { return i.desc }
func (i opItem) FilterValue() string { return i.title }

var operations = []list.Item{
	opItem{
		title: "Lint",
		desc:  "Check frontmatter, dead links, orphan pages, and duplicates",
		cmd:   []string{"claude", "/lint"},
	},
	opItem{
		title: "Gap Analysis",
		desc:  "Identify missing topics and suggest what to ingest next",
		cmd:   []string{"claude", "/gap-analysis"},
	},
	opItem{
		title: "Update Overview",
		desc:  "Regenerate wiki/overview.md from current page contents",
		cmd:   []string{"claude", "/update-overview"},
	},
}

// OpsModel shows the operations menu for a wiki.
type OpsModel struct {
	w      wiki.Wiki
	list   list.Model
	status string
	width  int
	height int
}

func NewOpsModel(w wiki.Wiki) OpsModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).Bold(true)

	l := list.New(operations, delegate, 70, 12)
	l.Title = fmt.Sprintf("%s — Operations", w.Name)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	return OpsModel{w: w, list: l}
}

func (m OpsModel) Init() tea.Cmd { return nil }

func (m OpsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return navigateToDetailMsg{wiki: m.w} }
		case "enter":
			sel, ok := m.list.SelectedItem().(opItem)
			if !ok {
				break
			}
			args := append([]string{}, sel.cmd[1:]...)
			return m, tea.ExecProcess(
				buildCmdInDir(m.w.Path, sel.cmd[0], args...),
				func(err error) tea.Msg {
					return navigateToDetailMsg{wiki: m.w}
				},
			)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m OpsModel) View() string {
	var sb strings.Builder
	sb.WriteString(m.list.View())
	sb.WriteString("\n")
	if m.status != "" {
		sb.WriteString("  " + dimStyle.Render(m.status) + "\n\n")
	}
	sb.WriteString(helpStyle.Render("  [enter] run operation   [esc] back"))
	return sb.String()
}
