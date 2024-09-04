package types

type AniResult struct {
	Title       string
	DisplayName string
	Episodes    int
}

type AniEpisode struct {
	Number       int
	Url          string
	GetPlayerUrl func() string
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
