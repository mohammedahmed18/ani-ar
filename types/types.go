package types

type AniResult struct {
	Id           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Episodes     int    `json:"episodes"`
	DisplayCover string `json:"displayCover"`
}
type AniVideo struct {
	Src string `json:"src"`
	Res string `json:"res"`
}
type AniEpisode struct {
	Anime                 AniResult         `json:"anime"`
	Number                int               `json:"number"`
	Url                   string            `json:"url"`
	GetPlayerUrl          func() string     `json:"-"`
	GetPlayersWithQuality func() []AniVideo `json:"-"`
}

type AnimeApi interface {
	getToken() string
	search(search string) []AniResult
}
