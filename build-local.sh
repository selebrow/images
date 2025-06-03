docker buildx build --platform linux/amd64 -t base/ubuntu-noble ./base/ubuntu-noble

docker run --rm --platform linux/amd64 -v `(pwd)`:/work \
	ubuntu:24.04 \
	/bin/bash -c 'cd work && apt-get update && \
	 apt-get install -y wget gpg && \
	 mkdir -m 700 ~/.gnupg && \
	 echo LATEST_FIREFOX_VERSION=$(./scripts/download-firefox.sh) >> .env && \
	 echo LATEST_CHROME_VERSION=$(./scripts/download-chrome.sh) >> .env'

docker run -d --name registry -p 5000:5000 registry:latest

sleep 5

export $(cat .env | xargs)

export REGISTRY=localhost:5000
export INSECURE_REGISTRY=true

go run scripts/updater.go update-firefox
go run scripts/updater.go update-chrome

chrome_version=$(echo $LATEST_CHROME_VERSION | cut -d . -f 1)
firefox_version=$(echo $LATEST_FIREFOX_VERSION | cut -d . -f 1)

mkdir playwright/browser_data
mkdir webdriver/browser_data

IMAGE_TYPE=webdriver go run scripts/updater.go download-firefox $firefox_version
IMAGE_TYPE=webdriver go run scripts/updater.go download-chrome $chrome_version
IMAGE_TYPE=playwright go run scripts/updater.go download-chrome $chrome_version

export $(go run scripts/updater.go update-meta)

make -C webdriver/src
make -C webdriver/chrome/src

for browser in firefox chrome
do
    docker buildx build -t selebrow/webdriver-${browser} --file ./webdriver/${browser}.Dockerfile \
        ./webdriver/ \
        --network=host \
        --platform linux/amd64 \
        --build-arg BASE_IMAGE=base/ubuntu-noble \
        --build-arg BASE_IMAGE_TAG=latest
done

playwright_version="1.51.1"
if [ ! -z "${LATEST_PLAYWRIGHT_VERSION}" ]; then
    playwright_version=$LATEST_PLAYWRIGHT_VERSION
fi

for browser in firefox chrome webkit
do
    docker buildx build -t selebrow/playwright-${browser} --file ./playwright/${browser}.Dockerfile \
        ./playwright/ \
        --network=host \
        --platform linux/amd64 \
        --build-arg BASE_IMAGE=base/ubuntu-noble \
        --build-arg BASE_IMAGE_TAG=latest \
        --build-arg PLAYWRIGHT_VERSION=$LATEST_PLAYWRIGHT_VERSION
done

docker rm -f registry
