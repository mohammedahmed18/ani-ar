package api

import (
	"sync"

	"github.com/ani/ani-ar/types"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type AniService interface {
	SearchAnimeResults(string) []*types.AniResult
	GetResultById(string) *types.AniResult
	GetAllEpisodes(animeId string) []*types.AniEpisode
	GetSingleEpisode(animeId string, episodeNum int) *types.AniEpisode
}

// creates a new fiber app with template engine
// and setup middlewares
func InitApp() *fiber.App {
	f := fiber.New(fiber.Config{
		AppName:                 "Ani-ar",
		EnableTrustedProxyCheck: true,
		PassLocalsToViews:       true,
		EnableIPValidation:      true,
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
	})
	var once sync.Once

	once.Do(func() {
		f.Use(logger.New(logger.Config{
			Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		}))
		f.Use(recover.New())

	})

	return f
}
