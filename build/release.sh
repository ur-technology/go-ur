#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# VERSIONS="gur-linux-386 gur-linux-amd64 gur-linux-arm64 gur-darwin-386 gur-darwin-amd64 gur-windows-386 gur-windows-amd64"
# rm build/bin/*
# make $VERSIONS

rm -rf build/release
mkdir build/release
cp COPYING build/release

FILES=build/bin/gur-*
for f in build/bin/gur-*
do
  echo "Processing $f file..."
  cp $f build/release/gur
  cd build/release

  filename=$(basename "$f")
  zip $filename.zip gur COPYING
  cd -
done

rm build/release/COPYING build/release/gur
