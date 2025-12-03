package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"slices"

	"updater/browsers"
	"updater/chrome"
	"updater/firefox"
	"updater/meta"
	"updater/registry"
)

var (
	containerregistry = "ghcr.io"
	insecureRegistry  = false
	project           = "selebrow/images"
	browsersFile      = "browsers.yaml"
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
	case "download-firefox":
		downloadFirefox()
	case "download-chrome":
		downloadChrome()
	case "release-images":
		releaseImages()
	case "generate-browsers":
		generateBrowsers()
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

	var matrix string

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

func downloadChrome() {
	files := map[string]struct{}{
		chrome.ChromeFile:         {},
		chrome.ChromeHeadlessFile: {},
		chrome.ChromedriverFile:   {},
	}

	downloadBrowsers(chrome.Image, files)
}

func downloadFirefox() {
	files := map[string]struct{}{
		firefox.GeckodriverFile: {},
		firefox.FirefoxFile:     {},
	}

	downloadBrowsers(firefox.Image, files)
}

func downloadBrowsers(image string, files map[string]struct{}) {
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

	var authToken string
	if containerregistry == "ghcr.io" {
		authToken, err = registry.GenerateGHCRToken(project, image)
		if err != nil {
			log.Fatalf("failed to generate ghcr token: %v", err)
		}
	}

	for file, link := range result {
		log.Println("downloading", file)
		if err := downloadFile(file, link, authToken); err != nil {
			log.Fatalf("failed to download file: %s, %v", file, err)
		}
	}
}

func downloadFile(file, link, authToken string) error {
	req, err := http.NewRequest(http.MethodGet, link, http.NoBody)
	if err != nil {
		return err
	}

	if authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	f, err := os.Create(path.Join(os.Getenv("IMAGE_TYPE"), "browser_data", file))
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
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

		imageTag := fmt.Sprintf("%s-%s", srcTag, entry.ImageTag)
		if err := reg.TagImage(imageTag, fmt.Sprintf("%s-%s", tag, entry.ImageTag)); err != nil {
			log.Fatalf("failed to tag image: %v", err)
		}
	}

	if os.Getenv("BASE_IMAGE_TAG") == "latest" {
		log.Println("base image was not updated, skipping")
		return
	}

	var baseImageName string
	for key := range m.Build["base"].Images {
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

func generateBrowsers() {
	m, err := meta.Load(meta.MetaFilePath)
	if err != nil {
		log.Fatalf("failed load meta: %v", err)
	}

	oldMeta, err := meta.Load("meta-old.json")
	if err != nil {
		log.Fatalf("failed to load old meta: %v", err)
	}

	data, err := browsers.Generate(m, oldMeta)
	if err != nil {
		log.Fatalf("failed to generate browsers: %v", err)
	}

	f, err := os.Create(browsersFile)
	if err != nil {
		log.Fatalf("failed to create browsers file: %v", err)
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		log.Fatalf("failed to write browsers file: %v", err)
	}
}
