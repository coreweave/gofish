#!/bin/bash
set -euxo pipefail

# Update from upstream
git checkout main
git fetch upstream
git rebase upstream/main

# Create a temporary branch
PATCH_BRANCH=apply-patch-$(date +%Y%m%d.%H.%M)
git checkout -b "${PATCH_BRANCH}"

# Reset and reapply the patch as a commit
git reset --hard upstream/main
git cherry-pick PATCH_COMMIT_SHA || (git status && exit 1)

# Merge back into main
git checkout main
git merge "${PATCH_BRANCH}"
git push origin main
