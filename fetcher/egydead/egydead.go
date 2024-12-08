package egydead

import (
	"fmt"
	"log"

	"github.com/ani/ani-ar/types"
	"github.com/playwright-community/playwright-go"
)

const baseUrl = "https://egyrbyeteuh.sbs"

// test id:
//مسلسل-arcane-الموسم-الثاني-مترجم-كامل

// TODO: use in memory cache
type EgydeadFetcher struct {
}

func GetEgyDeadFetcher() *EgydeadFetcher {
	return &EgydeadFetcher{}
}
func (e *EgydeadFetcher) Search(q string) []types.AniResult {
	return []types.AniResult{}
}
func (e *EgydeadFetcher) GetAnimeResult(id string) *types.AniResult {
	return nil
}

func (e *EgydeadFetcher) GetEpisodes(r types.AniResult) []types.AniEpisode {
	var episodes []types.AniEpisode
	for i := 0; i < r.Episodes; i++ {
		episodeNum := i + 1
		epUrl := fmt.Sprintf("%s/episode/%s", baseUrl, r.EpisodeIdFormatter(episodeNum))
		episodes = append(episodes, types.AniEpisode{
			Number:                episodeNum,
			GetPlayerUrl:          func() string { return e.getMediasForEpisode(epUrl)[0].Src },
			GetPlayersWithQuality: func() []types.AniVideo { return e.getMediasForEpisode(epUrl) },
			Url:                   epUrl,
			Anime:                 r,
		})
	}
	return episodes
}
func (e *EgydeadFetcher) getMediasForEpisode(epUrl string) []types.AniVideo {
	playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	pw, err := playwright.Run()

	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	// bravePath := "/usr/bin/brave-browser"
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		// ExecutablePath: &bravePath,
	})

	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	// Setup a listener to capture the response
	page.On("response", func(response playwright.Response) {
		if response.Request().Method() == "POST" {
			println(response.Request().Method() + " :: " + response.URL() + " status : " + fmt.Sprintf("%v", response.Status()))
		}
		// htmlBytes, _ := response.Body()
		// println(string(htmlBytes))
	})

	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto(epUrl); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	watchBtnLocator := page.Locator(".watchNow form button").First()
	watchBtnLocator.WaitFor()

	watchBtnLocator.Click()
	// formdata := make(map[string]any)
	// formdata["View"] = "1"
	// page.Request().Post(epUrl, playwright.APIRequestContextPostOptions{
	// 	Form: formdata,
	// })

	page.Locator(".watchAreaMaster").WaitFor()

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

	// bypass, err := NewBypasser(WithBrowserMode(true))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return nil
	// }

	// client := &http.Client{
	// 	Timeout:   10 * time.Second,
	// 	Transport: bypass.Transport,
	// }

	// // Prepare form data (equivalent to the hidden <input> field in the form).
	// formData := url.Values{}
	// formData.Set("View", "1")

	// // Create a new HTTP POST request
	// req, err := http.NewRequest("POST", epUrl, strings.NewReader(formData.Encode()))
	// req.Form = formData
	// if err != nil {
	// 	fmt.Println("Error creating request:", err)
	// 	return nil

	// }

	// // Set necessary headers
	// req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	// req.Header.Set("Referer", epUrl)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Set("Cookie", "_ga_Q2XZ3ZSWDY=GS1.1.1733129284.12.1.1733129292.0.0.0; _ga=GA1.1.262926114.1732630228; prefetchAd_8121878=true")

	// // Create a client and execute the request
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error sending request:", err)
	// 	return nil
	// }
	// defer resp.Body.Close()

	// // Print the response status
	// fmt.Println("Response Status:", resp.Status)
	// // Optionally read and display the response body
	// body, _ := io.ReadAll(resp.Body)
	// // if err == nil {
	// // 	fmt.Println("Response Body:", string(body))
	// // }

	return []types.AniVideo{}
}
