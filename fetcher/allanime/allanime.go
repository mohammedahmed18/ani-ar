package allanime

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ani/ani-ar/types"
	"github.com/goccy/go-json"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type AllAnimeFetcher struct{}

const allanimeApi = "https://api.allanime.day"

const subType = "sub"

// TODO: will we support dubbed animes ??
// const dubType = "dub"

func GetAllAnimeFetcher() *AllAnimeFetcher {
	return &AllAnimeFetcher{}
}

func (a *AllAnimeFetcher) Search(q string) []types.AniResult {
	vars := AllAnimeSearchVariables{
		Limit:           40,
		Page:            1,
		Search:          AllAnimeSearch{Query: q},
		TranslationType: subType,
	}
	query := `query($search: SearchInput, $limit: Int, $page: Int, $translationType: VaildTranslationTypeEnumType, $countryOrigin: VaildCountryOriginEnumType) {
		shows(search: $search, limit: $limit, page: $page, translationType: $translationType, countryOrigin: $countryOrigin) {
			edges {
				_id
				name
				availableEpisodes
				episodeCount
				thumbnail
			}
		}
	}`

	b, err := makeGraphqlRequest(query, vars)
	if err != nil {
		log.Printf("Error while sending graphql response %v\n", err)
		return nil
	}

	var decodedResponse AllAnimeSearchResponse
	err = json.Unmarshal(b, &decodedResponse)
	if err != nil {
		log.Printf("Error while decoding the response body %v\n", err)
		return make([]types.AniResult, 0)
	}
	var results []types.AniResult

	for _, item := range decodedResponse.Data.Shows.Edges {
		availableEpisodes, found := item.AvailableEpisodes[vars.TranslationType]
		if !found {
			availableEpisodes, _ = strconv.Atoi(item.EpisodeCount)

		}
		results = append(results, types.AniResult{
			Id:           item.Id,
			DisplayName:  item.Name,
			Episodes:     availableEpisodes,
			DisplayCover: item.Thumbnail,
		})
	}
	return results
}

func (a *AllAnimeFetcher) GetAnimeResult(id string) *types.AniResult {
	vars := AllAnimeGetByIdVariables{
		Id: id,
	}
	query := `query($id: String!) {
    show(_id: $id) {
     _id
     name
     episodeCount
	 availableEpisodes
     thumbnail
    }
  }`

	b, err := makeGraphqlRequest(query, vars)
	if err != nil {
		log.Printf("Error while sending graphql response %v\n", err)
		return nil
	}

	var decodedResponse AllAnimeShowResponse
	err = json.Unmarshal(b, &decodedResponse)
	if err != nil {
		log.Printf("Error while decoding the response body %v\n", err)
		return nil
	}

	show := decodedResponse.Data.Show
	// TODO: recieve the translationType through vars
	availableEpisodes, found := show.AvailableEpisodes[subType]
	if !found {
		availableEpisodes, _ = strconv.Atoi(show.EpisodeCount)

	}

	return &types.AniResult{
		Id:           decodedResponse.Data.Show.Id,
		DisplayName:  decodedResponse.Data.Show.Name,
		DisplayCover: decodedResponse.Data.Show.Thumbnail,
		Episodes:     availableEpisodes,
	}

}

func makeGraphqlRequest(query string, variables interface{}) (queryResult []byte, e error) {
	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	res, err := http.Post(allanimeApi+"/api", "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return nil, fmt.Errorf("error while searching %v", err)
	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading the response body %v", err)
	}
	return b, nil
}

func extractVideoLinks(response []byte) ([]types.AniVideo, error) {
	var episodeResp AllAnimeEpisodeResponse
	err := json.Unmarshal(response, &episodeResp)
	if err != nil {
		return nil, err
	}
	var videos []types.AniVideo
	for _, source := range episodeResp.Data.Episode.SourceUrls {
		// //////////////////// S-mp4 source ///////////////////////
		if source.SourceName == "S-mp4" {
			videoId := strings.Split(source.Downloads.DownloadUrl, "id=")[1]
			downloadUrl := fmt.Sprintf("https://allanime.day/apivtwo/clock.json?id=%s", videoId)
			resp, err := http.Get(downloadUrl)
			if err != nil {
				return nil, err
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var linksResponse AllAnimeEpisodeLinksResponse
			err = json.Unmarshal(b, &linksResponse)
			if err != nil {
				return nil, err
			}

			for _, link := range linksResponse.Links {
				height := link.ResolutionStr
				// ctx, cancelFn := context.WithTimeout(context.Background(), 9*time.Second)
				// defer cancelFn()
				// data, err := ffprobe.ProbeURL(ctx, link.Src)
				// if err != nil {
				// 	return nil, err
				// }
				// height := fmt.Sprintf("%v", getHightFromFfprobe(data))
				// if height == "0" {
				// 	height = link.ResolutionStr
				// }

				videos = append(videos, types.AniVideo{Src: link.Src, Res: height})
			}
		}
		// //////////////////// S-mp4 source ///////////////////////

	}
	return videos, nil

}
func getHightFromFfprobe(data *ffprobe.ProbeData) int {
	for _, stream := range data.Streams {
		if stream.Height > 0 {
			return stream.Height
		}
	}
	return 0
}

func (a *AllAnimeFetcher) lazyLoadEpisodeVideos(r types.AniResult, episodeNum int) ([]types.AniVideo, error) {
	episodeEmbedGql := `query Episode($showId: String!, $episodeString: String!, $translationType: VaildTranslationTypeEnumType!) {
    episode(showId: $showId, episodeString: $episodeString, translationType: $translationType) {
      episodeString
      sourceUrls
    }
  }`

	variables := map[string]interface{}{
		"showId":          r.Id,
		"episodeString":   fmt.Sprintf("%v", episodeNum),
		"translationType": subType,
	}
	response, err := makeGraphqlRequest(episodeEmbedGql, variables)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch episode %d data: %v", episodeNum, err)
	}

	videos, err := extractVideoLinks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract video links for episode %d: %v", episodeNum, err)
	}

	// Sort the videos by resolution (highest first)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Res > videos[j].Res
	})

	return videos, nil
}

// GetEpisodes fetches the episodes for the given anime and returns them with lazy-loaded video links.
func (a *AllAnimeFetcher) GetEpisodes(r types.AniResult) []types.AniEpisode {
	var episodes []types.AniEpisode

	for i := range r.Episodes {
		epNo := i + 1

		episode := types.AniEpisode{
			Anime:  r,
			Number: epNo,
			GetPlayersWithQuality: func() []types.AniVideo {
				videos, err := a.lazyLoadEpisodeVideos(r, epNo)
				if err != nil {
					fmt.Println("Error fetching videos:", err)
					return nil
				}
				return videos
			},
			GetPlayerUrl: func() string {
				videos, err := a.lazyLoadEpisodeVideos(r, epNo)
				if err != nil || len(videos) == 0 {
					fmt.Println("Error fetching player URL:", err)
					return ""
				}
				return videos[0].Src
			},
		}

		episodes = append(episodes, episode)
	}

	return episodes
}
