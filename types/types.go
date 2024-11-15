package types

type AniResult struct {
	Id           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Episodes     int    `json:"episodes"`
	DisplayCover string `json:"displayCover"`
}

type AniEpisode struct {
	Anime        AniResult     `json:"anime"`
	Number       int           `json:"number"`
	Url          string        `json:"url"`
	GetPlayerUrl func() string `json:"-"`
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
