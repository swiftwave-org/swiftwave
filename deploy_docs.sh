#!/usr/bin/env sh

GENERATED_DOCS_FOLDER=graphql-docs
DOCUMENTATION_FOLDER=./gh-pages
DOCUMENTATION_BRANCH=gh-pages
CURRENT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2)

# abort on errors
set -e

# install cli
npm i @magidoc/cli

# curl and dont throw error if branch doesnt exist
curl https://github.com/swiftwave-org/swiftwave/archive/refs/heads/$CURRENT_BRANCH.zip -o swiftwave.zip || true

# If download was successful, unzip and move to docs folder
if [ -f swiftwave.zip ]; then
  unzip swiftwave.zip
  mv swiftwave-$CURRENT_BRANCH $DOCUMENTATION_FOLDER
  rm swiftwave.zip
# else just create empty docs folder
else
  mkdir $DOCUMENTATION_FOLDER
fi

# build docs
npx magidoc generate

# Delete DOCUMENTATION_FOLDER/CURRENT_BRANCH folder if it exists
if [ -d $DOCUMENTATION_FOLDER/$CURRENT_BRANCH ]; then
  rm -rf $DOCUMENTATION_FOLDER/$CURRENT_BRANCH
fi

# Move generated docs contents to DOCUMENTATION_FOLDER/CURRENT_BRANCH
mv $GENERATED_DOCS_FOLDER/* $DOCUMENTATION_FOLDER/$CURRENT_BRANCH

# Cd into DOCUMENTATION_FOLDER
cd $DOCUMENTATION_FOLDER

# TODO: force push to gh-pages branch the DOCUMENTATION_FOLDER folder