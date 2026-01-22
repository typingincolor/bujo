#!/bin/bash
set -e

# Rename all semver releases to sequential build-N format
# Usage: GH_TOKEN=your_token ./scripts/rename-releases.sh [--dry-run]

DRY_RUN=false
if [ "$1" == "--dry-run" ]; then
  DRY_RUN=true
  echo "=== DRY RUN MODE ==="
fi

if [ -z "$GH_TOKEN" ]; then
  echo "Error: GH_TOKEN environment variable required"
  echo "Usage: GH_TOKEN=your_token ./scripts/rename-releases.sh [--dry-run]"
  exit 1
fi

export GH_TOKEN

REPO="typingincolor/bujo"

echo "Fetching all releases..."
RELEASES=$(gh release list --repo "$REPO" --limit 100 --json tagName,body,isDraft,isPrerelease)

# Sort tags by version (oldest first)
SORTED_TAGS=$(echo "$RELEASES" | jq -r '.[].tagName' | sort -V)

BUILD_NUM=1
for TAG in $SORTED_TAGS; do
  # Skip if already a build-N tag
  if [[ "$TAG" =~ ^build-[0-9]+$ ]]; then
    echo "Skipping $TAG (already renamed)"
    EXISTING_NUM=$(echo "$TAG" | sed 's/build-//')
    if [ "$EXISTING_NUM" -ge "$BUILD_NUM" ]; then
      BUILD_NUM=$((EXISTING_NUM + 1))
    fi
    continue
  fi

  NEW_TAG="build-${BUILD_NUM}"
  echo ""
  echo "=== Renaming $TAG -> $NEW_TAG ==="

  if [ "$DRY_RUN" == "true" ]; then
    echo "[DRY RUN] Would rename $TAG to $NEW_TAG"
    BUILD_NUM=$((BUILD_NUM + 1))
    continue
  fi

  # Get release details
  BODY_FILE=$(mktemp)
  echo "$RELEASES" | jq -r --arg tag "$TAG" '.[] | select(.tagName == $tag) | .body // ""' > "$BODY_FILE"
  IS_DRAFT=$(echo "$RELEASES" | jq -r --arg tag "$TAG" '.[] | select(.tagName == $tag) | .isDraft')
  IS_PRERELEASE=$(echo "$RELEASES" | jq -r --arg tag "$TAG" '.[] | select(.tagName == $tag) | .isPrerelease')

  # Get commit SHA for this tag
  COMMIT_SHA=$(gh api "repos/$REPO/git/refs/tags/$TAG" --jq '.object.sha' 2>/dev/null || echo "")

  # If it's an annotated tag, we need to dereference it
  if [ -n "$COMMIT_SHA" ]; then
    OBJ_TYPE=$(gh api "repos/$REPO/git/tags/$COMMIT_SHA" --jq '.object.type' 2>/dev/null || echo "")
    if [ "$OBJ_TYPE" == "commit" ]; then
      COMMIT_SHA=$(gh api "repos/$REPO/git/tags/$COMMIT_SHA" --jq '.object.sha' 2>/dev/null || echo "$COMMIT_SHA")
    fi
  fi

  if [ -z "$COMMIT_SHA" ]; then
    echo "Error: Could not find commit for tag $TAG, skipping"
    rm -f "$BODY_FILE"
    continue
  fi

  echo "Commit SHA: $COMMIT_SHA"

  # Download assets from old release
  ASSET_DIR=$(mktemp -d)
  echo "Downloading assets to $ASSET_DIR..."
  gh release download "$TAG" --repo "$REPO" --dir "$ASSET_DIR" 2>/dev/null || echo "No assets to download"

  # Create new tag via API
  echo "Creating tag $NEW_TAG..."
  gh api "repos/$REPO/git/refs" \
    -f ref="refs/tags/$NEW_TAG" \
    -f sha="$COMMIT_SHA" > /dev/null

  # Build release create command
  CREATE_ARGS=(--repo "$REPO" --title "$NEW_TAG" --notes-file "$BODY_FILE")
  if [ "$IS_PRERELEASE" == "true" ]; then
    CREATE_ARGS+=(--prerelease)
  fi
  if [ "$IS_DRAFT" == "true" ]; then
    CREATE_ARGS+=(--draft)
  fi

  # Add assets if any
  ASSETS=()
  if [ -n "$(ls -A "$ASSET_DIR" 2>/dev/null)" ]; then
    for asset in "$ASSET_DIR"/*; do
      ASSETS+=("$asset")
    done
  fi

  # Create new release
  echo "Creating release $NEW_TAG..."
  if [ ${#ASSETS[@]} -gt 0 ]; then
    gh release create "$NEW_TAG" "${CREATE_ARGS[@]}" "${ASSETS[@]}"
  else
    gh release create "$NEW_TAG" "${CREATE_ARGS[@]}"
  fi

  # Delete old release and tag
  echo "Deleting old release $TAG..."
  gh release delete "$TAG" --repo "$REPO" --yes
  gh api -X DELETE "repos/$REPO/git/refs/tags/$TAG"

  rm -rf "$ASSET_DIR" "$BODY_FILE"
  echo "Successfully renamed $TAG -> $NEW_TAG"

  BUILD_NUM=$((BUILD_NUM + 1))
done

echo ""
echo "=== Complete ==="
echo "Next build number will be: $BUILD_NUM"
