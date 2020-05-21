#!/bin/zsh

# Script to do gospiga server release. This script would output the built
# binaries in $TMP.  This script should NOT be responsible for doing any
# testing, or uploading to any server.  The sole task of this script is to
# build the binaries and prepare them such that any human or script can then
# pick these up and use them as they deem fit.

# Output colors
RED='\033[91;1m'
RESET='\033[0m'

# Don't use standard GOPATH. Create a new one.
unset GOBIN
GOPATH="/tmp/go"
if [ -d $GOPATH ]; then
   chmod -R 755 $GOPATH
fi
rm -Rf $GOPATH
mkdir $GOPATH

# Necessary to pick up Gobin binaries like protoc-gen-gofast
PATH="$GOPATH/bin:$PATH"

# The Go version used for release builds must match this version.
GOVERSION="1.14.3"

# Turn off go modules by default. Only enable go modules when needed.
export GO111MODULE=off

TAG=$1
# The Docker tag should not contain a slash e.g. feature/issue1234
# The initial slash is taken from the repository name dgraph/dgraph:tag
# DTAG=$(echo "$TAG" | tr '/' '-')

# DO NOT change the /tmp/build directory, because Dockerfile also picks up binaries from there.
TMP="/tmp/build"
rm -Rf $TMP
mkdir $TMP

if [ -z "$TAG" ]; then
  echo "Must specify which tag to build for."
  exit 1
fi
echo "Building gospiga server for tag: $TAG"

# Stop on first failure.
set -e
set -o xtrace

# Check for existence of strip tool.
type aarch64-linux-gnu-strip
type shasum

echo "Using Go version"
go version
if [[ ! "$(go version)" =~ $GOVERSION ]]; then
   echo -e "${RED}Go version is NOT expected. Should be $GOVERSION.${RESET}"
   exit 1
fi

go get -u src.techknowlogick.com/xgo

mkdir -p $GOPATH/src/github.com/kind84/gospiga
basedir=$GOPATH/src/github.com/kind84/gospiga
cp -r . $basedir

# pushd $basedir/dgraph
#   git pull
#   git checkout $TAG
#   # HEAD here points to whatever is checked out.
#   lastCommitSHA1=$(git rev-parse --short HEAD)
#   gitBranch=$(git rev-parse --abbrev-ref HEAD)
#   lastCommitTime=$(git log -1 --format=%ci)
#   release_version=$(git describe --always --tags)
# popd

# Build Linux.
pushd $basedir
  xgo -go="go-$GOVERSION" --targets=linux/arm64 -out server ./server/cmd/server
  # xgo -go="go-$GOVERSION" --targets=linux/arm64 -ldflags \
  #     "-X $release=$release_version -X $branch=$gitBranch -X $commitSHA1=$lastCommitSHA1 -X '$commitTime=$lastCommitTime'" .
  aarch64-linux-gnu-strip -x server-linux-arm64
  mkdir $TMP/linux
  mv server-linux-arm64 $TMP/linux/server
popd

# Create Docker image.
cp ./server/dev.Dockerfile $TMP
pushd $TMP
  docker build -t gospiga/server-dev:$TAG -f dev.Dockerfile .
popd
rm $TMP/dev.Dockerfile

rm -Rf $TMP/linux

echo "Release $TAG is ready."
