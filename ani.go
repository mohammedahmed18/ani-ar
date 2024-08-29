package main

type AniResult struct {
	name     string
	episodes int
}

type AniEpisode struct {
	number int
	url    string
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
