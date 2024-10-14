package download

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"

	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
)

type Downloader struct {
	Fetcher fetcher.Fetcher
}

func GetDownloader(fetcherName string) *Downloader {
	f, err := fetcher.GetFetcher(fetcherName)
	if err != nil {
		f = fetcher.GetDefaultFetcher()
	}
	return &Downloader{
		Fetcher: f,
	}
}

func (d *Downloader) getEpisodes(title string) ([]types.AniEpisode, error) {
	log.Println("searching for " + title)
	result := d.Fetcher.GetAnimeResult(title)
	if result == nil {
		return nil, errors.New("anime not found")
	}
	log.Printf("found anime %s\n", result.DisplayName)
	episodes := d.Fetcher.GetEpisodes(*result)

	log.Printf("total episode number is %v\n", len(episodes))
	return episodes, nil
}

func (d *Downloader) downloadEpisodeToDisk(episode types.AniEpisode, path string) error {
	url := episode.GetPlayerUrl()
	log.Printf("downloading episode (%v) to %s\n", episode.Number, path)

	log.Printf("found the episode url : %s\n", url)
	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch video data")
	}

	// Create a progress bar
	bar := progressbar.NewOptions64(resp.ContentLength,
		progressbar.OptionSetDescription("Downloading video..."),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
	)

	// Copy the response body to the file and update the progress bar
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}

func (d *Downloader) DownloadAllEpisodes(title string, path string) error {
	episodes, err := d.getEpisodes(title)
	if err != nil {
		return err
	}
	for i, ep := range episodes {
		epNumber := i + 1
		err := d.downloadEpisodeToDisk(
			ep,
			filepath.Join(path, fmt.Sprintf("%s-episode-%v.mp4", title, epNumber)),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Downloader) DownloadEpisode(title string, ep int, path string) error {
	episodes, err := d.getEpisodes(title)
	if err != nil {
		return err
	}

	if ep > len(episodes) {
		return errors.New("episode out of range")
	}
	index := ep - 1
	episode := episodes[index]

	return d.downloadEpisodeToDisk(
		episode,
		filepath.Join(path, fmt.Sprintf("%s-episode-%v.mp4", title, ep)),
	)
}
