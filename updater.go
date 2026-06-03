package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const AppVersion = "1.0.4"

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

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func TriggerSelfUpdate(latestTag string) error {
	// 1. Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Go up 3 levels to find the .app bundle (e.g. /Applications/Stay Awake.app)
	appPath := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))
	if filepath.Ext(appPath) != ".app" {
		return fmt.Errorf("current executable is not inside a macOS .app bundle: %s", appPath)
	}

	// 2. Create a temporary folder
	tempDir, err := os.MkdirTemp("", "stay-awake-update")
	if err != nil {
		return err
	}

	zipPath := filepath.Join(tempDir, "Stay.Awake.zip")
	downloadURL := fmt.Sprintf("https://github.com/princev89/stay-awake/releases/download/%s/Stay.Awake.zip", latestTag)

	// 3. Download zip
	if err := DownloadFile(zipPath, downloadURL); err != nil {
		return err
	}

	// 4. Unzip
	if err := Unzip(zipPath, tempDir); err != nil {
		return err
	}

	newAppPath := filepath.Join(tempDir, "Stay Awake.app")
	if _, err := os.Stat(newAppPath); os.IsNotExist(err) {
		return fmt.Errorf("extracted zip did not contain Stay Awake.app")
	}

	// 5. Create the background shell script to replace the app
	scriptPath := filepath.Join(tempDir, "update.sh")
	scriptContent := fmt.Sprintf(`#!/bin/bash
sleep 1
rm -rf "%s"
mv "%s" "%s"
xattr -cr "%s" 2>/dev/null || true
open "%s"
`, appPath, newAppPath, appPath, appPath, appPath)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return err
	}

	// 6. Run the script in the background and exit
	cmd := exec.Command("bash", scriptPath)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Exit the current app so the script can overwrite it
	os.Exit(0)
	return nil
}
