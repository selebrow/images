package meta

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type MatrixEntry struct {
	BrowserName string `json:"browser_name"`
	ImageTag    string `json:"image_tag"`
	BrowserTag  string `json:"browser_tag"`
	BaseTag     string `json:"base_image_tag"`
	Platform    string `json:"platform"`
	ImageType   string `json:"-"`
}

func (m *Meta) GeneratePWMatrix(oldMeta *Meta) (string, error) {
	matrix := m.GenerateMatrix(oldMeta, Playwright)
	data, err := json.Marshal(matrix)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (m *Meta) GenerateWDMatrix(oldMeta *Meta) (string, error) {
	matrix := m.GenerateMatrix(oldMeta, WebDriver)
	data, err := json.Marshal(matrix)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (m *Meta) GenerateMatrix(oldMeta *Meta, imageType string) []MatrixEntry {
	oldImages := oldMeta.Build[imageType].Images
	newImages := m.Build[imageType].Images

	var matrix []MatrixEntry

	baseTag := os.Getenv("BASE_IMAGE_TAG")

	gitRev := getRev()

	for browserName, browser := range newImages {
		for imageTag, browserTag := range browser.Tags {
			// if the base image was rebuilt the baseTag will not be "latest"
			// and we need to update all browser images
			if _, ok := oldImages[browserName]; ok && baseTag == "latest" {
				updated := checkBrowserImageUpdated(imageType, browserName, gitRev)
				if _, ok := oldImages[browserName].Tags[imageTag]; ok && !updated {
					continue
				}
			}

			matrix = append(matrix, MatrixEntry{
				BrowserName: browserName,
				ImageTag:    imageTag,
				BrowserTag:  browserTag,
				BaseTag:     baseTag,
				Platform:    browser.Platform,
				ImageType:   imageType,
			})
		}
	}

	return matrix
}

func getRev() string {
	if os.Getenv("GITHUB_REF") != "refs/heads/main" {
		return "origin/main"
	}

	rev, err := exec.Command("git", "rev-parse", "HEAD^").Output()
	if err != nil {
		log.Fatalf("failed to get rev: %v", err)
	}

	return strings.TrimSpace(string(rev))
}

func checkBrowserImageUpdated(platform, browser, rev string) bool {
	file := fmt.Sprintf("%s/%s.Dockerfile", platform, browser)
	return exec.Command("git", "diff", rev, "--exit-code", file).Run() != nil
}
