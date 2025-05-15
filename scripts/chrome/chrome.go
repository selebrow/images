package chrome

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"updater/registry"
)

const (
	Image = "chrome-src"

	ChromedriverFile   = "chromedriver-linux64.zip"
	ChromeFile         = "google-chrome-stable.deb"
	ChromeHeadlessFile = "chrome-headless-shell-linux64.zip"

	chromedriver        = "chromedriver"
	chromeHeadlessShell = "chrome-headless-shell"
)

var (
	downloads chromeDownloads
)

type (
	chromeVersions struct {
		Versions []chromeVersion `json:"versions"`
	}

	chromeVersion struct {
		Version   string          `json:"version"`
		Downloads chromeDownloads `json:"downloads"`
	}

	chromeDownloads struct {
		Chromedriver        []platformURL `json:"chromedriver"`
		ChromeHeadlessShell []platformURL `json:"chrome-headless-shell"`
	}

	platformURL struct {
		Platform string `json:"platform"`
		URL      string `json:"url"`
	}
)

func GetRevision(rev string, reg *registry.Registry) (string, error) {
	revision := strings.Split(rev, ".")[0]

	exists, err := reg.CheckImageExists(context.Background(), revision)
	if err != nil {
		return "", err
	}

	if !exists {
		return revision, nil
	}

	return "", nil
}

func DownloadDriver(major string) error {
	return downloadBinary(chromedriver, major, ChromedriverFile)
}

func DownloadHeadlessShell(major string) error {
	return downloadBinary(chromeHeadlessShell, major, ChromeHeadlessFile)
}

func UploadRevision(rev string, reg *registry.Registry) error {
	return reg.UploadFiles(rev, ChromeFile, ChromedriverFile, ChromeHeadlessFile)
}

func downloadBinary(name, major, filename string) error {
	if err := getVersions(major); err != nil {
		return err
	}

	major = fmt.Sprintf("%s.", major)

	var (
		platformURLs []platformURL
		downloadURL  string
	)

	switch name {
	case chromedriver:
		platformURLs = downloads.Chromedriver
	case chromeHeadlessShell:
		platformURLs = downloads.ChromeHeadlessShell
	}

	for _, pURL := range platformURLs {
		if pURL.Platform == "linux64" {
			downloadURL = pURL.URL
		}
	}

	log.Println(downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	binaryFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer binaryFile.Close()

	_, err = io.Copy(binaryFile, resp.Body)
	return err
}

func getVersions(major string) error {
	if len(downloads.Chromedriver) != 0 {
		return nil
	}

	resp, err := http.Get("https://googlechromelabs.github.io/chrome-for-testing/known-good-versions-with-downloads.json")
	if err != nil {
		return err
	}

	var versions chromeVersions
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return err
	}

	major = fmt.Sprintf("%s.", major)

	for _, version := range versions.Versions {
		if strings.HasPrefix(version.Version, major) {
			downloads = version.Downloads
		}
	}

	return nil
}
