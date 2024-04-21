#!/usr/bin/env sh


# check if atlas is installed
if ! command -v atlas > /dev/null; then
    echo "atlas is not installed"
    echo "Run \`curl -sSf https://atlasgo.sh | sh\` to install atlas or visit https://atlasgo.io/getting-started/ to learn more"
    exit 1
fi

# hash migration records
atlas migrate hash --env gorm