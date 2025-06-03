package browsers

import (
	"fmt"
	"updater/meta"

	"gopkg.in/yaml.v3"
)

const (
	WebdriverProtocol  BrowserProtocol = "webdriver"
	PlaywrightProtocol BrowserProtocol = "playwright"

	VNCPort        ContainerPort = "vnc"
	DevtoolsPort   ContainerPort = "devtools"
	FileserverPort ContainerPort = "fileserver"
	ClipboardPort  ContainerPort = "clipboard"
	BrowserPort    ContainerPort = "browser"

	imagePrefix = "ghcr.io/selebrow/images"
)

var (
	defaultLimits = map[string]string{
		"cpu":    "1",
		"memory": "2Gi",
	}

	defaultPlaywrightPorts = map[ContainerPort]int{
		BrowserPort: 4444,
		VNCPort:     5900,
	}

	defaultWebdriverPorts = map[ContainerPort]int{
		BrowserPort:    4444,
		ClipboardPort:  9090,
		DevtoolsPort:   7070,
		FileserverPort: 8080,
		VNCPort:        5900,
	}

	defaultEnv = map[string]string{}
)

type (
	BrowserProtocol string
	ContainerPort   string

	BrowserCatalog map[BrowserProtocol]Browsers
	Browsers       map[string]BrowserConfig

	BrowserConfig struct {
		Images map[string]BrowserImageConfig `yaml:"images"`
	}

	BrowserImageConfig struct {
		Image          string                `yaml:"image"`
		Cmd            []string              `yaml:"cmd"`
		DefaultVersion string                `yaml:"defaultVersion"`
		VersionTags    map[string]string     `yaml:"versionTags"`
		Ports          map[ContainerPort]int `yaml:"ports"`
		Path           string                `yaml:"path"`
		Env            map[string]string     `yaml:"env"`
		Limits         map[string]string     `yaml:"limits"`
		Labels         map[string]string     `yaml:"labels"`
		ShmSize        int64                 `yaml:"shmSize"`
		Tmpfs          []string              `yaml:"tmpfs"`
		Volumes        []string              `yaml:"volumes"`
	}
)

func Generate(m *meta.Meta) ([]byte, error) {
	browserCatalog := make(BrowserCatalog)

	for protocol, images := range m.Build {
		if protocol == "base" {
			continue
		}

		browsers := make(Browsers)
		for name, image := range images.Images {
			versionTags := make(map[string]string)
			for key, tag := range image.Tags {
				versionTags[key] = tag.ImageTag
			}

			imageConf := BrowserImageConfig{
				DefaultVersion: getLatestTag(image.Tags),
				Image:          fmt.Sprintf("%s/%s/%s", imagePrefix, protocol, name),
				Limits:         defaultLimits,
				VersionTags:    versionTags,
				Env:            defaultEnv,
			}

			switch BrowserProtocol(protocol) {
			case WebdriverProtocol:
				imageConf.Ports = defaultWebdriverPorts
			case PlaywrightProtocol:
				imageConf.Ports = defaultPlaywrightPorts
			}

			browsers[name] = BrowserConfig{
				Images: map[string]BrowserImageConfig{
					"default": imageConf,
				},
			}
		}

		browserCatalog[BrowserProtocol(protocol)] = browsers
	}

	data, err := yaml.Marshal(browserCatalog)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getLatestTag(tags map[string]meta.Tag) string {
	var tag string
	for key := range tags {
		if key > tag {
			tag = key
		}
	}

	return tag
}
