package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const jikanBaseUrl = "https://api.jikan.moe/v4"

type JikanApi struct {
	C *cache.Cache
}

var jikanApi *JikanApi

func GetJikanApi() *JikanApi {
	if jikanApi != nil {
		return jikanApi
	}
	jikanApi = &JikanApi{
		C: cache.New(5*time.Minute, 10*time.Minute),
	}
	return jikanApi
}

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

func (j *JikanApi) getBestMatchAnimeInfo(animeTitleOrId string) *JikanAnimeInfo {
	cacheKey := "jikan.result." + animeTitleOrId
	if v, found := j.C.Get(cacheKey); found {
		return v.(*JikanAnimeInfo)
	}
	res, err := http.Get(fmt.Sprintf("%s/anime?q=%s", jikanBaseUrl, animeTitleOrId))
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
	bestMatch := response.Data[0]
	j.C.Set(cacheKey, bestMatch, time.Hour*24*6)
	return bestMatch
}

func (j *JikanApi) getEpisodesWithPagination(episodes []*JikanAnimeEpisode, animeMalId int, page int) []*JikanAnimeEpisode {
	cacheKey := "jikan.episodes." + fmt.Sprintf("%v", animeMalId)
	if v, found := j.C.Get(cacheKey); found {
		return v.([]*JikanAnimeEpisode)
	}

	res, err := http.Get(fmt.Sprintf("%s/anime/%v/episodes?page=%v", jikanBaseUrl, animeMalId, page))
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
		j.C.Set(cacheKey, episodes, time.Hour*24*1)
		return episodes
	}
	return j.getEpisodesWithPagination(episodes, animeMalId, page+1)

}

func (j *JikanApi) getEpisodes(animeMalId int) []*JikanAnimeEpisode {
	return j.getEpisodesWithPagination([]*JikanAnimeEpisode{}, animeMalId, 1)
}

func (j *JikanApi) getSingleEpisode(animeMalId, episodeNum int) *JikanAnimeEpisode {
	cacheKey := "jikan.episodes." + fmt.Sprintf("%v", animeMalId)
	if v, found := j.C.Get(cacheKey); found {
		allCachedEpisodes := v.([]*JikanAnimeEpisode)
		return allCachedEpisodes[episodeNum-1]
	}

	res, err := http.Get(fmt.Sprintf("%s/anime/%d/episodes/%d", jikanBaseUrl, animeMalId, episodeNum))
	if err != nil {
		println(err.Error())
		return nil
	}

	defer res.Body.Close()
	type Response struct {
		Data *JikanAnimeEpisode `json:"data"`
	}
	var response Response
	json.NewDecoder(res.Body).Decode(&response)
	return response.Data
}
