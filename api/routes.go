package api

import (
	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
	"github.com/gofiber/fiber/v2"
)

func InitiateRoutes(app *fiber.App) {
	fetcher := fetcher.GetDefaultFetcher()
	var jikan JikanApi

	app.Get(searchAniResultsBaseUrl, func(c *fiber.Ctx) error {
		search := c.Query("q")
		results := fetcher.Search(search)
		return c.JSON(results)
	})

	app.Get(getResultByIdUrl, func(c *fiber.Ctx) error {
		animeIdOrTitle := c.Params("animeId")
		anime := fetcher.GetAnimeResult(animeIdOrTitle)
		if anime == nil {
			return c.Send([]byte("Anime not found"))
		}
		type EnhancedAnimeResult struct {
			Details *JikanAnimeInfo
			Data    *types.AniResult
		}
		details := jikan.getBestMatchAnimeInfo(animeIdOrTitle)
		enhancedResult := &EnhancedAnimeResult{Data: anime}
		if details.Episodes == anime.Episodes {
			enhancedResult.Details = details
		}
		return c.JSON(enhancedResult)
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

		fetcherEpisodes := fetcher.GetEpisodes(*anime)
		return c.JSON(fetcherEpisodes)
	})

}
