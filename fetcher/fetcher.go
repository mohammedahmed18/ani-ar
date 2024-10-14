package fetcher

import (
	"errors"

	"github.com/ani/ani-ar/types"
)

type Fetcher interface {
	GetAnimeResult(string) *types.AniResult
	Search(string) []types.AniResult
	GetEpisodes(types.AniResult) []types.AniEpisode
}

var fetchers = make(map[string]Fetcher)

func init() {
	registerFetcher("anime3rb", getAnime3rbFetcher())
	registerFetcher("anime4up", getAnime4upFetcher())
}

func registerFetcher(name string, f Fetcher) error {
	if _, ok := fetchers[name]; ok {
		return errors.New("fetcher already registered")
	}

	fetchers[name] = f
	return nil
}

func GetFetcher(name string) (Fetcher, error) {
	if f, ok := fetchers[name]; ok {
		return f, nil
	}
	return nil, errors.New("fetcher name is unknown")
}

// TODO: allow user to select the fetcher through args or something
func GetDefaultFetcher() Fetcher {
	f, _ := GetFetcher("anime3rb")
	return f
}
