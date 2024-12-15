package jellyfin

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
	"sync"
	"time"

	"github.com/ani/ani-ar/api"
	"github.com/ani/ani-ar/fetcher"
	"github.com/ani/ani-ar/types"
	"github.com/goccy/go-json"
	"github.com/kirsle/configdir"
)

var mu sync.Mutex

var remoteRevisionUrl string
var animeShowsPath string
var animeMoviesPath string

func init() {
	remoteRevisionUrl = os.Getenv("ANI_AR_REMOTE_REVISION_RAW_URL")
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

	CanBeEnhanced bool `json:"canBeEnhanced"`
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

func createFirstLocalRevisionFile() (*JellyfinRevision, error) {
	// config file doesn't exit
	err := os.MkdirAll(aniArConfigFolderPath, os.ModePerm)
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

func GetAndParseLocalRevision() (*JellyfinRevision, error) {
	if _, err := os.Stat(revisionFilePath); errors.Is(err, os.ErrNotExist) {
		createFirstLocalRevisionFile()
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
	if string(byteValue) == "" {
		return createFirstLocalRevisionFile()
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

func GetRemoteRevision() (*JellyfinRevision, error) {
	response, err := http.Get(remoteRevisionUrl)
	if err != nil {
		return nil, errors.New("Error while requesting the remote revision file, reason: " + err.Error())
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("Error while reading the remote revision file, reason: " + err.Error())

	}
	var remoteRev *JellyfinRevision
	err = json.Unmarshal(bytes, &remoteRev)
	if err != nil {
		return nil, errors.New("Error while parsing the remote revision file, reason: " + err.Error())
	}

	if remoteRev.RevisionId != "" {
		log.Printf("fetched the remote revision config with the id %s\n", remoteRev.RevisionId)
	}

	return remoteRev, nil
}

func DiffRevisions(old, new *JellyfinRevision) []*JellyfinRevisionDiff {
	log.Println("begin proccessing the diffs between the local and remote revisions...")
	// TODO: should we use revision id or no??
	// if old.RevisionId == new.RevisionId {
	// 	log.Println("same revision id, no diffs to check")
	// 	return []*JellyfinRevisionDiff{}
	// }

	var diffs []*JellyfinRevisionDiff
	checkedNewRevisions := make(map[string]interface{})
	for oldIdx, localRevItem := range old.Items {
		// search for the match
		found := false
		for _, remoteRevItem := range new.Items {
			if localRevItem.ID == remoteRevItem.ID {
				checkedNewRevisions[remoteRevItem.ID] = ""
				found = true
				// (same id exist in both revisions) check for updates
				updateDiff := &JellyfinRevisionDiff{
					Mode: fmt.Sprintf("UPDATE:%v", oldIdx),
				}
				changed := false
				if localRevItem.Type != remoteRevItem.Type {
					updateDiff.KeyNum = revKeyType
					updateDiff.New = remoteRevItem.Type
					changed = true
				}
				if localRevItem.Res != remoteRevItem.Res {
					updateDiff.KeyNum = revKeyRes
					updateDiff.New = remoteRevItem.Res
					changed = true
				}

				if remoteRevItem.Season == 0 {
					remoteRevItem.Season = 1
				}

				if localRevItem.Season != remoteRevItem.Season {
					updateDiff.KeyNum = revKeySeason
					updateDiff.New = fmt.Sprintf("%v", remoteRevItem.Season)
					changed = true
				}

				if changed {
					log.Printf("[diif] revision item %s got changed\n", remoteRevItem.ID)
					diffs = append(diffs, updateDiff)
				}
			}
		}
		if !found {
			// old rev was not found in the new, delete it
			log.Printf("[diif] revision item %s deleted\n", localRevItem.ID)
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
	localRev, err := GetAndParseLocalRevision()
	if err != nil {
		return err
	}

	remoteRev, err := GetRemoteRevision()
	if err != nil {
		return err
	}

	if remoteRev == nil {
		return errors.New("remote revision config can't be found, make sure you have set `ANI_AR_REMOTE_GIST` environment variable correctly ")
	}

	diffs := DiffRevisions(localRev, remoteRev)
	if len(diffs) == 0 {
		log.Println("No diffs to to perform, all good")
		return nil
	}

	updatedRev, err := ProcessDiff(diffs, localRev, remoteRev.RevisionId)
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
	return fmt.Sprintf("%s (%v).strm", r.Details.Title, r.Details.Aired.Prop.From.Year)
}

func downloadEpisode(aniEpisode *types.AniEpisode, filePath string, res string) error {
	medias := aniEpisode.GetPlayersWithQuality()
	resFound := false
	selectedSrc := ""
	for _, media := range medias {
		if media.Res == res {
			resFound = true
			selectedSrc = media.Res
		}
	}
	if !resFound && len(medias) > 0 {
		selectedSrc = medias[0].Res
	}
	if selectedSrc == "" {
		log.Printf("looks like there is no available links for %s episode %v, skipping", aniEpisode.Anime.DisplayName, aniEpisode.Number)
		return nil
	}
	err := os.WriteFile(filePath, []byte(medias[0].Src), 0755)
	if err != nil {
		return err
	}
	return nil
}
func getEnhancedResultForJellyfin(revItem *JellyfinRevisionItem) (*api.EnhancedAnimeResult, error) {
	fetcher := fetcher.GetDefaultFetcher()

	var enhancedAnimeResult *api.EnhancedAnimeResult
	if revItem.CanBeEnhanced {
		r, err := api.GetAnimeEnhancedResults(revItem.ID, fetcher)
		if err != nil {
			return nil, err
		}
		enhancedAnimeResult = r
	} else {
		// some revitem ids can't be enhanced because it has random string as id, instead of reaadable title or mal id
		anime := fetcher.GetAnimeResult(revItem.ID)

		enhancedAnimeResult = &api.EnhancedAnimeResult{
			Data: anime,
		}
		enhancedAnimeResult.Details = &api.JikanAnimeInfo{
			Title: anime.DisplayName,
			Type:  revItem.Type,
		}
		// TODO: fix this year
		enhancedAnimeResult.Details.Aired.Prop.From.Year = 2005
	}

	return enhancedAnimeResult, nil
}

func RemoveJellyfinMedia(revItem *JellyfinRevisionItem) error {
	enhancedAnimeResult, err := getEnhancedResultForJellyfin(revItem)
	if err != nil {
		return err
	}
	animePath := ""
	if enhancedAnimeResult.Details.Type == "TV" {
		animePath = filepath.Join(animeShowsPath, enhancedAnimeResult.Details.Title)
	} else if enhancedAnimeResult.Details.Type == "Movie" {
		animePath = filepath.Join(animeMoviesPath, getFormattedAnimeMovieName(enhancedAnimeResult))
	}

	err = os.RemoveAll(animePath)
	if err != nil {
		return err
	}
	return nil
}

func AddJellyfinMedia(revItem *JellyfinRevisionItem) error {
	fetcher := fetcher.GetDefaultFetcher()
	log.Printf("Adding new media item, id: [%s]\n", revItem.ID)

	enhancedAnimeResult, err := getEnhancedResultForJellyfin(revItem)
	if err != nil {
		return err
	}

	if enhancedAnimeResult.Details.Type != revItem.Type {
		return fmt.Errorf("the anime %s is of type %s not %s, make sure you are specifing the correct anime", enhancedAnimeResult.Details.Title, enhancedAnimeResult.Details.Type, revItem.Type)
	}

	isShow := enhancedAnimeResult.Details.Type == "TV"
	isMovie := enhancedAnimeResult.Details.Type == "Movie"

	animePath := ""

	// season will only be used for shows
	season := revItem.Season
	if season == 0 {
		season = 1
	}

	if isShow {
		animePath = filepath.Join(animeShowsPath, enhancedAnimeResult.Details.Title, fmt.Sprintf("Season %v", season))
	} else if isMovie {
		animePath = filepath.Join(animeMoviesPath)
	}

	log.Printf("target download folder: %s", animePath)

	err = os.MkdirAll(animePath, 0755)
	if err != nil {
		return err
	}

	// adding episodes
	fetcherEpisodes := fetcher.GetEpisodes(*enhancedAnimeResult.Data)
	for episodeIdx := range enhancedAnimeResult.Data.Episodes {
		log.Printf("adding episode [%v] of [%s]\n", episodeIdx+1, enhancedAnimeResult.Details.Title)

		fetcherEpisode := fetcherEpisodes[episodeIdx]
		episodePath := ""
		if isShow {
			episodeFileName := fmt.Sprintf("%s S%vE%v.strm", enhancedAnimeResult.Details.Title, season, episodeIdx+1)
			episodePath = filepath.Join(animePath, episodeFileName)
		}
		if isMovie {
			episodePath = filepath.Join(animeMoviesPath, getFormattedAnimeMovieName(enhancedAnimeResult))
		}
		downloadEpisode(&fetcherEpisode, episodePath, revItem.Res)
	}

	return nil

}

func RefreshLocalMediaItems() error {
	mu.Lock()
	defer mu.Unlock()

	localRevision, err := GetAndParseLocalRevision()
	if err != nil {
		return err
	}

	var fakeDiffs []*JellyfinRevisionDiff
	for _, localRevItem := range localRevision.Items {
		b, err := json.Marshal(localRevItem)
		if err != nil {
			return err
		}

		fakeDiffs = append(fakeDiffs, &JellyfinRevisionDiff{
			Mode: "ADD",
			New:  string(b),
		})
	}

	_, err = ProcessDiff(fakeDiffs, localRevision, localRevision.RevisionId)
	if err != nil {
		return err
	}
	return nil
}

func InfiniteLoop() error {
	log.Println("ANI_AR_REMOTE_REVISION_RAW_URL: " + remoteRevisionUrl)
	log.Println("ANI_AR_ANIME_SHOWS_FOLDER_PATH: " + animeShowsPath)
	log.Println("ANI_AR_ANIME_MOVIES_FOLDER_PATH: " + animeMoviesPath)
	println(`
	
		 /$$$$$$            /$$                             
		/$$__  $$          |__/                             
		| $$  \ $$ /$$$$$$$  /$$         /$$$$$$   /$$$$$$  
		| $$$$$$$$| $$__  $$| $$ /$$$$$$|____  $$ /$$__  $$ 
		| $$__  $$| $$  \ $$| $$|______/ /$$$$$$$| $$  \__/ 
		| $$  | $$| $$  | $$| $$        /$$__  $$| $$       
		| $$  | $$| $$  | $$| $$       |  $$$$$$$| $$       
		|__/  |__/|__/  |__/|__/        \_______/|__/       
															
	`)

	// TODO: let the user choose the interval
	go func() {
		ticker := time.NewTicker(time.Hour * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Refreshing links for local media items")
				err := RefreshLocalMediaItems()
				if err != nil {
					log.Printf("Error refreshing media items: %v", err)
				}
			}
		}
	}()

	// Revision loop runs every 5 minutes
	interval := time.Minute * 5
	for {
		mu.Lock()
		err := PerformRevision()
		mu.Unlock()
		if err != nil {
			log.Printf("Error during PerformRevision: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		time.Sleep(interval)
	}
}
