package main

type Fetcher interface {
	search(string) []AniResult
	getEpisodes(AniResult) []AniEpisode
}

// TODO: allow user to select the fetcher through args or something
func getMainFetcher() Fetcher {
	f := &Anime3rb{}
	return f
}
