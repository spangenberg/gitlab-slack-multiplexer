#!/bin/bash

set -e
cd $(dirname "$BASH_SOURCE")/..

VERSION_PKG=github.com/spangenberg/gitlab-slack-multiplexer/src/version
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X $VERSION_PKG.version=$(git describe --tags --always --abbrev=9 || echo)"
LDFLAGS="$LDFLAGS -X $VERSION_PKG.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo -ldflags \'$LDFLAGS\'
