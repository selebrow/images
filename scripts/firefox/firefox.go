package firefox

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"updater/github"
	"updater/registry"
)

const (
	Image = "firefox-src"

	GeckodriverFile = "geckodriver-linux64.tar.gz"
	FirefoxFile     = "firefox.deb"
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

func DownloadGecko() error {
	release, err := github.GetRelease("https://api.github.com/repos/mozilla/geckodriver/releases/latest")

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, "-linux64.tar.gz") {
			downloadURL = asset.BrowserDownloadURL
		}
	}

	log.Println(downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	geckoFile, err := os.Create(GeckodriverFile)
	if err != nil {
		return err
	}
	defer geckoFile.Close()

	_, err = io.Copy(geckoFile, resp.Body)
	return err
}

func UploadRevision(rev string, reg *registry.Registry) error {
	return reg.UploadFiles(rev, GeckodriverFile, FirefoxFile)
}
