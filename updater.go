package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const AppVersion = "1.0.3"

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

func CheckForUpdate(currentVersion string) (bool, string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/princev89/stay-awake/releases/latest", nil)
	if err != nil {
		return false, "", "", err
	}
	
	// Add user-agent header as required by GitHub API
	req.Header.Set("User-Agent", "Stay-Awake-App-Updater")
	
	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return false, "", "", err
	}

	latestClean := rel.TagName
	if len(latestClean) > 0 && latestClean[0] == 'v' {
		latestClean = latestClean[1:]
	}
	currentClean := currentVersion
	if len(currentClean) > 0 && currentClean[0] == 'v' {
		currentClean = currentClean[1:]
	}

	if latestClean != currentClean {
		return true, rel.TagName, rel.HtmlUrl, nil
	}
	return false, "", "", nil
}
