#!/bin/bash

NEW_BRANCH="update-images-${RUN_ID}"
COMMIT_MESSAGE="Auto-update images"
DESCRIPTION="Chrome ${LATEST_CHROME_VERSION:-???}, Firefox ${LATEST_FIREFOX_VERSION:-???}, Playwright ${LATEST_PLAYWRIGHT_VERSION:-???}"

gh auth login --with-token
gh auth setup-git

git checkout -b "${NEW_BRANCH}"
git add meta.json
git commit -m "$COMMIT_MESSAGE" -m "$DESCRIPTION"
git push -u origin $(git rev-parse --abbrev-ref HEAD)

gh pr create --base main --head "${NEW_BRANCH}" --fill
