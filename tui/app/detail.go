package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicolaieilstrup/your-ai-memory/tui/wiki"
)

var (
	detailTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	panelStyle       = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8")).
				Padding(0, 1)
)

// DetailModel shows stats and recent log entries for a single wiki.
type DetailModel struct {
	w      wiki.Wiki
	stats  wiki.Stats
	wikis  []wiki.Wiki
	width  int
	height int
}

func NewDetailModel(w wiki.Wiki, wikis []wiki.Wiki) DetailModel {
	stats := wiki.ReadWikiStats(w.Path)
	return DetailModel{w: w, stats: stats, wikis: wikis}
}

func (m DetailModel) Init() tea.Cmd { return nil }

func (m DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg { return navigateToHubMsg{} }
		case "o":
			return m, tea.ExecProcess(
				buildCmdInDir(m.w.Path, "claude"),
				func(err error) tea.Msg { return nil },
			)
		case "s":
			return m, func() tea.Msg {
				return navigateToSearchMsg{wiki: m.w, crossWiki: false}
			}
		case "a":
			return m, func() tea.Msg {
				return navigateToSourcesMsg{wiki: m.w}
			}
		case "i":
			return m, tea.ExecProcess(
				buildCmdInDir(m.w.Path, "claude", "--dangerously-skip-permissions", "/ingest"),
				func(err error) tea.Msg { return nil },
			)
		case "l":
			return m, func() tea.Msg {
				return navigateToOpsMsg{wiki: m.w}
			}
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	var sb strings.Builder

	sb.WriteString(detailTitleStyle.Render(m.w.Name))
	sb.WriteString("  ")
	sb.WriteString(dimStyle.Render(m.w.Domain))
	sb.WriteString("\n\n")

	// Stats panel (left)
	var statLines []string
	statLines = append(statLines, statLabelStyle.Render("Pages"))
	for _, cat := range []string{"concepts", "patterns", "papers", "people", "connections", "questions"} {
		count := m.stats.PageCounts[cat]
		statLines = append(statLines, fmt.Sprintf(
			"  %-12s %s",
			statLabelStyle.Render(cat),
			statValueStyle.Render(fmt.Sprintf("%d", count)),
		))
	}
	statLines = append(statLines, "")
	statLines = append(statLines, fmt.Sprintf(
		"%s %s",
		statLabelStyle.Render("Total:"),
		statValueStyle.Render(fmt.Sprintf("%d", m.stats.TotalPages)),
	))
	statsPanel := panelStyle.Render(strings.Join(statLines, "\n"))

	// Log panel (right)
	var logLines []string
	logLines = append(logLines, statLabelStyle.Render("Recent ingests"))
	if len(m.stats.RecentLogEntries) == 0 {
		logLines = append(logLines, dimStyle.Render("  (no ingests yet)"))
	} else {
		for _, entry := range m.stats.RecentLogEntries {
			logLines = append(logLines, "  "+dimStyle.Render(entry))
		}
	}
	logPanel := panelStyle.Render(strings.Join(logLines, "\n"))

	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, statsPanel, "  ", logPanel))
	sb.WriteString("\n\n")

	help := helpStyle.Render(
		"[o] open in Claude Code   [s] search   [a] add source   [i] ingest   [l] ops   [esc] back",
	)
	sb.WriteString(help)

	return sb.String()
}
