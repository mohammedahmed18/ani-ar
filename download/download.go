package download

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
)

type Downloader struct {
	Fetcher fetcher.Fetcher
}

var p *tea.Program

type progressWriter struct {
	total      int
	downloaded int
	file       *os.File
	reader     io.Reader
	onProgress func(float64)
}

func (pw *progressWriter) Start() {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		p.Send(progressErrMsg{err})
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.downloaded) / float64(pw.total))
	}
	return len(p), nil
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
	log.Printf("downloading episode (%v) to %s\n", episode.Number, path)
	url := episode.GetPlayerUrl()
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

	pw := &progressWriter{
		total:  int(resp.ContentLength),
		file:   out,
		reader: resp.Body,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	m := model{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}
	// Start Bubble Tea
	p = tea.NewProgram(m)

	// Start the download
	go pw.Start()

	if _, err := p.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
	// Copy the response body to the file and update the progress bar
	_, err = io.Copy(out, resp.Body)
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
