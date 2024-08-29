package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type Anime3rb struct{}

const baseUrl = "https://anime3rb.com"

func (a *Anime3rb) getToken() string {
	res, err := http.Get(baseUrl)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	b, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	html := string(b)

	re := regexp.MustCompile(`<meta\s+name="csrf-token"\s+content="([^"]+)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	} else {
		fmt.Println("CSRF Token not found")
		return ""
	}
}

func (a *Anime3rb) search(search string) []AniResult {
	time.Sleep(time.Second * 1)
	return []AniResult{
		{
			name:     search,
			episodes: 12,
		},
		{
			name:     search,
			episodes: 12,
		},
		{
			name:     search,
			episodes: 12,
		},
		{
			name:     search,
			episodes: 12,
		},
		{
			name:     search,
			episodes: 12,
		},
		{
			name:     search,
			episodes: 12,
		},
	}
}

func (a *Anime3rb) getEpisodes(e AniResult) []AniEpisode {
	time.Sleep(time.Second * 1)
	return []AniEpisode{
		{
			number: 1,
			url:    "https://video.wixstatic.com/video/588cd5_408f07a9c9424376ae0acb55e763d91b/720p/mp4/file.mp4",
		},
		{
			number: 2,
			url:    "https://video.wixstatic.com/video/588cd5_408f07a9c9424376ae0acb55e763d91b/720p/mp4/file.mp4",
		},
		{
			number: 3,
			url:    "https://video.wixstatic.com/video/588cd5_408f07a9c9424376ae0acb55e763d91b/720p/mp4/file.mp4",
		},
	}
}
