package main

type AniResult struct {
	title       string
	displayName string
	episodes    int
}

type AniEpisode struct {
	number  int
	getUrl  func() string
	result  AniResult
	quality string // TODO: display the quality in the player title
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
