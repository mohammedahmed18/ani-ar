package extractors

import (
	"encoding/base64"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func GetVideoFromVoe(link string) string {
	res, _ := http.Get(link)
	b, _ := io.ReadAll(res.Body)
	defer res.Body.Close()

	html := string(b)

	re := regexp.MustCompile(`window.location.href[\s+]=\s+'https(.*);`)
	matches := re.FindStringSubmatch(html)

	parts := strings.Split(matches[0], "=")
	forwardUrl := parts[1]
	forwardUrl = strings.TrimSpace(forwardUrl)
	forwardUrl = strings.TrimPrefix(forwardUrl, "'")
	forwardUrl = strings.TrimSuffix(forwardUrl, "';")

	req, _ := http.NewRequest("GET", forwardUrl, nil)
	// important to get the mp4 link
	req.Header.Add(
		"User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64; rv:129.0) Gecko/20100101 Firefox/129.0",
	)
	req.Header.Add("Accept", "text/html")
	res1, _ := http.DefaultClient.Do(req)
	b, _ = io.ReadAll(res1.Body)
	defer res1.Body.Close()

	html = string(b)
	re = regexp.MustCompile(`'mp4'[\s+]?:[\s+]?(.*)'`)
	matches = re.FindStringSubmatch(html)
	base64VideoUrl := strings.Split(matches[0], ":")[1]
	base64VideoUrl = strings.TrimSpace(base64VideoUrl)
	base64VideoUrl = strings.TrimPrefix(base64VideoUrl, "'")
	base64VideoUrl = strings.TrimSuffix(base64VideoUrl, "'")
	decoded, _ := base64.RawStdEncoding.DecodeString(base64VideoUrl)
	mp4Url := string(decoded)

	res2, _ := http.Get(mp4Url)
	if res2.StatusCode != 200 {
		return ""
	}
	return mp4Url
}
