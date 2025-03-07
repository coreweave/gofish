#!/bin/bash
set -euxo pipefail

PATCH_BRANCH="patch-storage"
PATCH_FILE="location_indicator.diff"

# Ensure the patch-storage branch exists and fetch the latest patch
git fetch origin "${PATCH_BRANCH}" || (echo "Patch storage branch not found! Exiting." && exit 1)
git checkout "${PATCH_BRANCH}"
git pull origin "${PATCH_BRANCH}"

# Ensure the patch file exists
if [[ ! -f "${PATCH_FILE}" ]]; then
    echo "Patch file not found in ${PATCH_BRANCH}! Exiting."
    exit 1
fi

# Store the patch in a temporary location and convert it to mailbox format
cp "${PATCH_FILE}" /tmp/location_indicator.patch
git format-patch --stdout HEAD~1 > /tmp/location_indicator.patch

# Switch to main and update from upstream
git checkout main
git fetch upstream
git rebase upstream/main

# Create a unique branch for applying the patch
APPLY_BRANCH="apply-patch-$(date +%Y%m%d.%H%M%S)"
git checkout -b "${APPLY_BRANCH}"

# Apply the patch using `git am` for smarter merging
if [[ -f /tmp/location_indicator.patch ]]; then
    git am --3way /tmp/location_indicator.patch || (git status && exit 1)
else
    echo "Patch file not found! Exiting..."
    exit 1
fi

# Merge back into main
git checkout main
git merge --no-ff "${APPLY_BRANCH}"
git push origin main
