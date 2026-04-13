package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicolaieilstrup/your-ai-memory/tui/wiki"
)

type sourcesMode int

const (
	sourcesBrowse sourcesMode = iota
	sourcesAddURL
)

type sourceItem struct {
	relPath string
	size    int64
}

func (i sourceItem) Title() string       { return filepath.Base(i.relPath) }
func (i sourceItem) Description() string { return dimStyle.Render(filepath.Dir(i.relPath)) }
func (i sourceItem) FilterValue() string { return i.relPath }

// SourcesModel browses raw/ and allows adding URL stubs.
type SourcesModel struct {
	w      wiki.Wiki
	list   list.Model
	mode   sourcesMode
	urlIn  textinput.Model
	status string
	width  int
	height int
}

func NewSourcesModel(w wiki.Wiki) SourcesModel {
	items := loadSourceItems(w.Path)
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 20)
	l.Title = fmt.Sprintf("%s — raw/", w.Name)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	ti := textinput.New()
	ti.Placeholder = "https://example.com/article"
	ti.CharLimit = 500
	ti.Width = 70

	return SourcesModel{w: w, list: l, urlIn: ti}
}

func (m SourcesModel) Init() tea.Cmd { return nil }

func (m SourcesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case sourcesBrowse:
			switch msg.String() {
			case "esc":
				return m, func() tea.Msg { return navigateToDetailMsg{wiki: m.w} }
			case "a":
				m.mode = sourcesAddURL
				m.urlIn.SetValue("")
				m.urlIn.Focus()
				return m, textinput.Blink
			case "m":
				if _, ok := m.list.SelectedItem().(sourceItem); ok {
					return m, func() tea.Msg {
						return navigateToDetailMsg{wiki: m.w}
					}
				}
			}
		case sourcesAddURL:
			switch msg.String() {
			case "esc":
				m.mode = sourcesBrowse
				return m, nil
			case "enter":
				url := strings.TrimSpace(m.urlIn.Value())
				if url != "" {
					if err := m.writeURLStub(url); err != nil {
						m.status = fmt.Sprintf("Error: %v", err)
					} else {
						m.status = "Stub written to raw/articles/"
						// Reload list
						m.list.SetItems(loadSourceItems(m.w.Path))
					}
				}
				m.mode = sourcesBrowse
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.mode {
	case sourcesBrowse:
		m.list, cmd = m.list.Update(msg)
	case sourcesAddURL:
		m.urlIn, cmd = m.urlIn.Update(msg)
	}
	return m, cmd
}

func (m SourcesModel) View() string {
	var sb strings.Builder

	sb.WriteString(m.list.View())
	sb.WriteString("\n")

	if m.status != "" {
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(m.status) + "\n\n")
	}

	if m.mode == sourcesAddURL {
		sb.WriteString("\n  Add URL:\n  ")
		sb.WriteString(m.urlIn.View())
		sb.WriteString("\n  " + dimStyle.Render("[enter] save   [esc] cancel"))
	} else {
		sb.WriteString(helpStyle.Render("  [a] add URL   [m] mark for ingest   [esc] back"))
	}

	return sb.String()
}

func (m SourcesModel) writeURLStub(url string) error {
	slug := slugify(url)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("queued-%s-%s.md", date, slug)
	destDir := filepath.Join(m.w.Path, "raw", "articles")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	content := fmt.Sprintf("---\nurl: %s\nqueued: %s\nstatus: queued\n---\n\n# Source\n\n%s\n\n*Paste or summarize the content here before ingesting.*\n", url, date, url)
	return os.WriteFile(filepath.Join(destDir, filename), []byte(content), 0644)
}

func loadSourceItems(wikiPath string) []list.Item {
	var items []list.Item
	rawDir := filepath.Join(wikiPath, "raw")
	filepath.Walk(rawDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(rawDir, path)
		items = append(items, sourceItem{relPath: rel, size: info.Size()})
		return nil
	})
	return items
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	// Strip URL scheme
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}
