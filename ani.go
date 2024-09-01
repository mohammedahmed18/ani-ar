package main

type AniResult struct {
	title       string
	displayName string
	episodes    int
}

type AniEpisode struct {
	number int
	getUrl func() string
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
