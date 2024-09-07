package gui

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
)

type AniModel struct {
	textInput                textinput.Model
	choicesModelAnimeList    ChoicesModel
	choicesModelAnimeEpisode ChoicesModel
	err                      error
	// stage 0 is search anime ,
	// stage 1 is selecting the anime from the list
	// stage 2 is selecting an episode
	stage   int
	fetcher fetcher.Fetcher
	info    string
}

func InitialModel() tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Death note"
	ti.Focus()
	ti.Width = 50

	return AniModel{
		textInput:                ti,
		err:                      nil,
		choicesModelAnimeList:    initialChoicesModelForAnimeTitles(),
		choicesModelAnimeEpisode: initialChoicesModelForAnimeEpisode(),
		fetcher:                  fetcher.GetDefaultFetcher(),
		stage:                    0,
	}
}

func (m AniModel) Init() tea.Cmd {
	return tea.Batch(m.choicesModelAnimeList.Init(), m.choicesModelAnimeEpisode.Init())
}

func (m AniModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlB:
			if m.stage == 1 {
				m.stage = 0
			} else if m.stage == 2 {
				m.stage = 1
			}
			updatedModel, _ := m.Update(nil)
			m = updatedModel.(AniModel)
			return m, cmd
		case tea.KeyEnter:
			if m.stage == 0 {
				searchKey := m.textInput.Value()
				// update stage
				m.stage = 1
				m.info = "fetching data..."
				updatedModel, _ := m.Update(nil)
				m = updatedModel.(AniModel)

				choicesModel, c := m.choicesModelAnimeList.fetchChoices(func() []interface{} {
					results := m.fetcher.Search(searchKey)
					b := make([]interface{}, len(results))
					for i := range results {
						b[i] = results[i]
					}
					return b
				}, searchKey)
				m.info = ""
				m.choicesModelAnimeList = choicesModel.(ChoicesModel)
				return m, c
			}
			if m.stage == 1 {
				m.stage = 2
				updatedModel, _ := m.Update(nil)
				m = updatedModel.(AniModel)

				// anime is selected let's fetch it's episodes
				selectedAnime := m.choicesModelAnimeList.getSelectedChoice()
				anime := selectedAnime.(types.AniResult)
				newEpisodeModal, c := m.choicesModelAnimeEpisode.fetchChoices(func() []interface{} {
					episodes := m.fetcher.GetEpisodes(anime)
					b := make([]interface{}, len(episodes))
					for i := range episodes {
						b[i] = episodes[i]
					}
					return b
				}, anime.DisplayName+" episodes")

				m.choicesModelAnimeEpisode = newEpisodeModal.(ChoicesModel)

				return m, c
			}
			if m.stage == 2 {
				// play the episode
				selectedEpisode := m.choicesModelAnimeEpisode.getSelectedChoice()
				ep := selectedEpisode.(types.AniEpisode)

				epUrl := ep.GetPlayerUrl()

				title := fmt.Sprintf("%s - episode %v - %s", ep.Anime.DisplayName, ep.Number, ep.Anime.SelectedQuality)
				_, err := exec.Command(
					"mpv",
					"--title="+title,
					 epUrl).Output()
				if err != nil {
					fmt.Printf("error playing the episode %s", err)
				}

			}
			return m, cmd
		}

	case spinner.TickMsg:
		// send the tick message to the two choice lists
		m1, c1 := m.choicesModelAnimeList.Update(msg)
		m.choicesModelAnimeList = m1.(ChoicesModel)
		m2, c2 := m.choicesModelAnimeEpisode.Update(msg)
		m.choicesModelAnimeEpisode = m2.(ChoicesModel)
		return m, tea.Batch(c1, c2)
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

	// only recieve updates for choices modal for anime episodes when stage is 2 (selecting an episode)
	if m.stage == 2 {
		newChoicesModel, _ := m.choicesModelAnimeEpisode.Update(msg)
		m.choicesModelAnimeEpisode = newChoicesModel.(ChoicesModel)
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

	msg += m.info
	msg += "\n"

	if m.stage == 0 {
		msg += renderANewLine("Search anime ", true)
		msg += m.textInput.View()
	}

	if m.stage == 1 {
		msg += m.choicesModelAnimeList.View()
	}

	if m.stage == 2 {
		msg += m.choicesModelAnimeEpisode.View()
	}

	return msg
}
