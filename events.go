package main

// events

// search anime
type SearchAnimeEvent struct {
	animeResult []AniResult
}

func newSearchAnimeEvent(results []AniResult) SearchAnimeEvent {
	return SearchAnimeEvent{
		animeResult: results,
	}
}

// episode fetched
type EpisodeFetchedEvent struct {
	episodes []AniEpisode
}

func newEpisodeFetchedEvent(episodes []AniEpisode) EpisodeFetchedEvent {
	return EpisodeFetchedEvent{
		episodes: episodes,
	}
}
