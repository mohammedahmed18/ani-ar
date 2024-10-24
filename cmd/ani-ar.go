package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"

	"github.com/ani/ani-ar/download"
	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/gui"
	"github.com/ani/ani-ar/player"
)

func main() {
	app := &cli.App{
		Name:  "ani-ar",
		Usage: "watch anime from terminal with arabic sub",
		Commands: []*cli.Command{
			{
				Name: "search",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "q",
						Value:    "",
						Usage:    "query search for the anime",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					q := ctx.String("q")
					results := fetcher.GetDefaultFetcher().Search(q)
					if len(results) == 0 {
						return errors.New("no results found")
					}
					// display results
					for _, r := range results {
						fmt.Printf("%s (%v episode(s))\n", r.Title, r.Episodes)
					}
					return nil
				},
			},
			{
				Name: "watch",
				Args: true,
				Action: func(ctx *cli.Context) error {
					title := ctx.Args().First()
					episode := ctx.Args().Get(1)
					animeEpisode, _ := strconv.Atoi(episode)
					result := fetcher.GetDefaultFetcher().GetAnimeResult(title)
					if result == nil {
						return errors.New("can't find anime")
					}
					episodes := fetcher.GetDefaultFetcher().GetEpisodes(*result)
					ep := episodes[animeEpisode-1]
					log.Println("getting the episode video...")
					videoUrl := ep.GetPlayerUrl()
					log.Println("found it")
					return player.RunVideo(
						videoUrl,
						fmt.Sprintf("%s-episode-%v", title, animeEpisode),
					)
				},
			},
			{
				Name: "download",
				Args: true,
				// Aliases: []string{""},
				Usage: "download anime episode or download all episodes",
				Action: func(ctx *cli.Context) error {
					animeTitle := ctx.Args().First()
					episode := ctx.Args().Get(1)
					animeEpisode, _ := strconv.Atoi(episode)
					path := filepath.Join("anime/")
					os.MkdirAll(path, 0777)
					if animeEpisode == 0 {
						return download.GetDownloader("").DownloadAllEpisodes(animeTitle, path)
					}
					return download.GetDownloader("").
						DownloadEpisode(animeTitle, animeEpisode, path)
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			p := tea.NewProgram(gui.InitialModel())
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
