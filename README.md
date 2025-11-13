# Selebrow Images

This repository contains build scripts for the browser images used in Selebrow.

Images are built using a scheduled pipeline that checks for browser and Playwright updates every week and generates a build matrix accordingly.

### Project structure

`base/` - Ubuntu image with preinstalled dependencies used as a base layer for all browser images.

`playwright/` - Playwright images and scripts.

`webdriver` - WebDriver images and scripts.

`scripts/` - Contains a Go CLI tool and helper scripts. The CLI tool manages updates and build matrix generation from the meta file.

`meta.json` - Metadata file that contains image build information, such as versions, platforms, and tags. On release the [browser image manifest](https://github.com/selebrow/selebrow.github.io/blob/main/browsers.yaml) is generated from this file.
