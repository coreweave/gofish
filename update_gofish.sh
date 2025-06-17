#!/bin/bash

#
# Updates internal gofish by creating a new branch $PR_BRANCH from internal/main then rebasing $PR_BRANCH onto upstream/main
#

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 PR_BRANCH_NAME"
    exit 1
fi

set -euxo pipefail

PATCH_BRANCH="patch-storage"
PATCH_FILE="location_indicator.diff"
PR_BRANCH_NAME="$1"


# Setup the remotes properly
CW_INTERNAL_REMOTE="git@github.com:coreweave/gofish.git"
UPSTREAM_REMOTE="git@github.com:stmcginnis/gofish.git"
git remote show internal || git remote add internal "${CW_INTERNAL_REMOTE}"
git remote show upstream || git remote add upstream "${UPSTREAM_REMOTE}"


# Ensure the patch-storage branch exists and fetch the latest patch (from internal remote)
git fetch internal "${PATCH_BRANCH}" || (echo "Patch storage branch not found! Exiting." && exit 1)
git checkout "${PATCH_BRANCH}"
git pull internal "${PATCH_BRANCH}"

# Ensure the patch file exists in patch-storage
if [[ ! -f "${PATCH_FILE}" ]]; then
    echo "Patch file not found in ${PATCH_BRANCH}! Exiting."
    exit 1
fi

# Store the patch in a temporary location and convert it to mailbox format
cp "${PATCH_FILE}" /tmp/location_indicator.patch
git format-patch --stdout HEAD~1 > /tmp/location_indicator.patch

# Switch to internal/main and update from upstream/main
git checkout main
# create a new branch off internal/main
git checkout -b "${PR_BRANCH_NAME}"
git fetch upstream
# rebase $PR_BRANCH (aka internal/main) onto upstream/main
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

# Merge our patches from $APPLY_BRANCH into $PR_BRANCH then push to internal
git checkout "${PR_BRANCH_NAME}"
git merge --no-ff "${APPLY_BRANCH}"
git push internal "${PR_BRANCH_NAME}"
