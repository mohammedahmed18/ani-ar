package api

import (
	"github.com/ani/ani-ar/fetcher"
	"github.com/gofiber/fiber/v2"
)

func InitiateRoutes(app *fiber.App) {
	fetcher := fetcher.GetDefaultFetcher()

	app.Get(searchAniResultsBaseUrl, func(c *fiber.Ctx) error {
		search := c.Query("q")
		results := fetcher.Search(search)
		return c.JSON(results)
	})

}
