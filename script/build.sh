#!/usr/bin/env bash

function usage()
{
	bold=$(tput bold)
	normal=$(tput sgr0)
	echo "NAME"
	echo "    build.sh - Build flaki-service"
	echo "SYNOPSIS"
	echo "    ${bold}build.sh${normal} ${bold}--env${normal} environment"
}

#
# Main
#
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"

while [ "$1" != "" ];
do
	case $1 in
		--env ) shift
				ENV=$1
				;;
		* ) 	usage
				exit 1
	esac
	shift
done

if [ -z ${ENV} ]; then
	usage
	exit 1
fi

# Directories.
FLATBUF_DIR="./pkg/flaki/flatbuffer"

# Delete the old dirs.
echo "==> Removing old directories..."
rm -f bin/*
mkdir -p bin/
rm -f "$FLATBUF_DIR"/fb/*

# Flatbuffers.
echo
echo "==> Flatbuffers:"
flatc --grpc --go -o "$FLATBUF_DIR" "$FLATBUF_DIR"/flaki.fbs 
ls -hl "$FLATBUF_DIR"/fb

# Build.
echo
echo "==> Build:"

cd cmd

# Get the git commit.
GIT_COMMIT="$(git rev-parse HEAD)"

# Override the variables GitCommit and Environment in the main package.
LD_FLAGS="-X main.GitCommit=${GIT_COMMIT} -X main.Environment=${ENV}"

#export CGO_ENABLED="0"

go build -ldflags "$LD_FLAGS" -o ../bin/flakid
echo "Build commit '${GIT_COMMIT}' for '${ENV}' environment."
ls -hl ../bin/

exit 0
