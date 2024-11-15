package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JikanApi struct{}

type JikanAnimeInfo struct {
	MalID          int        `json:"mal_id"`
	URL            string     `json:"url"`
	Images         Images     `json:"images"`
	Trailer        Trailer    `json:"trailer"`
	Approved       bool       `json:"approved"`
	Titles         []Title    `json:"titles"`
	Title          string     `json:"title"`
	TitleEnglish   string     `json:"title_english"`
	TitleJapanese  string     `json:"title_japanese"`
	TitleSynonyms  []string   `json:"title_synonyms"`
	Type           string     `json:"type"`
	Source         string     `json:"source"`
	Episodes       int        `json:"episodes"`
	Status         string     `json:"status"`
	Airing         bool       `json:"airing"`
	Aired          AiredDates `json:"aired"`
	Duration       string     `json:"duration"`
	Rating         string     `json:"rating"`
	Score          float64    `json:"score"`
	ScoredBy       int        `json:"scored_by"`
	Rank           int        `json:"rank"`
	Popularity     int        `json:"popularity"`
	Members        int        `json:"members"`
	Favorites      int        `json:"favorites"`
	Synopsis       string     `json:"synopsis"`
	Background     string     `json:"background"`
	Season         string     `json:"season"`
	Year           int        `json:"year"`
	Broadcast      Broadcast  `json:"broadcast"`
	Producers      []Company  `json:"producers"`
	Licensors      []Company  `json:"licensors"`
	Studios        []Company  `json:"studios"`
	Genres         []Genre    `json:"genres"`
	ExplicitGenres []Genre    `json:"explicit_genres"`
	Themes         []Genre    `json:"themes"`
	Demographics   []Genre    `json:"demographics"`
}

type JikanAnimeEpisode struct {
	MalID         int      `json:"mal_id"`
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	TitleJapanese string   `json:"title_japanese"`
	TitleRomanji  string   `json:"title_romanji"`
	Aired         string   `json:"aired"`
	Score         *float64 `json:"score,omitempty"`
	Filler        bool     `json:"filler"`
	Recap         bool     `json:"recap"`
	ForumURL      string   `json:"forum_url"`
}

// Images represents the image URLs for both JPG and WebP formats
type Images struct {
	JPG  JPGImages `json:"jpg"`
	WEBP JPGImages `json:"webp"`
}

// JPGImages represents the URLs for image sizes in JPG/WebP format
type JPGImages struct {
	ImageURL      string `json:"image_url"`
	SmallImageURL string `json:"small_image_url"`
	LargeImageURL string `json:"large_image_url"`
}

// Trailer represents the trailer information for the anime
type Trailer struct {
	YoutubeID string `json:"youtube_id"`
	URL       string `json:"url"`
	EmbedURL  string `json:"embed_url"`
}

// Title represents the different titles of the anime
type Title struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

// AiredDates represents the airing dates of the anime
type AiredDates struct {
	From string    `json:"from"`
	To   string    `json:"to"`
	Prop DateRange `json:"prop"`
}

// DateRange represents the detailed date range
type DateRange struct {
	From   DateDetail `json:"from"`
	To     DateDetail `json:"to"`
	String string     `json:"string"`
}

// DateDetail represents the day, month, and year for specific dates
type DateDetail struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

// Broadcast represents the broadcast time information
type Broadcast struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	Timezone string `json:"timezone"`
	String   string `json:"string"`
}

// Company represents companies like producers, licensors, and studios
type Company struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// Genre represents a genre or demographic (could be used for explicit genres, themes, etc.)
type Genre struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

func (JikanApi) getBestMatchAnimeInfo(animeTitleOrId string) *JikanAnimeInfo {
	res, err := http.Get(fmt.Sprintf("https://api.jikan.moe/v4/anime?q=%s", animeTitleOrId))
	if err != nil {
		println(err.Error())
		return nil
	}

	defer res.Body.Close()
	type Response struct {
		Data []*JikanAnimeInfo `json:"data"`
	}
	var response Response
	json.NewDecoder(res.Body).Decode(&response)
	return response.Data[0]
}

func (j JikanApi) getEpisodesWithPagination(episodes []*JikanAnimeEpisode, animeMalId int, page int) []*JikanAnimeEpisode {
	res, err := http.Get(fmt.Sprintf("https://api.jikan.moe/v4/anime/%v/episodes?page=%v", animeMalId, page))
	if err != nil {
		println(err.Error())
		return []*JikanAnimeEpisode{}
	}

	defer res.Body.Close()
	type Response struct {
		Data       []*JikanAnimeEpisode `json:"data"`
		Pagination struct {
			LastVisiblePage int `json:"last_visible_page"`
		} `json:"pagination"`
	}
	var response Response
	json.NewDecoder(res.Body).Decode(&response)
	episodes = append(episodes, response.Data...)

	if page == response.Pagination.LastVisiblePage {
		return episodes
	}
	return j.getEpisodesWithPagination(episodes, animeMalId, page+1)

}

func (j JikanApi) getEpisodes(animeMalId int) []*JikanAnimeEpisode {
	return j.getEpisodesWithPagination([]*JikanAnimeEpisode{}, animeMalId, 1)
}
