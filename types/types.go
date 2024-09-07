package types

type AniResult struct {
	Title           string
	DisplayName     string
	Episodes        int
	SelectedQuality string
}

type AniEpisode struct {
	Anime        AniResult
	Number       int
	Url          string
	GetPlayerUrl func() string
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
