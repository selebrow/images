package meta

import (
	"encoding/json"
	"os"
)

type MatrixEntry struct {
	BrowserName string `json:"browser_name"`
	ImageTag    string `json:"image_tag"`
	BrowserTag  string `json:"browser_tag"`
	BaseTag     string `json:"base_image_tag"`
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

func (m *Meta) GenerateMatrix(oldMeta *Meta, browser string) []MatrixEntry {
	oldImages := oldMeta.Build[browser].Images
	newImages := m.Build[browser].Images

	var matrix []MatrixEntry

	baseTag := os.Getenv("BASE_IMAGE_TAG")

	for browserName, tags := range newImages {
		for imageTag, browserTag := range tags.Tags {
			// if the base image was rebuilt we need to update all browser images
			if _, ok := oldImages[browserName]; ok && baseTag == "latest" {
				if _, ok := oldImages[browserName].Tags[imageTag]; ok {
					continue
				}
			}

			matrix = append(matrix, MatrixEntry{
				BrowserName: browserName,
				ImageTag:    imageTag,
				BrowserTag:  browserTag.Version,
				BaseTag:     baseTag,
				ImageType:   browser,
			})
		}
	}

	return matrix
}
