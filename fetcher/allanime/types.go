package allanime

type AllAnimeSearch struct {
	// allowAdult bool
	AllowUnknown bool   `json:"allowUnknown"`
	Query        string `json:"query"`
}
type AllAnimeSearchVariables struct {
	Search          AllAnimeSearch `json:"search"`
	Limit           int            `json:"limit"`
	Page            int            `json:"page"`
	TranslationType string         `json:"translationType"`
}

type AllAnimeGetByIdVariables struct {
	Id string `json:"id"`
}
type AllAnimeEpisodesVariables struct {
	ShowId          string `json:"showId"`
	TranslationType string `json:"translationType"`
	EpisodeNumStart int    `json:"episodeNumStart"`
	EpisodeNumEnd   int    `json:"episodeNumEnd"`
}

type AllAnimeShow struct {
	Id           string `json:"_id"`
	Name         string `json:"name"`
	EpisodeCount string `json:"episodeCount"`
	Thumbnail    string `json:"thumbnail"`
}

type AllAnimeShowsData struct {
	Edges []AllAnimeShow `json:"edges"`
}
type AllAnimeSearchData struct {
	Shows AllAnimeShowsData `json:"shows"`
}
type AllAnimeSearchResponse struct {
	Data AllAnimeSearchData `json:"data"`
}

type AllAnimeShowResponse struct {
	Data struct {
		Show AllAnimeShow `json:"show"`
	} `json:"data"`
}

type AllAnimeEpisodeSource struct {
	SourceUrl  string  `json:"sourceUrl"`
	Priority   float64 `json:"priority"`
	SourceName string  `json:"sourceName"`
	Type       string  `json:"type"`
	ClassName  string  `json:"className"`
	StreamerId string  `json:"streamerId"`
	Downloads  struct {
		SourceName  string `json:"sourceName"`
		DownloadUrl string `json:"downloadUrl"`
	} `json:"downloads"`
}

type AllAnimeEpisode struct {
	EpisodeString string                  `json:"episodeString"`
	SourceUrls    []AllAnimeEpisodeSource `json:"sourceUrls"`
}

type AllAnimeEpisodeLinksResponse struct {
	Links []struct {
		Src           string `json:"src"`
		ResolutionStr string `json:"resolutionStr"`
	} `json:"links"`
}
type AllAnimeEpisodeResponse struct {
	Data struct {
		Episode AllAnimeEpisode `json:"episode"`
	} `json:"data"`
}
