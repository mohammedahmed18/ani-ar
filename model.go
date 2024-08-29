package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AniModel struct {
	textInput             textinput.Model
	choicesModelAnimeList ChoicesModel
	err                   error
	// stage 0 is search anime ,
	// stage 1 is selecting the anime from the list
	// stage 2 is selecting an episode
	stage int
}

func initialModel() tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Death note"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return AniModel{
		textInput:             ti,
		err:                   nil,
		choicesModelAnimeList: initialChoicesModelForAnimeTitles(),
		stage:                 0,
	}
}

func (m AniModel) Init() tea.Cmd {
	return m.choicesModelAnimeList.Init()
}

func (m AniModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.choicesModelAnimeList.resultsShown {
				// if the results are shown then we can select an anime from a list
			} else {
				// if there is no results we fetch the results
				// m.choicesModelAnimeList = m.choicesModelAnimeList.startLoading().(ChoicesModel)
				searchKey := m.textInput.Value()
				choicesModel, c := m.choicesModelAnimeList.fetchChoices(func() []interface{} {
					var anicli Anime3rb
					results := anicli.search(searchKey)
					b := make([]interface{}, len(results))
					for i := range results {
						b[i] = results[i]
					}
					return b
				}, searchKey)
				m.choicesModelAnimeList = choicesModel.(ChoicesModel)
				m.stage = 1
				return m, c
			}

			return m, cmd
		}

	case error:
		m.err = msg
		return m, nil

	}

	// only recieve updates for the main modal input when stage is 0 (searchin anime)
	if m.stage == 0 {
		m.textInput, _ = m.textInput.Update(msg)
	}

	// only recieve updates for choices modal for anime titles when stage is 1 (selecting the anime from the list)
	if m.stage == 1 {
		newChoicesModel, _ := m.choicesModelAnimeList.Update(msg)
		m.choicesModelAnimeList = newChoicesModel.(ChoicesModel)
	}

	return m, cmd
}

func renderANewLine(msg string, highlight bool) string {
	highlightText := lipgloss.NewStyle().TabWidth(-1).Foreground(lipgloss.Color("#2c70b0"))
	normalText := lipgloss.NewStyle().TabWidth(-1).Foreground(lipgloss.Color("#f5f3f2"))

	styledText := normalText.Render(msg)
	if highlight {
		styledText = highlightText.Render(msg)
	}

	// Align text if needed
	return lipgloss.NewStyle().Align(lipgloss.Left).Render(styledText)
}

func (m AniModel) View() string {
	msg := ""

	if m.stage == 0 {
		msg += renderANewLine("Search anime ", true)
		msg += m.textInput.View()
	}

	if m.stage == 1 {
		msg += m.choicesModelAnimeList.View()
	}

	return msg
}
