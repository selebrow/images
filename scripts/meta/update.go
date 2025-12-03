package meta

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"updater/github"
)

const (
	browserChrome   = "chrome"
	browserChromium = "chromium"
	browserFirefox  = "firefox"
	browserWebkit   = "webkit"

	wdKeepTags = 4
	pwKeepTags = 8

	Playwright = "playwright"
	WebDriver  = "webdriver"

	MetaFilePath = "meta.json"
)

type (
	Meta struct {
		Build      map[string]*Image `json:"build"`
		pwBrowsers map[string]playwrightBrowsers
	}

	Image struct {
		Images map[string]*Browser `json:"images"`
	}

	Browser struct {
		Platform string            `json:"platform"`
		Tags     map[string]string `json:"tags"`
	}

	playwrightBrowsers struct {
		Browsers []playwrightBrowser `json:"browsers"`
	}

	playwrightBrowser struct {
		Name    string `json:"name"`
		Version string `json:"browserVersion"`
	}
)

func Load(filepath string) (*Meta, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m Meta
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Meta) Update() error {
	if err := m.updateWD(); err != nil {
		return err
	}

	if err := m.updatePW(); err != nil {
		return err
	}

	return m.save()
}

func (m *Meta) save() error {
	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(MetaFilePath, data, os.ModePerm)
}

func (m *Meta) updateWD() error {
	chromeTag, err := parseVersionTag("LATEST_CHROME_VERSION")
	if err != nil {
		return err
	}

	if chromeTag != "" {
		m.updateTags(WebDriver, browserChrome, chromeTag, "", wdKeepTags)
	}

	firefoxTag, err := parseVersionTag("LATEST_FIREFOX_VERSION")
	if err != nil {
		return err
	}

	if firefoxTag != "" {
		m.updateTags(WebDriver, browserFirefox, firefoxTag, "", wdKeepTags)
	}

	return nil
}

func (m *Meta) updatePW() error {
	chromeTags, err := m.getAvailableChromeTags()
	if err != nil {
		return err
	}

	releaseTags, err := getLatestPlaywrightTags()
	if err != nil {
		return err
	}

	pwImages := m.Build[Playwright].Images

	var changed bool

	for _, tag := range releaseTags {
		playwrightTag := normalizePWTag(tag)

		// drop it if it's not valid for chrome - we cannot have version skew for playwright
		if _, ok := pwImages[browserChrome].Tags[playwrightTag]; !ok {
			chromeTag, err := m.getPWBrowserTag(browserChrome, tag)
			if err != nil {
				return err
			}

			if _, ok := chromeTags[chromeTag]; !ok {
				log.Println("Chrome tag", chromeTag, "is still not available")
				continue
			}
		} else {
			log.Println("Playwright tag", playwrightTag, "is already defined")

			// skip all remaining tags since we start from the latest one
			break
		}

		chromeTag, err := m.getPWBrowserTag(browserChrome, tag)
		if err != nil {
			return err
		}
		changed = changed || m.updateTags(Playwright, browserChrome, playwrightTag, chromeTag, pwKeepTags)

		chromiumTag, err := m.getPWBrowserTag(browserChromium, tag)
		if err != nil {
			return err
		}
		changed = changed || m.updateTags(Playwright, browserChromium, playwrightTag, chromiumTag, pwKeepTags)

		firefoxTag, err := m.getPWBrowserTag(browserFirefox, tag)
		if err != nil {
			return err
		}
		changed = changed || m.updateTags(Playwright, browserFirefox, playwrightTag, firefoxTag, pwKeepTags)

		webkitTag, err := m.getPWBrowserTag(browserWebkit, tag)
		if err != nil {
			return err
		}
		changed = changed || m.updateTags(Playwright, browserWebkit, playwrightTag, webkitTag, pwKeepTags)

		if changed {
			fmt.Printf("LATEST_PLAYWRIGHT_VERSION=%s\n", tag)
			break
		}
	}

	return nil
}

func (m *Meta) getPWBrowserTag(browser, tag string) (string, error) {
	version, err := m.getPWBrowserVersion(browser, tag)
	if err != nil {
		return "", err
	}

	if browser != browserChrome {
		return version, nil
	}

	v, err := strconv.Atoi(strings.Split(version, ".")[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.0", v-1), nil
}

func (m *Meta) getPWBrowserVersion(browser, tag string) (string, error) {
	if m.pwBrowsers == nil {
		m.pwBrowsers = make(map[string]playwrightBrowsers)
	}

	if _, ok := m.pwBrowsers[tag]; !ok {
		resp, err := http.Get(fmt.Sprintf(
			"https://raw.githubusercontent.com/microsoft/playwright/%s/packages/playwright-core/browsers.json",
			tag,
		))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var browsers playwrightBrowsers
		if err := json.NewDecoder(resp.Body).Decode(&browsers); err != nil {
			return "", err
		}

		m.pwBrowsers[tag] = browsers
	}

	filter := browser
	if browser == browserChrome {
		filter = "chromium"
	}

	for _, b := range m.pwBrowsers[tag].Browsers {
		if b.Name == filter {
			return b.Version, nil
		}
	}

	// should not be reachable
	return "", nil
}

func (m *Meta) getAvailableChromeTags() (map[string]struct{}, error) {
	tags := make(map[string]struct{})

	for tag := range m.Build[WebDriver].Images[browserChrome].Tags {
		tags[tag] = struct{}{}
	}

	return tags, nil
}

func (m *Meta) updateTags(image, browser, tag, version string, keepTags int) bool {
	log.Printf("Checking %s tag %s for %s\n", image, tag, browser)

	browserImage := m.Build[image]

	if browserImage.Images[browser].Tags == nil {
		browserImage.Images[browser] = &Browser{Tags: make(map[string]string)}
	}

	if _, ok := browserImage.Images[browser].Tags[tag]; ok {
		log.Println("Tag", tag, "is already defined for", browser)
		return false
	}

	if version == "" {
		version = tag
	}

	if browser != "webkit" {
		version = strings.Split(version, ".")[0]
	}

	browserImage.Images[browser].Tags[tag] = version

	log.Println("Tag", tag, "added for", browser)

	if len(browserImage.Images[browser].Tags) > keepTags {
		delTag := m.findOldestTag(image, browser)
		delete(browserImage.Images[browser].Tags, delTag)

		log.Println("Tag", delTag, "deleted for", browser)
	}

	m.Build[image] = browserImage

	return true
}

func (m *Meta) findOldestTag(image, browser string) string {
	var oldest string

	for tag := range m.Build[image].Images[browser].Tags {
		if oldest == "" {
			oldest = tag
			continue
		}

		if tag < oldest {
			oldest = tag
		}
	}

	return oldest
}

func normalizePWTag(tag string) string {
	return strings.TrimPrefix(tag, "v")
}

func getLatestPlaywrightTags() ([]string, error) {
	releases, err := github.GetReleases("https://api.github.com/repos/microsoft/playwright/releases?per_page=5")
	if err != nil {
		return nil, err
	}

	var tags []string
	for _, rel := range releases {
		tags = append(tags, rel.TagName)
	}

	return tags, nil
}

func parseVersionTag(envVar string) (string, error) {
	ver := os.Getenv(envVar)
	if ver == "" {
		log.Println(envVar, "is not defined, skipping")
		return "", nil
	}

	parsedVersion := regexp.MustCompile(`^(\d+\.\d+)`).FindString(ver)
	if parsedVersion == "" {
		return "", fmt.Errorf("malformed env var: %s", envVar)
	}

	return parsedVersion, nil
}
