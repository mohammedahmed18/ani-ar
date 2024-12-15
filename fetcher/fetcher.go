package fetcher

// TODO: use plugin system for fetches and make them open source to allow people make their own fetchers
import (
	"errors"

	"github.com/ani/ani-ar/fetcher/allanime"
	"github.com/ani/ani-ar/fetcher/anime3rb"
	"github.com/ani/ani-ar/fetcher/egydead"
	"github.com/ani/ani-ar/types"
)

type Fetcher interface {
	Search(q string) []types.AniResult
	GetAnimeResult(id string) *types.AniResult
	GetEpisodes(types.AniResult) []types.AniEpisode
}

var fetchers = make(map[string]Fetcher)

func init() {
	registerFetcher("anime3rb", anime3rb.GetAnime3rbFetcher())
	registerFetcher("allanime", allanime.GetAllAnimeFetcher())
	registerFetcher("egydead", egydead.GetEgyDeadFetcher())
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
