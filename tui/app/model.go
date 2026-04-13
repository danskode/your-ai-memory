package app

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/nicolaieilstrup/your-ai-memory/tui/wiki"
)

type screen int

const (
	screenHub screen = iota
	screenDetail
	screenSearch
	screenSources
	screenOps
)

// RootModel is the top-level Bubbletea model. It holds the active screen and
// delegates Update/View to the appropriate sub-model.
type RootModel struct {
	screen  screen
	wikis   []wiki.Wiki
	width   int
	height  int

	hub     HubModel
	detail  DetailModel
	search  SearchModel
	sources SourcesModel
	ops     OpsModel
}

func NewRootModel(wikis []wiki.Wiki) RootModel {
	m := RootModel{
		screen: screenHub,
		wikis:  wikis,
	}
	m.hub = NewHubModel(wikis)
	return m
}

func (m RootModel) Init() tea.Cmd {
	return m.hub.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.hub.width = msg.Width
		m.hub.height = msg.Height
		m.detail.width = msg.Width
		m.detail.height = msg.Height
		m.search.width = msg.Width
		m.search.height = msg.Height
		m.sources.width = msg.Width
		m.sources.height = msg.Height

	case navigateToDetailMsg:
		m.screen = screenDetail
		m.detail = NewDetailModel(msg.wiki, m.wikis)
		return m, m.detail.Init()

	case navigateToSearchMsg:
		m.screen = screenSearch
		m.search = NewSearchModel(msg.wiki, m.wikis, msg.crossWiki)
		return m, m.search.Init()

	case navigateToSourcesMsg:
		m.screen = screenSources
		m.sources = NewSourcesModel(msg.wiki)
		return m, m.sources.Init()

	case navigateToOpsMsg:
		m.screen = screenOps
		m.ops = NewOpsModel(msg.wiki)
		return m, m.ops.Init()

	case navigateToHubMsg:
		m.screen = screenHub
		// Reload wikis in case registry changed
		reloaded, _ := wiki.LoadRegistry()
		m.wikis = reloaded
		m.hub = NewHubModel(m.wikis)
		return m, m.hub.Init()
	}

	switch m.screen {
	case screenHub:
		newHub, cmd := m.hub.Update(msg)
		m.hub = newHub.(HubModel)
		return m, cmd
	case screenDetail:
		newDetail, cmd := m.detail.Update(msg)
		m.detail = newDetail.(DetailModel)
		return m, cmd
	case screenSearch:
		newSearch, cmd := m.search.Update(msg)
		m.search = newSearch.(SearchModel)
		return m, cmd
	case screenSources:
		newSources, cmd := m.sources.Update(msg)
		m.sources = newSources.(SourcesModel)
		return m, cmd
	case screenOps:
		newOps, cmd := m.ops.Update(msg)
		m.ops = newOps.(OpsModel)
		return m, cmd
	}
	return m, nil
}

func (m RootModel) View() string {
	switch m.screen {
	case screenHub:
		return m.hub.View()
	case screenDetail:
		return m.detail.View()
	case screenSearch:
		return m.search.View()
	case screenSources:
		return m.sources.View()
	case screenOps:
		return m.ops.View()
	}
	return ""
}

// Navigation messages — sent by sub-models to trigger screen transitions.

type navigateToDetailMsg struct{ wiki wiki.Wiki }
type navigateToSearchMsg struct {
	wiki      wiki.Wiki
	crossWiki bool
}
type navigateToSourcesMsg struct{ wiki wiki.Wiki }
type navigateToOpsMsg struct{ wiki wiki.Wiki }
type navigateToHubMsg struct{}
