package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Anime3rb struct{}

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

func (a *Anime3rb) search(key string) []AniResult {
	return a.searchPages(key, []AniResult{}, 1)
}

func (a *Anime3rb) searchPages(key string, results []AniResult, page int) []AniResult {
	searchUrl := fmt.Sprintf("https://anime3rb.com/search?q=%s&page=%v", url.QueryEscape(key), page)
	res, err := http.Get(searchUrl)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	queryResults := doc.Find(".search-results a")

	if queryResults == nil || queryResults.Length() == 0 || len(results) >= 20 {
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
		results = append(results, AniResult{
			title:       title,
			displayName: displayName,
			episodes:    episodes,
		})
	})

	return a.searchPages(key, results, page+1)
}

func (a *Anime3rb) getEpisodes(e AniResult) []AniEpisode {
	var episodes []AniEpisode
	for i := 0; i < e.episodes; i++ {
		episodeNum := i + 1
		episodes = append(episodes, AniEpisode{
			number: episodeNum,
			getUrl: getLazyEpisodeGetterFunc(e, episodeNum),
		})
	}
	return episodes
}

func getVideoUrl(html string, res ...string) string {
	re := regexp.MustCompile(`var\s+videos\s+=\s+\[((.|\n)*)\]`)
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
		return ""
	}
	for _, v := range videos {
		for _, r := range res {
			if v.Res == r {
				return v.Src
			}
		}
	}
	return videos[0].Src
}

func getLazyEpisodeGetterFunc(anime AniResult, episodeNum int) func() string {
	return func() string {
		url := fmt.Sprintf("https://anime3rb.com/episode/%s/%d", anime.title, episodeNum)
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
			vidUrl := getVideoUrl(string(b), "720", "480")
			return vidUrl
		} else {
			fmt.Println("No URL found")
			return ""
		}
	}
}
