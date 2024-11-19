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
	"time"

	"github.com/ani/ani-ar/api"
	"github.com/ani/ani-ar/fetcher"
	"github.com/goccy/go-json"
	"github.com/kirsle/configdir"
)

var remoteGistId string
var gistFileName string
var animeShowsPath string

func init() {
	remoteGistId = os.Getenv("ANI_AR_REMOTE_GIST_ID")
	gistFileName = os.Getenv("ANI_AR_REMOTE_GIST_FILE_NAME")
	animeShowsPath = os.Getenv("ANI_AR_ANIME_SHOWS_FOLDER_PATH")
}

// var animeMoviesPath = os.Getenv("ANI_AR_ANIME_MOVIES_FOLDER_PATH")

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

type ImportantGistRes struct {
	Files map[string]struct {
		FileName string `json:"fileName"`
		Content  string `json:"content"`
	} `json:"files"`
}

func GetAndParseLocalRevision() *JellyfinRevision {
	if _, err := os.Stat(revisionFilePath); errors.Is(err, os.ErrNotExist) {
		// config file doesn't exit
		err = os.MkdirAll(aniArConfigFolderPath, os.ModePerm)
		if err != nil {
			log.Println("couldn't create ani-ar config folder, reason :" + err.Error())
			return nil
		}
		_, err := os.Create(revisionFilePath)
		if err != nil {
			log.Println("couldn't create inital revision file, reason :" + err.Error())
			return nil
		}

		log.Println("revision file is created successfully")
	}
	revFile, err := os.Open(revisionFilePath)
	if err != nil {
		log.Println("couldn't read local revision file, reason :" + err.Error())
		return nil
	}
	defer revFile.Close()

	byteValue, _ := io.ReadAll(revFile)

	var rev JellyfinRevision
	json.Unmarshal([]byte(byteValue), &rev)

	return &rev
}

func GetRemoteRevision() *JellyfinRevision {
	gistApiUrl := fmt.Sprintf("https://api.github.com/gists/%s", remoteGistId)
	response, err := http.Get(gistApiUrl)
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
	var ghRes *ImportantGistRes
	err = json.Unmarshal(bytes, &ghRes)
	if err != nil {
		log.Println("Error while parsing the remote revision file, reason: " + err.Error())
		return nil
	}

	var remoteRev *JellyfinRevision
	json.Unmarshal([]byte(ghRes.Files[gistFileName].Content), &remoteRev)

	return remoteRev
}

func DiffRevisions(old, new *JellyfinRevision) []*JellyfinRevisionDiff {
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
					err := RemoveJellyfinTvFolder(&oldRevItem)
					err1 := AddJellyfinTvFolder(&newRevItem)
					if err == nil && err1 == nil {
						diffs = append(diffs, updateDiff)
					}
				}
			}
		}
		if !found {
			// old rev was not found in the new, delete it
			err := RemoveJellyfinTvFolder(&oldRevItem)
			if err == nil {
				diffs = append(diffs, &JellyfinRevisionDiff{
					Mode:   fmt.Sprintf("DEL:%v", oldIdx),
					KeyNum: revKeyId,
				})
			}

		}
	}

	// get the new added items (items that are in the new rev and not in the checkedNewRevisions)
	for _, newRevItem := range new.Items {
		_, found := checkedNewRevisions[newRevItem.ID]
		if !found {
			// add item
			err := AddJellyfinTvFolder(&newRevItem)
			if err == nil {
				encodedBytes, err := json.Marshal(newRevItem)
				if err != nil {
					log.Println("Error while encoding newly added rev item: " + err.Error())
					return nil
				}
				diffs = append(diffs, &JellyfinRevisionDiff{
					Mode: "ADD",
					New:  string(encodedBytes),
				})
			}

		}
	}
	return diffs
}

// process the diffs and create a new revision
func ProcessDiff(diffs []*JellyfinRevisionDiff, old *JellyfinRevision, newRevId string) *JellyfinRevision {
	var newRev = *old
	for _, diff := range diffs {
		if strings.HasPrefix(diff.Mode, "DEL") {
			modeParts := strings.Split(diff.Mode, ":")
			idxStr := modeParts[1]
			idx, _ := strconv.Atoi(idxStr)
			newRev.Items = append(old.Items[:idx], old.Items[idx+1:]...)
		} else if strings.HasPrefix(diff.Mode, "UPDATE") {
			modeParts := strings.Split(diff.Mode, ":")
			idxStr := modeParts[1]
			idx, _ := strconv.Atoi(idxStr)
			switch diff.KeyNum {
			case revKeyType:
				newRev.Items[idx].Type = diff.New
			case revKeyRes:
				newRev.Items[idx].Res = diff.New
			case revKeySeason:
				seasonNumber, _ := strconv.Atoi(diff.New)
				newRev.Items[idx].Season = seasonNumber
			}
		} else if diff.Mode == "ADD" {
			var newRevItem JellyfinRevisionItem
			json.Unmarshal([]byte(diff.New), &newRevItem)
			if newRevItem.Season == 0 {
				newRevItem.Season = 1
			}
			newRev.Items = append(newRev.Items, newRevItem)
		}
	}

	newRev.RevisionId = newRevId

	return &newRev
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
	currentRev := GetAndParseLocalRevision()
	remoteRev := GetRemoteRevision()

	if currentRev == nil {
		currentRev = &JellyfinRevision{}
	}
	if remoteRev == nil {
		currentRev = &JellyfinRevision{}
	}
	diffs := DiffRevisions(currentRev, remoteRev)
	updatedRev := ProcessDiff(diffs, currentRev, remoteRev.RevisionId)
	if len(diffs) == 0 {
		return nil
	}
	err := writeNewRevLocally(updatedRev)
	if err != nil {
		log.Println("error while writing the new rev config file, reason: " + err.Error())
		return err
	}
	b, _ := json.Marshal(updatedRev)
	println("revision done, content : " + string(b))
	return nil
}

func RemoveJellyfinTvFolder(revItem *JellyfinRevisionItem) error {
	fetcher := fetcher.GetDefaultFetcher()
	animeResult, err := api.GetAnimeEnhancedResults(revItem.ID, fetcher)
	if err != nil {
		return err
	}
	animePath := filepath.Join(animeShowsPath, animeResult.Details.TitleEnglish)
	err = os.RemoveAll(animePath)
	if err != nil {
		return err
	}
	return nil
}

func AddJellyfinTvFolder(revItem *JellyfinRevisionItem) error {
	fetcher := fetcher.GetDefaultFetcher()

	animeResult, err := api.GetAnimeEnhancedResults(revItem.ID, fetcher)
	if err != nil {
		return err
	}

	if animeResult.Details.Type != revItem.Type {
		return fmt.Errorf("the anime %s is of type %s not %s, make sure you are specifing the correct anime", animeResult.Details.TitleEnglish, animeResult.Details.Type, revItem.Type)
	}
	season := revItem.Season
	if season == 0 {
		season = 1
	}
	// animeResult.Details
	animePath := filepath.Join(animeShowsPath, animeResult.Details.TitleEnglish, fmt.Sprintf("Season %v", season))
	err = os.MkdirAll(animePath, 0755)
	if err != nil {
		return err
	}

	// adding episodes
	for episodeIdx := range animeResult.Data.Episodes {
		fetcherEpisodes := fetcher.GetEpisodes(*animeResult.Data)

		fetcherEpisode := fetcherEpisodes[episodeIdx]
		medias := fetcherEpisode.GetPlayersWithQuality()
		for _, media := range medias {
			if media.Res == revItem.Res {
				episodeFileName := fmt.Sprintf("%s S%vE%v.strm", animeResult.Details.TitleEnglish, season, episodeIdx+1)
				err = os.WriteFile(filepath.Join(animePath, episodeFileName), []byte(media.Src), 0755)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil

}

func InfiniteLoop() error {
	log.Println("ANI_AR_REMOTE_GIST_ID: " + remoteGistId)
	log.Println("ANI_AR_REMOTE_GIST_FILE_NAME: " + gistFileName)
	log.Println("ANI_AR_ANIME_SHOWS_FOLDER_PATH: " + animeShowsPath)
	for {
		err := PerformRevision()
		if err != nil {
			return err
		}
		time.Sleep(time.Minute * 5)
	}
}
