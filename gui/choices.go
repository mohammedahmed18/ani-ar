package gui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ani/ani-ar/types"
)

type ChoicesModel struct {
	choices      []interface{}
	cursor       int
	spinner      spinner.Model
	loading      bool
	resultsShown bool

	searchKey        string
	choiceFormatFunc func(interface{}) string

	textInput                textinput.Model
	viewport                 viewport.Model
	firstChoiceVisibleCursor int
}

const vpHight = 20

func getSpinnerForChoices() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

func getFilterTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Filter Search result"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return ti
}

func initialChoicesModelForAnimeTitles() ChoicesModel {
	vp := viewport.New(120, vpHight)
	vp.SetContent(`loading...`)
	return ChoicesModel{
		spinner:   getSpinnerForChoices(),
		textInput: getFilterTextInput(),
		viewport:  vp,
		choiceFormatFunc: func(i interface{}) string {
			anime := i.(types.AniResult)
			suf := "episode"
			if anime.Episodes > 1 {
				suf = "episodes"
			}
			return fmt.Sprintf("%s - %v %s", anime.DisplayName, anime.Episodes, suf)
		},
	}
}

func initialChoicesModelForAnimeEpisode() ChoicesModel {
	vp := viewport.New(30, vpHight)
	return ChoicesModel{
		spinner:   getSpinnerForChoices(),
		textInput: getFilterTextInput(),
		viewport:  vp,
		choiceFormatFunc: func(i interface{}) string {
			episode := i.(types.AniEpisode)
			return fmt.Sprintf("episode #%v", episode.Number)
		},
	}
}

func (m ChoicesModel) getSelectedChoice() interface{} {
	return m.getFilteredChoices(m.choices)[m.cursor]
}

func (m ChoicesModel) getFilteredChoices(choices []interface{}) []interface{} {
	var filteredChoices []interface{}
	for _, r := range choices {
		formatted := m.choiceFormatFunc(r)
		filterKey := m.textInput.Value()
		if filterKey != "" {
			if !strings.Contains(strings.ToLower(formatted), strings.ToLower(filterKey)) {
				continue
			}
		}
		filteredChoices = append(filteredChoices, r)
	}
	return filteredChoices
}

func (m ChoicesModel) getViewportContentFromChoices(choices []interface{}, cursor int) string {
	content := ""
	if cursor == -1 {
		cursor = m.cursor
	}
	// Display choices
	filtered := m.getFilteredChoices(choices)
	for i, r := range filtered {
		displayCursor := " "
		if cursor == i {
			displayCursor = ">"
		}
		formatted := m.choiceFormatFunc(r)

		content += fmt.Sprintf("%s %v- %s\n", displayCursor, i+1, formatted)
	}
	if len(choices) == 0 {
		content += "No matched results!!\n"
	}
	return content
}

func (m ChoicesModel) fetchChoices(
	searchfunc func() []interface{},
	key string,
) (tea.Model, tea.Cmd) {
	m.searchKey = key

	// Show the spinner
	m.loading = true

	// Return the model to show the spinner
	newModel, _ := m.Update(newChoicesLoadingEvent())
	m = newModel.(ChoicesModel)

	// Fetch data in a separate command
	fetchDataCmd := func() tea.Msg {
		results := searchfunc()
		return newChoicesShownEvent(results)
	}

	return m, tea.Sequence(
		func() tea.Msg {
			return newChoicesLoadingEvent()
		},
		tea.Cmd(func() tea.Msg {
			return fetchDataCmd()
		}),
	)
}

func (m ChoicesModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.viewport.Init())
}

func (m ChoicesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd   tea.Cmd
		vpCmd tea.Cmd
	)
	autoUpdateViewPort := true
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyDown:
			autoUpdateViewPort = false
			filtered := m.getFilteredChoices(m.choices)
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}
			lastVisibleItemCursor := m.firstChoiceVisibleCursor + vpHight - 1
			if m.cursor > lastVisibleItemCursor {
				m.viewport.LineDown(1)
				m.firstChoiceVisibleCursor++
			}
		case tea.KeyUp:
			autoUpdateViewPort = false
			if m.cursor > 0 {
				m.cursor--
			}
			if m.cursor < m.firstChoiceVisibleCursor {
				m.viewport.LineUp(1)
				m.firstChoiceVisibleCursor--
			}
		case tea.KeyEnter:
			return m, cmd
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ChoicesLoadingEvent:
		m.loading = true
		m.choices = []interface{}{}
		m.resultsShown = false
		return m, cmd

	case ChoicesShownEvent:
		m.cursor = 0
		m.loading = false
		m.choices = msg.results
		m.resultsShown = true
		m.viewport.SetContent(m.getViewportContentFromChoices(msg.results, 0))
		return m, cmd

	default:
		return m, tea.Batch(cmd, vpCmd)
	}

	if autoUpdateViewPort {
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	// only update the input when results are shown
	if m.resultsShown {
		keyType, ok := msg.(tea.KeyMsg)
		m.textInput, _ = m.textInput.Update(msg)
		m.viewport.SetContent(m.getViewportContentFromChoices(m.choices, -1))
		if keyType.Type != tea.KeyUp && keyType.Type != tea.KeyDown && ok {
			m.cursor = 0
			m.viewport.GotoTop()
			m.viewport.SetContent(m.getViewportContentFromChoices(m.choices, -1))
		}
	}

	return m, tea.Batch(cmd, vpCmd)
}

func (m ChoicesModel) View() string {
	msg := ""

	if m.loading {
		// Show spinner while loading
		msg += m.spinner.View() + " Loading...\n"
	}

	if m.resultsShown {
		msg += m.textInput.View()
		msg += "\n"
		msg += "Showing " + strconv.Itoa(len(m.choices)) + " results for " + m.searchKey + "\n\n"
		msg += m.viewport.View()
	}

	return msg
}
