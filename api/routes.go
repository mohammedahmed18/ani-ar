package api

import (
	"errors"
	"strconv"
	"time"

	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

func InitiateRoutes(app *fiber.App) {
	fetcher := fetcher.GetDefaultFetcher()
	var jikan *JikanApi = &JikanApi{
		C: cache.New(time.Minute*5, time.Minute*10),
	}

	app.Get(searchAniResultsBaseUrl, func(c *fiber.Ctx) error {
		search := c.Query("q")
		results := fetcher.Search(search)
		return c.JSON(results)
	})

	app.Get(getResultByIdUrl, func(c *fiber.Ctx) error {
		animeId := c.Params("animeId")
		enhanced, err := getAnimeEnhancedResults(animeId, fetcher)
		if err != nil {
			return err
		}
		return c.JSON(enhanced)
	})

	app.Get(getEpisodesBaseUrl, func(c *fiber.Ctx) error {
		animeIdOrTitle := c.Params("animeId")
		anime := fetcher.GetAnimeResult(animeIdOrTitle)
		if anime == nil {
			return c.Send([]byte("Anime not found"))
		}
		bestMatch := jikan.getBestMatchAnimeInfo(animeIdOrTitle)
		if bestMatch != nil {
			if bestMatch.Episodes == anime.Episodes {
				// 95% it's the same anime
				episodes := jikan.getEpisodes(bestMatch.MalID)
				return c.JSON(episodes)
			}

		}

		// if no match we can return the fetcher episodes instead
		fetcherEpisodes := fetcher.GetEpisodes(*anime)
		return c.JSON(fetcherEpisodes)
	})
	app.Get(getSingleEpisodeBaseUrl, func(c *fiber.Ctx) error {
		animeIdOrTitle := c.Params("animeId")
		episodeNumParam := c.Params("episodeNum")

		fetcherAnime := fetcher.GetAnimeResult(animeIdOrTitle)
		if fetcherAnime == nil {
			return c.Send([]byte("Anime not found"))
		}

		episodeNum, err := strconv.Atoi(episodeNumParam)
		if err != nil {
			return c.Status(400).JSON(map[string]string{"message": "invalid episode number"})
		}

		bestMatch := jikan.getBestMatchAnimeInfo(animeIdOrTitle)
		jikanEpisode := jikan.getSingleEpisode(bestMatch.MalID, episodeNum)
		fetcherEpisodes := fetcher.GetEpisodes(*fetcherAnime)

		fetcherEpisode := fetcherEpisodes[episodeNum-1]
		medias := fetcherEpisode.GetPlayersWithQuality()

		type EpisodeType struct {
			ArMedias []types.AniVideo `json:"arMediaUrl"`
			// TODO: add more episode info
		}
		type EnhancedEpisodeType struct {
			MalAnime   *JikanAnimeInfo    `json:"malAnime"`
			MalEpisode *JikanAnimeEpisode `json:"malEpisode"`
			Episode    *EpisodeType       `json:"episode"`
		}
		return c.JSON(&EnhancedEpisodeType{
			MalAnime:   bestMatch,
			MalEpisode: jikanEpisode,
			Episode: &EpisodeType{
				ArMedias: medias,
			},
		})
	})

}

type EnhancedAnimeResult struct {
	Details *JikanAnimeInfo
	Data    *types.AniResult
}

func getAnimeEnhancedResults(animeIdOrTitle string, fetcher fetcher.Fetcher) (*EnhancedAnimeResult, error) {
	var jikan JikanApi
	anime := fetcher.GetAnimeResult(animeIdOrTitle)
	if anime == nil {
		return nil, errors.New("anime not found")
	}
	details := jikan.getBestMatchAnimeInfo(animeIdOrTitle)
	enhancedResult := &EnhancedAnimeResult{Data: anime}
	if details.Episodes == anime.Episodes {
		enhancedResult.Details = details
	}
	return enhancedResult, nil
}
