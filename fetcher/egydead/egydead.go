package egydead

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ani/ani-ar/types"
	"github.com/goccy/go-json"
)

const baseUrl = "https://egydeadw.sbs"

//go:embed browser_script/browser.py
var browserScript string

// TODO: use in memory cache
type EgydeadFetcher struct {
}

func GetEgyDeadFetcher() *EgydeadFetcher {
	return &EgydeadFetcher{}
}
func (e *EgydeadFetcher) Search(q string) []types.AniResult {
	return []types.AniResult{}
}
func (e *EgydeadFetcher) GetAnimeResult(id string) *types.AniResult {
	return nil
}

func (e *EgydeadFetcher) GetEpisodes(r types.AniResult) []types.AniEpisode {
	var episodes []types.AniEpisode
	for i := 0; i < r.Episodes; i++ {
		episodeNum := i + 1
		epUrl := fmt.Sprintf("%s/episode/%s", baseUrl, r.EpisodeIdFormatter(episodeNum))
		episodes = append(episodes, types.AniEpisode{
			Number:                episodeNum,
			GetPlayerUrl:          func() string { return e.getMediasForEpisode(epUrl)[0].Src },
			GetPlayersWithQuality: func() []types.AniVideo { return e.getMediasForEpisode(epUrl) },
			Url:                   epUrl,
			Anime:                 r,
		})
	}
	return episodes
}

func (e *EgydeadFetcher) getMediasForEpisode(epUrl string) []types.AniVideo {
	args := []string{"-", epUrl}
	c := exec.Command("python3", args...)
	c.Stdin = strings.NewReader(browserScript)
	output, err := c.Output()
	if err != nil {
		fmt.Println("error while executing the python script: ", err)
		return nil
	}
	var links []string
	jsonOutput := strings.ReplaceAll(string(output), "'", "\"")
	jsonOutput = strings.TrimSpace(jsonOutput)
	err = json.Unmarshal([]byte(jsonOutput), &links)
	if err != nil {
		fmt.Println("error while parsing python script output: ", err)
		return nil
	}
	if len(links) == 0 {
		fmt.Println("No available links for this episode ", err)
		return nil
	}
	episodeServerUrl := links[0]
	res, err := http.Get(episodeServerUrl)
	if err != nil {
		fmt.Printf("error while requesting the episode server page (%s) : %v\n", episodeServerUrl, err)
		return nil
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("error while reading response body : %v\n", err)
		return nil
	}
	html := string(bodyBytes)
	re := regexp.MustCompile(`eval(.*)\n+<\/script>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) == 0 {
		fmt.Println("can't find any matches for the eval script")
		return nil
	}
	script := matches[0]
	script = strings.TrimSuffix(script, "</script>")
	unpacker, _ := NewDEUnpacker(script)
	original, _ := unpacker.Unpack()

	re = regexp.MustCompile(`file:"(https?:\/\/[^"]+)`)

	matches = re.FindStringSubmatch(original)

	epFinalStreamUrl := strings.TrimPrefix(matches[0], "file:\"")
	println(epFinalStreamUrl)
	return []types.AniVideo{
		{
			Src: epFinalStreamUrl,
			Res: "High",
		},
	}
}
