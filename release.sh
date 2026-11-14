#!/bin/bash
set -e

# Builds the release archives into dist/ and the release notes into notes.md. A version tag such
# as v1.0.0 makes a release of that version, and anything else makes the "development" pre-release.
# The release details go to $GITHUB_OUTPUT when it is set.

ref_type=${GITHUB_REF_TYPE:-branch}
ref_name=${GITHUB_REF_NAME:-$(git rev-parse --abbrev-ref HEAD)}
sha=${GITHUB_SHA:-$(git rev-parse HEAD)}

rm -rf build/release dist
mkdir -p build/release dist

# The Mullvad app is only available for these architectures on Linux.
for arch in amd64 arm64; do
  # No version in the name, so the download link of the latest release stays the same.
  name="mullvad-web-controller-linux-$arch"
  dir="build/release/$name"
  mkdir -p "$dir"
  CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -ldflags='-s -w' -o "$dir/" .
  cp mullvad-web-controller.service service.sh README.md LICENSE.txt "$dir/"
  tar -czf "dist/$name.tar.gz" -C build/release "$name"
done
(cd dist && sha256sum *.tar.gz > SHA256SUMS)

if [[ "$ref_type" == tag ]]; then
  tag="$ref_name"
  title="$tag"
  development=false
  # Release notes: the commit subjects since the previous release tag.
  previous=$(git describe --tags --abbrev=0 --match 'v*' HEAD^ 2>/dev/null || true)
  git log --no-merges --format='- %s' ${previous:+$previous..}HEAD > notes.md
else
  tag=development
  title="Development build"
  development=true
  echo "Development build of ${sha:0:7}." > notes.md
fi

if [[ -n "$GITHUB_OUTPUT" ]]; then
  { echo "tag=$tag"; echo "title=$title"; echo "development=$development"; } >> "$GITHUB_OUTPUT"
fi
