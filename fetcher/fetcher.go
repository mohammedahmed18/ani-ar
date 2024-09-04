package fetcher

import "github.com/ani/ani-ar/types"

type Fetcher interface {
	Search(string) []types.AniResult
	GetEpisodes(types.AniResult) []types.AniEpisode
}

// TODO: allow user to select the fetcher through args or something
func GetDefaultFetcher() Fetcher {
	return GetAnime4upFetcher()
	// f := &Anime3rb{}
	// return f
}
