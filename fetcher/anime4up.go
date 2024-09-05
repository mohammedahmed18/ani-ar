package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/ani/ani-ar/extractors"
	"github.com/ani/ani-ar/types"
)

const anime4upBaseUrl string = "https://aname4up.shop"

type Anime4up struct{}

func getAnime4upFetcher() Fetcher {
	return Anime4up{}
}

func (a Anime4up) Search(q string) []types.AniResult {
	url := fmt.Sprintf("%s/?search_param=animes&s=%s", anime4upBaseUrl, url.QueryEscape(q))
	res, _ := http.Get(url)

	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)

	var results []types.AniResult

	doc.Find(".anime-card-container").Each(func(i int, card *goquery.Selection) {
		cardDetails := card.Find(".anime-card-details .anime-card-title").First()
		cardLink := cardDetails.Find("a").First()
		displayName := cardLink.Text()

		title := ""
		link, _ := cardLink.Attr("href")
		//
		linkParts := strings.Split(link, "/")
		lastPart := linkParts[len(linkParts)-1]

		if lastPart == "" {
			title = linkParts[len(linkParts)-2]
		} else {
			title = lastPart
		}
		results = append(results, types.AniResult{
			Title:       title,
			DisplayName: displayName,
			Episodes:    getEpisodesCountForAnime(link),
		})
	})
	return results
}

func getEpisodesCountForAnime(link string) int {
	res, _ := http.Get(link)
	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	return doc.Find(".episodes-card").Length()
}

func (a Anime4up) GetEpisodes(r types.AniResult) []types.AniEpisode {
	var episodes []types.AniEpisode
	link := fmt.Sprintf("%s/anime/%s", anime4upBaseUrl, r.Title)

	res, _ := http.Get(link)
	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	doc.Find(".episodes-card").Each(func(i int, episodeCard *goquery.Selection) {
		epUrl, _ := episodeCard.Find(".episodes-card-title a").Attr("href")
		episodes = append(episodes, types.AniEpisode{
			Number: i + 1,
			Url:    epUrl,
			GetPlayerUrl: func() string {
				return a.GetLazyVideoUrl(epUrl)
			},
		})
	})
	return episodes
}

func (a Anime4up) GetLazyVideoUrl(epUrl string) string {
	res, _ := http.Get(epUrl)
	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	watchForm := doc.Find(".watchForm").First()

	actionUrl, _ := watchForm.Find("form").First().Attr("action")
	var form url.Values = url.Values{}
	watchForm.Find("input[type=\"hidden\"]").Each(func(i int, input *goquery.Selection) {
		name, _ := input.Attr("name")
		value, _ := input.Attr("value")
		form.Add(name, value)
	})
	form.Add("submit", "submit")

	res1, err := http.PostForm(actionUrl, form)
	if err != nil {
		println(err.Error())
		return ""
	}

	episodeServersDoc, _ := goquery.NewDocumentFromReader(res1.Body)
	defer res1.Body.Close()

	finalVideoUrl := ""
	episodeServersDoc.Find("#episode-servers li").
		EachWithBreak(func(i int, serverItem *goquery.Selection) bool {
			episodeServer, _ := serverItem.Find("a").First().Attr("data-ep-url")
			if episodeServer != "" {
				videoUrl := a.extractVideoUrlFromEpisodeServer(episodeServer)
				if videoUrl != "" {
					finalVideoUrl = videoUrl
					return false
				}
			}
			return true
		})

	return finalVideoUrl
}

func (a Anime4up) extractVideoUrlFromEpisodeServer(episodeServer string) string {
	u, _ := url.Parse(episodeServer)
	switch u.Host {
	case "voe.sx", "www.voe.sx":
		l := extractors.GetVideoFromVoe(episodeServer)
		if l != "" {
			return l
		}
	case "d000d.com":
		l, _ := extractors.GetUrlFromDownstream(episodeServer)
		if l != "" {
			return l
		}
	}
	return ""
}
