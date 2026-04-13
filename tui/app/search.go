package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicolaieilstrup/your-ai-memory/tui/wiki"
)

var (
	searchTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	crossWikiStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	resultFileStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	resultExcerptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	selectedRowStyle  = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("15"))
)

// SearchModel handles both single-wiki and cross-wiki search.
type SearchModel struct {
	wikiAnchor wiki.Wiki
	allWikis   []wiki.Wiki
	crossWiki  bool
	input      textinput.Model
	results    []wiki.SearchResult
	cursor     int
	preview    viewport.Model
	showPreview bool
	width      int
	height     int
	err        string
}

func NewSearchModel(w wiki.Wiki, allWikis []wiki.Wiki, crossWiki bool) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "type to search…"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	vp := viewport.New(80, 20)

	return SearchModel{
		wikiAnchor: w,
		allWikis:   allWikis,
		crossWiki:  crossWiki,
		input:      ti,
		preview:    vp,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return navigateToDetailMsg{wiki: m.wikiAnchor} }

		case "tab":
			// Toggle single/cross-wiki
			m.crossWiki = !m.crossWiki
			m.results = m.doSearch(m.input.Value())
			m.cursor = 0
			m.updatePreview()
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updatePreview()
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
				m.updatePreview()
			}
			return m, nil

		case "p":
			m.showPreview = !m.showPreview
			return m, nil

		case "c":
			// Ask Claude — assemble cross-wiki context and open claude
			return m, m.askClaude()

		case "enter":
			// Open the selected result file in the system editor or Claude
			if m.cursor < len(m.results) {
				r := m.results[m.cursor]
				return m, tea.ExecProcess(
					buildCmd("claude", filepath.Join(m.wikiAnchor.Path, r.File)),
					func(err error) tea.Msg { return nil },
				)
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	prev := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	if m.input.Value() != prev {
		m.results = m.doSearch(m.input.Value())
		m.cursor = 0
		m.updatePreview()
	}

	if m.showPreview {
		m.preview, cmd = m.preview.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *SearchModel) doSearch(q string) []wiki.SearchResult {
	if strings.TrimSpace(q) == "" {
		return nil
	}
	if m.crossWiki {
		return wiki.CrossSearch(m.allWikis, q)
	}
	return wiki.Search(m.wikiAnchor.Path, m.wikiAnchor.Name, q)
}

func (m *SearchModel) updatePreview() {
	if !m.showPreview || m.cursor >= len(m.results) {
		return
	}
	r := m.results[m.cursor]
	content, err := os.ReadFile(r.File)
	if err != nil {
		m.preview.SetContent(dimStyle.Render("(could not read file)"))
		return
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.preview.Width),
	)
	if err != nil {
		m.preview.SetContent(string(content))
		return
	}
	rendered, err := renderer.Render(string(content))
	if err != nil {
		m.preview.SetContent(string(content))
		return
	}
	m.preview.SetContent(rendered)
}

func (m SearchModel) askClaude() tea.Cmd {
	// Determine active topic
	topic := ""
	if len(m.wikiAnchor.Topics) > 0 {
		topic = m.wikiAnchor.Topics[0]
	}
	ctx := wiki.AssembleCrossWikiContext(m.allWikis, topic, m.input.Value())
	if ctx == "" {
		ctx = m.input.Value()
	}

	// Write to temp file
	tmp, err := os.CreateTemp("", fmt.Sprintf("your-ai-memory-context-%d-*.md", time.Now().Unix()))
	if err != nil {
		return nil
	}
	tmp.WriteString(ctx)
	tmp.Close()

	return tea.ExecProcess(
		buildCmd("claude", "-p", tmp.Name()),
		func(err error) tea.Msg {
			os.Remove(tmp.Name())
			return nil
		},
	)
}

func (m SearchModel) View() string {
	var sb strings.Builder

	mode := searchTitleStyle.Render(m.wikiAnchor.Name)
	if m.crossWiki {
		topic := "all"
		if len(m.wikiAnchor.Topics) > 0 {
			topic = m.wikiAnchor.Topics[0]
		}
		mode = crossWikiStyle.Render(fmt.Sprintf("cross-wiki [%s]", topic))
	}
	sb.WriteString(fmt.Sprintf("  Search — %s\n\n", mode))
	sb.WriteString("  " + m.input.View() + "\n\n")

	if len(m.results) == 0 && m.input.Value() != "" {
		sb.WriteString(dimStyle.Render("  No results.\n"))
	}

	for i, r := range m.results {
		prefix := "  "
		line := fmt.Sprintf("%s  %s:%d  %s",
			resultFileStyle.Render(r.WikiName),
			resultFileStyle.Render(r.File),
			r.Line,
			resultExcerptStyle.Render(r.Excerpt),
		)
		if i == m.cursor {
			line = selectedRowStyle.Render(prefix + line)
		} else {
			line = prefix + line
		}
		sb.WriteString(line + "\n")
		if i >= 15 {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  … %d more\n", len(m.results)-i-1)))
			break
		}
	}

	if m.showPreview && len(m.results) > 0 {
		sb.WriteString("\n")
		sb.WriteString(panelStyle.Render(m.preview.View()))
		sb.WriteString("\n")
	}

	help := helpStyle.Render(
		"[↑↓] navigate   [tab] toggle cross-wiki   [p] preview   [c] ask Claude   [enter] open   [esc] back",
	)
	sb.WriteString("\n" + help)
	return sb.String()
}
