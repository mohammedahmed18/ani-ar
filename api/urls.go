package api

const (
	baseUrl            = "/api"
	aniResultsBaseUrl  = baseUrl + "/ani-results"
	aniEpisodesBaseUrl = baseUrl + "/ani-episodes"

	searchAniResultsBaseUrl = aniResultsBaseUrl + "/search"
	getResultByIdUrl        = aniResultsBaseUrl + "/info/:animeId"

	getEpisodesBaseUrl      = aniEpisodesBaseUrl + "/:animeId/all"
	getSingleEpisodeBaseUrl = aniEpisodesBaseUrl + "/info/:episodeId"
)
