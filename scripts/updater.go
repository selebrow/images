package main

import (
	"fmt"
	"log"
	"maps"
	"os"
	"slices"

	"updater/chrome"
	"updater/firefox"
	"updater/meta"
	"updater/registry"
)

var (
	containerregistry = "ghcr.io/selebrow"
	insecureRegistry  = false
	project           = "images"
)

func main() {
	if reg := os.Getenv("REGISTRY"); reg != "" {
		containerregistry = reg
	}
	if insecure := os.Getenv("INSECURE_REGISTRY"); insecure != "" {
		insecureRegistry = true
	}
	if proj := os.Getenv("PROJECT"); proj != "" {
		project = proj
	}

	command := os.Args[1]

	switch command {
	case "update-firefox":
		updateFirefox()
	case "update-chrome":
		updateChrome()
	case "update-meta":
		updateMeta()
	case "generate-wd-matrix":
		generateWDMatrix()
	case "generate-pw-matrix":
		generatePWMatrix()
	case "firefox-links":
		getFirefoxLinks()
	case "chrome-links":
		getChromeLinks()
	case "release-images":
		releaseImages()
	default:
		log.Fatalf("unknown command %q\n", command)
	}
}

func updateFirefox() {
	reg, err := registry.InitRegistry(insecureRegistry, containerregistry, project, firefox.Image)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	currentVersion := os.Getenv("LATEST_FIREFOX_VERSION")
	if currentVersion == "" {
		log.Println("current version not found, exiting")
		return
	}

	log.Println("current revision of Firefox is set to", currentVersion)
	log.Println("getting Firefox revision...")

	rev, err := firefox.GetRevision(currentVersion, reg)
	if err != nil {
		log.Fatalf("failed to get Firefox revision: %v", err)
	}

	if rev == "" {
		log.Println("Firefox version is up to date, exiting")
		return
	}

	log.Println("downloading Gecko driver")
	if err := firefox.DownloadGecko(); err != nil {
		log.Fatalf("failed to download gecko: %v", err)
	}

	log.Println("uploading new Firefox revision")
	if err := firefox.UploadRevision(rev, reg); err != nil {
		log.Fatalf("failed to upload Firefox revision: %v", err)
	}
}

func updateChrome() {
	reg, err := registry.InitRegistry(insecureRegistry, containerregistry, project, chrome.Image)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	currentVersion := os.Getenv("LATEST_CHROME_VERSION")
	if currentVersion == "" {
		log.Println("current version not found, exiting")
		return
	}

	log.Println("current revision of Chrome is set to", currentVersion)
	log.Println("getting Chrome revision...")

	rev, err := chrome.GetRevision(currentVersion, reg)
	if err != nil {
		log.Fatalf("failed to get Chrome revision: %v", err)
	}

	if rev == "" {
		log.Println("Chrome version is up to date, exiting")
		return
	}

	log.Println("downloading Chrome driver")
	if err := chrome.DownloadDriver(rev); err != nil {
		log.Fatalf("failed to download Chrome driver: %v", err)
	}

	log.Println("downloading Chrome headless shell")
	if err := chrome.DownloadHeadlessShell(rev); err != nil {
		log.Fatalf("failed to download Chrome headless shell: %v", err)
	}

	log.Println("uploading new Chrome revision")
	if err := chrome.UploadRevision(rev, reg); err != nil {
		log.Fatalf("failed to upload Chrome revision: %v", err)
	}
}

func updateMeta() {
	m, err := meta.Load(meta.MetaFilePath)
	if err != nil {
		log.Fatalf("failed to load meta: %v", err)
	}

	if err := m.Update(); err != nil {
		log.Fatalf("failed to update meta: %v", err)
	}
}

func generateWDMatrix() {
	generateMatrix(meta.WebDriver)
}

func generatePWMatrix() {
	generateMatrix(meta.Playwright)
}

func generateMatrix(image string) {
	oldMeta, err := meta.Load("meta-old.json")
	if err != nil {
		log.Fatalf("failed to load old meta: %v", err)
	}

	m, err := meta.Load(meta.MetaFilePath)
	if err != nil {
		log.Fatalf("failed to load meta: %v", err)
	}

	var (
		matrix string
	)

	switch image {
	case meta.Playwright:
		matrix, err = m.GeneratePWMatrix(oldMeta)
	case meta.WebDriver:
		matrix, err = m.GenerateWDMatrix(oldMeta)
	}

	if err != nil {
		log.Fatalf("failed to generate %s build matrix: %v", image, err)
	}

	fmt.Println(matrix)
}

func getChromeLinks() {
	files := map[string]string{
		chrome.ChromeFile:         "CHROME_URL",
		chrome.ChromeHeadlessFile: "CHROME_HEADLESS_URL",
		chrome.ChromedriverFile:   "CHROMEDRIVER_URL",
	}

	getBrowserLinks(chrome.Image, files)
}

func getFirefoxLinks() {
	files := map[string]string{
		firefox.GeckodriverFile: "GECKODRIVER_URL",
		firefox.FirefoxFile:     "FIREFOX_URL",
	}

	getBrowserLinks(firefox.Image, files)
}

func getBrowserLinks(image string, files map[string]string) {
	revision := os.Args[2]
	if revision == "" {
		log.Fatalf("browser revision not provided")
	}

	reg, err := registry.InitRegistry(insecureRegistry, containerregistry, project, image)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	result, err := reg.GetDownloadLinks(revision, files)
	if err != nil {
		log.Fatalf("failed to get download links: %v", err)
	}

	for key, value := range result {
		fmt.Printf("%s=%s\n", key, value)
	}
}

func releaseImages() {
	srcTag := os.Getenv("SOURCE_TAG")
	if srcTag == "" {
		log.Fatal("empty source tag")
	}

	tag := os.Getenv("TAG_NAME")
	if tag == "" {
		log.Fatal("empty git tag")
	}

	oldMeta, err := meta.Load("meta-old.json")
	if err != nil {
		log.Fatalf("failed to load old meta: %v", err)
	}

	m, err := meta.Load(meta.MetaFilePath)
	if err != nil {
		log.Fatalf("failed to load meta: %v", err)
	}

	pwMatrix := m.GenerateMatrix(oldMeta, meta.Playwright)
	wdMatrix := m.GenerateMatrix(oldMeta, meta.WebDriver)

	for _, entry := range slices.Concat(pwMatrix, wdMatrix) {
		reg, err := registry.InitRegistry(insecureRegistry, containerregistry, project, fmt.Sprintf("%s/%s", entry.ImageType, entry.BrowserName))
		if err != nil {
			log.Fatalf("failed to init registry: %v", err)
		}

		if err := reg.TagImage(srcTag, fmt.Sprintf("%s-%s", tag, entry.ImageTag)); err != nil {
			log.Fatalf("failed to tag image: %v", err)
		}
	}

	var baseImageName string
	for key := range maps.Keys(m.Build["base"].Images) {
		baseImageName = key
		break
	}

	baseReg, err := registry.InitRegistry(insecureRegistry, containerregistry, project, fmt.Sprintf("base/%s", baseImageName))
	if err != nil {
		log.Fatalf("failed to init base registry: %v", err)
	}

	if err := baseReg.TagImage(srcTag, "latest"); err != nil {
		log.Fatalf("failed to tag base image: %v", err)
	}
}
