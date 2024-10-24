package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/ani/ani-ar/types"
)

type Anime3rb struct{}

func getAnime3rbFetcher() Fetcher {
	return &Anime3rb{}
}

type Ani3rbVideo struct {
	Src string `json:"src"`
	Res string `json:"res"`
}

const baseUrl = "https://anime3rb.com"

func (a *Anime3rb) getToken() string {
	res, err := http.Get(baseUrl)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	b, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	html := string(b)

	re := regexp.MustCompile(`<meta\s+name="csrf-token"\s+content="([^"]+)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	} else {
		fmt.Println("CSRF Token not found")
		return ""
	}
}

func (a *Anime3rb) GetAnimeResult(title string) *types.AniResult {
	displayNameRe := regexp.MustCompile(
		`<h1\s+class="text-2xl font-bold uppercase inline">(.*)<\/h1>`,
	)
	episodesRe := regexp.MustCompile(`<p class="(.*)">الحلقات<\/p>\n+<p(.*)<\/p>`)
	animePageUrl := fmt.Sprintf("%s/titles/%s", baseUrl, title)
	res, err := http.Get(animePageUrl)
	if res.StatusCode != 200 || err != nil {
		return nil
	}
	log.Println("found anime page : status 200 OK")
	defer res.Body.Close()
	htmlBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil
	}
	log.Println("parsing the html document to extract info...")
	displayNameMatches := displayNameRe.FindStringSubmatch(string(htmlBytes))
	episodeNumberMatches := episodesRe.FindStringSubmatch(string(htmlBytes))

	displayNameDoc, err := goquery.NewDocumentFromReader(strings.NewReader(displayNameMatches[0]))
	if err != nil {
		return nil
	}
	episodeNumberDoc, err := goquery.NewDocumentFromReader(
		strings.NewReader(episodeNumberMatches[0]),
	)
	if err != nil {
		return nil
	}

	displayName := displayNameDoc.Find("span:nth-child(1)").Text()
	episodesCount := episodeNumberDoc.Find("p:nth-child(2)").Text()
	epCoutnInt, _ := strconv.Atoi(episodesCount)

	r := &types.AniResult{
		Title:       title,
		DisplayName: displayName,
		Episodes:    epCoutnInt,
	}
	log.Printf("found anime info %v", r)
	return r
}

func (a *Anime3rb) Search(key string) []types.AniResult {
	log.Println("searching...")
	return a.searchPages(key, []types.AniResult{}, 1)
}

func (a *Anime3rb) searchPages(
	key string,
	results []types.AniResult,
	page int,
) []types.AniResult {
	if len(results) > 20 {
		return results
	}
	searchUrl := fmt.Sprintf("%s/search?q=%s&page=%v", baseUrl, url.QueryEscape(key), page)
	res, err := http.Get(searchUrl)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	queryResults := doc.Find(".search-results a")

	if queryResults == nil || queryResults.Length() == 0 {
		return results
	}
	queryResults.Each(func(i int, result *goquery.Selection) {
		// For each item found, get the title
		displayName := result.Find("h4").Text()

		animeUrl, _ := result.Attr("href")
		parts := strings.Split(animeUrl, "/")
		title := parts[len(parts)-1]

		var episodes int = -1

		result.Find("span").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			// Check if the text contains "حلقات" and extract the number
			if strings.Contains(text, "حلقات") {
				// Use a regular expression to extract the number before "حلقات"
				re := regexp.MustCompile(`(\d+)\s*حلقات`)
				matches := re.FindStringSubmatch(text)
				if len(matches) > 1 {
					episodes, _ = strconv.Atoi(matches[1])
				}
			}
		})
		results = append(results, types.AniResult{
			Title:       title,
			DisplayName: displayName,
			Episodes:    episodes,
		})
	})

	if page == 3 {
		return results
	}
	return a.searchPages(key, results, page+1)
}

func (a *Anime3rb) GetEpisodes(e types.AniResult) []types.AniEpisode {
	var episodes []types.AniEpisode
	for i := 0; i < e.Episodes; i++ {
		episodeNum := i + 1
		epUrl := fmt.Sprintf("%s/episode/%s/%d", baseUrl, e.Title, episodeNum)
		episodes = append(episodes, types.AniEpisode{
			Number:       episodeNum,
			GetPlayerUrl: getLazyEpisodeGetterFunc(&e, epUrl),
			Url:          epUrl,
			Anime:        e,
		})
	}
	return episodes
}

func getVideoUrl(html string, res ...string) (string, string) {
	re := regexp.MustCompile(`var\s+videos\s+=\s+\[((.|\n)*)\},\n+\s+\]`)
	// Find the match
	match := re.FindStringSubmatch(html)
	parts := strings.Split(match[0], "videos =")
	stringifyArray := parts[1]
	attrs := []string{"src", "type", "label", "res"}
	stringifyArray = strings.ReplaceAll(stringifyArray, "'", "\"")
	for _, attr := range attrs {
		stringifyArray = strings.ReplaceAll(stringifyArray, attr, fmt.Sprintf("\"%s\"", attr))
	}
	// remove trailing comma from last object
	re = regexp.MustCompile(`\},\n+\s+\]`)
	stringifyArray = re.ReplaceAllString(stringifyArray, "}]")

	var videos []Ani3rbVideo
	err := json.Unmarshal([]byte(stringifyArray), &videos)
	if err != nil {
		fmt.Println("error while parsing videos " + err.Error())
		return "", ""
	}
	for _, v := range videos {
		for _, r := range res {
			if v.Res == r {
				return v.Src, v.Res
			}
		}
	}
	return videos[0].Src, videos[0].Res
}

func getLazyEpisodeGetterFunc(anime *types.AniResult, url string) func() string {
	return func() string {
		res, err := http.Get(url)
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}

		resBytes, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
		html := string(resBytes)
		re := regexp.MustCompile(`videoSource:\s*'([^']+)'`)
		// Find the match
		match := re.FindStringSubmatch(html)
		if len(match) > 1 {
			// Extracted URL
			url := match[1]
			// Replace escaped characters
			unescapedURL := strings.ReplaceAll(url, `\/`, `/`)
			unescapedURL = strings.ReplaceAll(unescapedURL, `\u0026`, `&`)

			res, _ := http.Get(unescapedURL)
			b, _ := io.ReadAll(res.Body)
			defer res.Body.Close()
			vidUrl, finalRes := getVideoUrl(string(b), "720")
			anime.SelectedQuality = finalRes
			return vidUrl
		} else {
			fmt.Println("No URL found")
			return ""
		}
	}
}
