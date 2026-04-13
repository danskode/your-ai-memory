package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicolaieilstrup/your-ai-memory/tui/wiki"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	groupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).MarginTop(1)
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)
)

type wikiItem struct {
	w wiki.Wiki
}

func (i wikiItem) Title() string       { return i.w.Name }
func (i wikiItem) Description() string { return i.w.Domain }
func (i wikiItem) FilterValue() string { return i.w.Name + " " + i.w.Domain }

// HubModel is the home screen showing all wikis grouped by topic.
type HubModel struct {
	list   list.Model
	wikis  []wiki.Wiki
	width  int
	height int
	// flat ordered list mirroring list.Model items for index lookup
	ordered []wiki.Wiki
}

func NewHubModel(wikis []wiki.Wiki) HubModel {
	items, ordered := buildItems(wikis)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("12")).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("14"))

	l := list.New(items, delegate, 80, 24)
	l.Title = "your-ai-memory"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return HubModel{list: l, wikis: wikis, ordered: ordered}
}

func buildItems(wikis []wiki.Wiki) ([]list.Item, []wiki.Wiki) {
	groups := wiki.GroupByTopic(wikis)
	tags := make([]string, 0, len(groups))
	for t := range groups {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	var items []list.Item
	var ordered []wiki.Wiki
	for _, tag := range tags {
		_ = tag // grouping shown via description prefix
		for _, w := range groups[tag] {
			items = append(items, wikiItem{w: w})
			ordered = append(ordered, w)
		}
	}
	return items, ordered
}

func (m HubModel) Init() tea.Cmd { return nil }

func (m HubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			selected, ok := m.list.SelectedItem().(wikiItem)
			if ok {
				return m, func() tea.Msg {
					return navigateToDetailMsg{wiki: selected.w}
				}
			}

		case "n":
			// Shell out to `npx your-ai-memory create`
			return m, tea.ExecProcess(
				buildCmd("npx", "your-ai-memory", "create"),
				func(err error) tea.Msg { return navigateToHubMsg{} },
			)

		case "/":
			// Cross-wiki search (no specific wiki selected)
			return m, func() tea.Msg {
				// Pass first wiki as anchor, crossWiki=true
				if len(m.wikis) > 0 {
					sel, _ := m.list.SelectedItem().(wikiItem)
					return navigateToSearchMsg{wiki: sel.w, crossWiki: true}
				}
				return nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m HubModel) View() string {
	if len(m.wikis) == 0 {
		empty := strings.Join([]string{
			titleStyle.Render("your-ai-memory"),
			"",
			"  No wikis registered yet.",
			"",
			dimStyle.Render("  Run: npx your-ai-memory create"),
		}, "\n")
		return empty
	}

	// Render grouped list header above the bubbles list
	groups := wiki.GroupByTopic(m.wikis)
	tags := make([]string, 0, len(groups))
	for t := range groups {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	var header strings.Builder
	for _, tag := range tags {
		header.WriteString(groupStyle.Render(fmt.Sprintf("  [%s]", tag)) + "\n")
		for _, w := range groups[tag] {
			accessed := w.LastAccessed
			if accessed == "" {
				accessed = w.Created
			}
			header.WriteString(fmt.Sprintf("    %-20s %-35s %s\n",
				w.Name,
				dimStyle.Render(w.Domain),
				dimStyle.Render(accessed),
			))
		}
		header.WriteString("\n")
	}

	help := helpStyle.Render("  [n] new wiki   [enter] open   [/] search   [q] quit")
	return header.String() + "\n" + m.list.View() + "\n" + help
}
