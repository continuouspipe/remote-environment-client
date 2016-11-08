#!/bin/bash

FILE_NAME=cp-remote
MANIFEST_BASE_URL=https://continuouspipe.github.io/remote-environment-client
set -e


if [ $# -ne 1 ]; then
  echo "Usage: `basename $0` <tag>"
  exit 65
fi

TAG=$1

#
# Build master branch
#
git checkout master
git tag ${TAG}

# Add the build details
cat $FILE_NAME | awk 'NR==2 {print "\# Version: '$TAG'"} 1' > $FILE_NAME.build

#
# Copy executable file into GH pages
#
git checkout gh-pages

cp $FILE_NAME.build downloads/$FILE_NAME-$TAG
git add downloads/$FILE_NAME-$TAG

SHA1=$(openssl sha1 downloads/$FILE_NAME-$TAG | awk '{print $2}')

JSON='name:"'$FILE_NAME'"'
JSON="${JSON},sha1:\"${SHA1}\""
JSON="${JSON},url:\"${MANIFEST_BASE_URL}/downloads/${FILE_NAME}-${TAG}\""
JSON="${JSON},version:\"${TAG}\""

#
# Update manifest
#
cat manifest.json | jsawk -a "this.push({${JSON}})" | python -mjson.tool > manifest.json.tmp
mv manifest.json.tmp manifest.json
git add manifest.json

# Symlink latest version
rm downloads/$FILE_NAME-latest
cp downloads/$FILE_NAME-${TAG} downloads/$FILE_NAME-latest
git add downloads/$FILE_NAME-latest

git commit -m "Release version ${TAG}"

#
# Go back to master
#
git checkout master
rm $FILE_NAME.build

echo "New version created. Now you should run:"
echo "git push origin gh-pages"
echo "git push ${TAG}"

