package extractors

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Generate a random string of alphanumeric characters
func createHashTable() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 10)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range result {
		result[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(result)
}

// Extract base URL from the full URL
func getBaseUrl(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

// Extract video URL from the given URL
func GetUrlFromDownstream(videoPageURL string) (string, error) {
	// Perform HTTP GET request
	resp, err := http.Get(videoPageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseText := string(body)

	// Get the base URL
	host := getBaseUrl(resp.Request.URL.String())

	// Find the md5 value
	md5Regex := regexp.MustCompile(`/pass_md5/[^']*`)
	md5Match := md5Regex.FindString(responseText)
	if md5Match == "" {
		return "", fmt.Errorf("md5 value not found")
	}
	md5 := host + md5Match

	// Perform another GET request for the md5 value
	resp2, err := http.Get(md5)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	// Read the second response body
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", err
	}
	trueUrl := string(
		body2,
	) + createHashTable() + "?token=" + strings.Split(md5, "/")[len(strings.Split(md5, "/"))-1]

	// Extract quality from title
	qualityRegex := regexp.MustCompile(`\d{3,4}p`)
	qualityMatch := qualityRegex.FindString(responseText)
	if qualityMatch == "" {
		qualityMatch = "unknown"
	}

	// Return the final video URL
	return trueUrl, nil
}
