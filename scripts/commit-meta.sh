#!/bin/bash

NEW_BRANCH="update-images-${RUN_ID}"
COMMIT_MESSAGE="Auto-update images"
DESCRIPTION="Chrome ${LATEST_CHROME_VERSION:-???}, Firefox ${LATEST_FIREFOX_VERSION:-???}, Playwright ${LATEST_PLAYWRIGHT_VERSION:-???}"

git config user.name "selebrow-ci"
git config user.email "213042858+selebrow-ci[bot]@users.noreply.github.com"

git checkout -b "${NEW_BRANCH}"
git add meta.json
git commit -m "$COMMIT_MESSAGE" -m "$DESCRIPTION" >> /dev/null
git push -u origin $(git rev-parse --abbrev-ref HEAD) >> /dev/null

gh auth login --with-token
gh pr create --base main --head "${NEW_BRANCH}" --fill
