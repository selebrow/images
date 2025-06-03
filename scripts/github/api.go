package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type (
	Release struct {
		TagName string  `json:"tag_name"`
		Assets  []Asset `json:"assets"`
	}

	// needed for geckodriver
	Asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}
)

func GetReleases(url string) ([]Release, error) {
	var target []Release
	err := getReleases(url, &target)

	return target, err
}

func GetRelease(url string) (Release, error) {
	var target Release
	err := getReleases(url, &target)

	return target, err
}

func getReleases(url string, target any) error {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}

	token := os.Getenv("GH_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GH_TOKEN")))
		req.Header.Set("Accept", "application/vnd.github+json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}

	return nil
}
