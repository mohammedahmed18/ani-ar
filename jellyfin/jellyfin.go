package jellyfin

// TODO: add more debug logs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ani/ani-ar/api"
	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
	"github.com/goccy/go-json"
	"github.com/kirsle/configdir"
)

var remoteGist string
var animeShowsPath string
var animeMoviesPath string

func init() {
	remoteGist = os.Getenv("ANI_AR_REMOTE_GIST")
	animeShowsPath = os.Getenv("ANI_AR_ANIME_SHOWS_FOLDER_PATH")
	animeMoviesPath = os.Getenv("ANI_AR_ANIME_MOVIES_FOLDER_PATH")
}

var aniArConfigFolderPath = filepath.Join(configdir.LocalConfig(), "ani-ar")
var revisionFilePath = filepath.Join(aniArConfigFolderPath, "rev.cfg")

type JellyfinRevisionItem struct {
	// anime id or title (for anime3rb fetcher)
	ID string `json:"id"`
	// either `TV`` or `Movie``
	Type string `json:"type"`
	// selected resolution, eg: "720" , "1080"
	Res string `json:"res"`
	// season number, eg: 1, 2 , 3 (default 1)
	Season int `json:"season"`
}

type JellyfinRevision struct {
	// revision id for last revision executed
	RevisionId string                 `json:"revisionId"`
	Items      []JellyfinRevisionItem `json:"items"`
}

const (
	revKeyId = iota
	revKeyType
	revKeyRes
	revKeySeason
)

type JellyfinRevisionDiff struct {
	// either ADD or DEL:<old_rev_item_idx> or UPDATE:<old_rev_item_idx>
	Mode string `json:"mode"`
	// rev filed number
	KeyNum int `json:"keyNum"`
	// new value
	New string `json:"new"`
}

func GetAndParseLocalRevision() (*JellyfinRevision, error) {
	if _, err := os.Stat(revisionFilePath); errors.Is(err, os.ErrNotExist) {
		// config file doesn't exit
		err = os.MkdirAll(aniArConfigFolderPath, os.ModePerm)
		if err != nil {
			return nil, errors.New("couldn't create ani-ar config folder, reason :" + err.Error())
		}
		err = os.WriteFile(revisionFilePath, []byte(""), 0755)
		if err != nil {
			return nil, errors.New("couldn't create inital revision file, reason :" + err.Error())
		}

		log.Println("initial revision file is created successfully")
		return &JellyfinRevision{
			RevisionId: "(((((0)))))",
		}, nil
	}

	revFile, err := os.Open(revisionFilePath)
	if err != nil {
		return nil, errors.New("couldn't read local revision file, reason :" + err.Error())
	}

	defer revFile.Close()

	byteValue, err := io.ReadAll(revFile)
	if err != nil {
		return nil, errors.New("couldn't read local revision file, reason :" + err.Error())
	}

	var rev JellyfinRevision
	err = json.Unmarshal([]byte(byteValue), &rev)
	if err != nil {
		return nil, errors.New("couldn't parse the local revision file, reason :" + err.Error())
	}

	if rev.RevisionId != "" {
		log.Printf("parsed the local revision config with the id %s\n", rev.RevisionId)
	}

	return &rev, nil
}

func GetRemoteRevision() *JellyfinRevision {
	gistRawContentUrl := fmt.Sprintf("https://gist.githubusercontent.com/%s/raw", remoteGist)
	response, err := http.Get(gistRawContentUrl)
	if err != nil {
		log.Println("Error while requesting the remote revision file, reason: " + err.Error())
		return nil
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Error while reading the remote revision file, reason: " + err.Error())
		return nil
	}
	var remoteRev *JellyfinRevision
	json.Unmarshal(bytes, &remoteRev)

	if remoteRev.RevisionId != "" {
		log.Printf("fetched the remote revision config with the id %s\n", remoteRev.RevisionId)
	}

	return remoteRev
}

func DiffRevisions(old, new *JellyfinRevision) []*JellyfinRevisionDiff {
	log.Println("begin proccessing the diffs between the local and remote revisions...")
	if old.RevisionId == new.RevisionId {
		log.Println("same revision id, no diffs to check")
		return []*JellyfinRevisionDiff{}
	}

	var diffs []*JellyfinRevisionDiff
	var checkedNewRevisions map[string]interface{} = make(map[string]interface{})
	for oldIdx, oldRevItem := range old.Items {
		// search for the match
		found := false
		for _, newRevItem := range new.Items {
			if oldRevItem.ID == newRevItem.ID {
				checkedNewRevisions[newRevItem.ID] = ""
				found = true
				// (same id exist in both revisions) check for updates
				updateDiff := &JellyfinRevisionDiff{
					Mode: fmt.Sprintf("UPDATE:%v", oldIdx),
				}
				changed := false
				if oldRevItem.Type != newRevItem.Type {
					updateDiff.KeyNum = revKeyType
					updateDiff.New = newRevItem.Type
					changed = true
				}
				if oldRevItem.Res != newRevItem.Res {
					updateDiff.KeyNum = revKeyRes
					updateDiff.New = newRevItem.Res
					changed = true
				}
				if oldRevItem.Season != 0 && newRevItem.Season == 0 {
					// the season number was removed from the remote revision config, then we will set this to season 1
					updateDiff.KeyNum = revKeySeason
					updateDiff.New = fmt.Sprintf("%v", 1)
					changed = true
				}
				if oldRevItem.Season != newRevItem.Season {
					updateDiff.KeyNum = revKeySeason
					updateDiff.New = fmt.Sprintf("%v", newRevItem.Season)
					changed = true
				}
				if changed {
					log.Printf("[diif] revision item %s got changed\n", newRevItem.ID)
					diffs = append(diffs, updateDiff)
				}
			}
		}
		if !found {
			// old rev was not found in the new, delete it
			log.Printf("[diif] revision item %s deleted\n", oldRevItem.ID)
			diffs = append(diffs, &JellyfinRevisionDiff{
				Mode:   fmt.Sprintf("DEL:%v", oldIdx),
				KeyNum: revKeyId,
			})

		}
	}

	// get the new added items (items that are in the new rev and not in the checkedNewRevisions)
	for _, newRevItem := range new.Items {
		_, found := checkedNewRevisions[newRevItem.ID]
		if !found {
			// add item
			encodedBytes, err := json.Marshal(newRevItem)
			if err != nil {
				log.Println("Error while encoding newly added rev item: " + err.Error())
				return nil
			}
			log.Printf("[diif] revision item %s added\n", newRevItem.ID)
			diffs = append(diffs, &JellyfinRevisionDiff{
				Mode: "ADD",
				New:  string(encodedBytes),
			})

		}
	}
	return diffs
}

func printDiffInfo(diffs []*JellyfinRevisionDiff) {
	addDiffsCount := 0
	updateDiffsCount := 0
	delDiffsCount := 0
	for _, diff := range diffs {
		if strings.HasPrefix(diff.Mode, "DEL") {
			delDiffsCount += 1
		} else if strings.HasPrefix(diff.Mode, "UPDATE") {
			updateDiffsCount += 1
		} else if diff.Mode == "ADD" {
			addDiffsCount += 1
		}
	}
	log.Printf("[%d] addition diffs, [%d] update diffs, [%d] delete diffs", addDiffsCount, updateDiffsCount, delDiffsCount)
}

// process the diffs and create a new revision
func ProcessDiff(diffs []*JellyfinRevisionDiff, old *JellyfinRevision, newRevId string) (*JellyfinRevision, error) {
	log.Println("Start proccessing the diffs...")
	printDiffInfo(diffs)

	var newRev = *old
	for _, diff := range diffs {
		if strings.HasPrefix(diff.Mode, "DEL") {
			modeParts := strings.Split(diff.Mode, ":")
			idxStr := modeParts[1]
			oldRevItemIdx, _ := strconv.Atoi(idxStr)

			err := RemoveJellyfinMedia(&old.Items[oldRevItemIdx])
			if err != nil {
				return nil, err
			}

			newRev.Items = append(old.Items[:oldRevItemIdx], old.Items[oldRevItemIdx+1:]...)
		} else if strings.HasPrefix(diff.Mode, "UPDATE") {
			modeParts := strings.Split(diff.Mode, ":")
			idxStr := modeParts[1]
			idx, _ := strconv.Atoi(idxStr)
			RemoveJellyfinMedia(&old.Items[idx])
			switch diff.KeyNum {
			case revKeyType:
				newRev.Items[idx].Type = diff.New
			case revKeyRes:
				newRev.Items[idx].Res = diff.New
			case revKeySeason:
				seasonNumber, _ := strconv.Atoi(diff.New)
				newRev.Items[idx].Season = seasonNumber
			}
			err := AddJellyfinMedia(&newRev.Items[idx])
			if err != nil {
				return nil, err
			}
		} else if diff.Mode == "ADD" {
			var newRevItem JellyfinRevisionItem
			json.Unmarshal([]byte(diff.New), &newRevItem)
			if newRevItem.Season == 0 {
				newRevItem.Season = 1
			}
			err := AddJellyfinMedia(&newRevItem)
			if err != nil {
				return nil, err
			}
			newRev.Items = append(newRev.Items, newRevItem)
		}
	}

	newRev.RevisionId = newRevId

	return &newRev, nil
}

func writeNewRevLocally(newRev *JellyfinRevision) error {
	file, err := os.OpenFile(revisionFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	if err != nil {
		return err
	}
	defer file.Close()
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	revBytes, err := json.Marshal(newRev)
	if err != nil {
		return err
	}

	_, err = file.Write(revBytes)
	if err != nil {
		return err
	}

	return nil
}

func PerformRevision() error {
	currentRev, err := GetAndParseLocalRevision()
	if err != nil {
		return err
	}

	remoteRev := GetRemoteRevision()

	if remoteRev == nil {
		return errors.New("remote revision config can't be found, make sure you have set `ANI_AR_REMOTE_GIST` environment variable correctly ")
	}

	diffs := DiffRevisions(currentRev, remoteRev)
	if len(diffs) == 0 {
		return nil
	}

	updatedRev, err := ProcessDiff(diffs, currentRev, remoteRev.RevisionId)
	if err != nil {
		return errors.New("error while processing the diffs, reason: " + err.Error())
	}

	err = writeNewRevLocally(updatedRev)
	if err != nil {
		return errors.New("error while writing the new rev config file, reason: " + err.Error())
	}
	b, _ := json.Marshal(updatedRev)
	println("revision done, content : " + string(b))
	return nil
}

func getFormattedAnimeMovieName(r *api.EnhancedAnimeResult) string {
	return fmt.Sprintf("%s (%v).strm", r.Details.TitleEnglish, r.Details.Aired.Prop.From.Year)
}

func downloadEpisode(aniEpisode *types.AniEpisode, filePath string, res string) error {
	medias := aniEpisode.GetPlayersWithQuality()
	for _, media := range medias {
		if media.Res == res {
			err := os.WriteFile(filePath, []byte(media.Src), 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RemoveJellyfinMedia(revItem *JellyfinRevisionItem) error {
	fetcher := fetcher.GetDefaultFetcher()
	animeResult, err := api.GetAnimeEnhancedResults(revItem.ID, fetcher)
	if err != nil {
		return err
	}
	animePath := ""
	if animeResult.Details.Type == "TV" {
		animePath = filepath.Join(animeShowsPath, animeResult.Details.TitleEnglish)
	} else if animeResult.Details.Type == "Movie" {
		animePath = filepath.Join(animeMoviesPath, getFormattedAnimeMovieName(animeResult))
	}

	err = os.RemoveAll(animePath)
	if err != nil {
		return err
	}
	return nil
}

func AddJellyfinMedia(revItem *JellyfinRevisionItem) error {
	fetcher := fetcher.GetDefaultFetcher()

	animeResult, err := api.GetAnimeEnhancedResults(revItem.ID, fetcher)

	log.Printf("adding a new media [%s] [%s]\n", animeResult.Details.TitleEnglish, animeResult.Details.Type)

	if err != nil {
		return err
	}

	if animeResult.Details.Type != revItem.Type {
		return fmt.Errorf("the anime %s is of type %s not %s, make sure you are specifing the correct anime", animeResult.Details.TitleEnglish, animeResult.Details.Type, revItem.Type)
	}

	isShow := revItem.Type == "TV"
	isMovie := revItem.Type == "Movie"

	animePath := ""

	// season will only be used for shows
	season := revItem.Season
	if season == 0 {
		season = 1
	}

	if isShow {
		animePath = filepath.Join(animeShowsPath, animeResult.Details.TitleEnglish, fmt.Sprintf("Season %v", season))
	} else if isMovie {
		animePath = filepath.Join(animeMoviesPath)
	}

	log.Printf("target download folder: %s", animePath)

	err = os.MkdirAll(animePath, 0755)
	if err != nil {
		return err
	}

	// adding episodes
	fetcherEpisodes := fetcher.GetEpisodes(*animeResult.Data)
	for episodeIdx := range animeResult.Data.Episodes {
		log.Printf("adding episode [%v] of [%s]\n", episodeIdx+1, animeResult.Details.TitleEnglish)

		fetcherEpisode := fetcherEpisodes[episodeIdx]
		episodePath := ""
		if isShow {
			episodeFileName := fmt.Sprintf("%s S%vE%v.strm", animeResult.Details.TitleEnglish, season, episodeIdx+1)
			episodePath = filepath.Join(animePath, episodeFileName)
		}
		if isMovie {
			episodePath = filepath.Join(animeMoviesPath, getFormattedAnimeMovieName(animeResult))
		}
		downloadEpisode(&fetcherEpisode, episodePath, revItem.Res)
	}

	return nil

}

func InfiniteLoop() error {
	log.Println("ANI_AR_REMOTE_GIST: " + remoteGist)
	log.Println("ANI_AR_ANIME_SHOWS_FOLDER_PATH: " + animeShowsPath)
	log.Println("ANI_AR_ANIME_MOVIES_FOLDER_PATH: " + animeMoviesPath)
	err := PerformRevision()
	if err != nil {
		return err
	}
	// time.Sleep(time.Minute * 5)
	return nil
}
