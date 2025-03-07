#!/bin/bash
set -euxo pipefail

PATCH_BRANCH="patch-storage"
PATCH_FILE="location_indicator.diff"

# Ensure the patch-storage branch exists and fetch the latest patch
git fetch origin "${PATCH_BRANCH}" || (echo "Patch storage branch not found! Exiting." && exit 1)
git checkout "${PATCH_BRANCH}"
git pull origin "${PATCH_BRANCH}"

# Ensure the patch file exists in patch-storage
if [[ ! -f "${PATCH_FILE}" ]]; then
    echo "Patch file not found in ${PATCH_BRANCH}! Exiting."
    exit 1
fi

# Store patch in a temporary location
cp "${PATCH_FILE}" /tmp/

# Switch to main and update from upstream
git checkout main
git fetch upstream
git rebase upstream/main

# Create a unique branch for applying the patch
APPLY_BRANCH="apply-patch-$(date +%Y%m%d.%H%M%S)"
git checkout -b "${APPLY_BRANCH}"

# Apply the patch from /tmp without copying it into the branch
git apply --reject --whitespace=fix /tmp/"${PATCH_FILE}" || (git status && exit 1)
git commit -am "Reapply patch: LocationIndicatorActive any"

# Merge back into main
git checkout main
git merge --no-ff "${APPLY_BRANCH}"
git push origin main
