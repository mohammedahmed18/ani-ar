package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChoicesModel struct {
	choices      []interface{}
	cursor       int
	spinner      spinner.Model
	loading      bool
	resultsShown bool

	searchKey        string
	choiceFormatFunc func(interface{}) string

	textInput textinput.Model
	viewport  viewport.Model
}

func getSpinnerForChoices() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Moon
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
	vp := viewport.New(30, 20)

	return ChoicesModel{
		spinner:   getSpinnerForChoices(),
		textInput: getFilterTextInput(),
		viewport:  vp,
		choiceFormatFunc: func(i interface{}) string {
			anime := i.(AniResult)
			return fmt.Sprintf("%s [%v episode(s)]", anime.name, anime.episodes)
		},
	}
}

func initialChoicesModelForAnimeEpisode() ChoicesModel {
	vp := viewport.New(30, 20)
	return ChoicesModel{
		spinner:   getSpinnerForChoices(),
		textInput: getFilterTextInput(),
		viewport:  vp,
		choiceFormatFunc: func(i interface{}) string {
			episode := i.(AniEpisode)
			return fmt.Sprintf("%v", episode.number)
		},
	}
}

func (m ChoicesModel) getSelectedChoice() interface{} {
	return m.choices[m.cursor]
}

func (m ChoicesModel) setViewportContent(choices []interface{}) {
	content := ""
	// Display header
	content += "Showing results for " + m.searchKey + "\n\n"

	content += m.textInput.View()
	content += "\n"

	displayedResults := 0
	// Display choices
	for i, r := range choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		formatted := m.choiceFormatFunc(r)
		filterKey := m.textInput.Value()
		if filterKey != "" {
			if !strings.Contains(strings.ToLower(formatted), strings.ToLower(filterKey)) {
				continue
			}
		}

		displayedResults += 1
		content += fmt.Sprintf("%s %s\n", cursor, formatted)
	}
	if displayedResults == 0 {
		content += "No matched results!!\n"
	}

	m.viewport.SetContent(content)
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
	return m.spinner.Tick
}

func (m ChoicesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
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
		return m, cmd

	case ChoicesShownEvent:
		m.loading = false
		m.choices = msg.results
		m.resultsShown = true
		m.setViewportContent(msg.results)
		m.viewport.GotoBottom()
		return m, nil

	default:
		var cmd tea.Cmd
		return m, cmd
	}

	// only update the input when results are shown
	if m.resultsShown {
		keymsg := msg.(tea.KeyMsg)

		//  move cursor to top when filtering
		if keymsg.Type != tea.KeyDown && keymsg.Type != tea.KeyUp {
			m.cursor = 0
		}
		m.textInput, _ = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m ChoicesModel) View() string {
	msg := ""

	if m.loading {
		// Show spinner while loading
		msg += m.spinner.View() + "\n"
	}

	if m.resultsShown {
		msg += m.viewport.View()
	}

	return msg
}
