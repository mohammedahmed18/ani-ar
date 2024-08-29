package main

// events

// events
/////////////////////////////////////////////////////////////////

type ChoicesLoadingEvent struct{}

func newChoicesLoadingEvent() ChoicesLoadingEvent {
	return ChoicesLoadingEvent{}
}

/////////////////////////////////////////////////////////////////

type ChoicesShownEvent struct {
	results []interface{}
}

func newChoicesShownEvent(results []interface{}) ChoicesShownEvent {
	return ChoicesShownEvent{
		results: results,
	}
}

/////////////////////////////////////////////////////////////////

type EpisodesLoadingEvent struct{}

func newEpisodesLoadingEvent() EpisodesLoadingEvent {
	return EpisodesLoadingEvent{}
}

// ///////////////////////////////////////////////////////////////
type EpisodesLoadedEvent struct {
	results []interface{}
}

func newEpisodesLoadedEvent(results []interface{}) EpisodesLoadedEvent {
	return EpisodesLoadedEvent{
		results: results,
	}
}
