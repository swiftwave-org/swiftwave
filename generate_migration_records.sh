#!/usr/bin/env sh

# join of all arguments from first to last using _ as separator
migration_name=$(echo "$@" | tr ' ' '_')

# ask for the migration name
if [ -z "$migration_name" ]; then
    read -p "Enter the migration name: " migration_name
fi

# if the migration name is empty, exit
if [ -z "$migration_name" ]; then
    echo "Migration name cannot be empty"
    exit 1
fi

echo "Migration name: $migration_name"

# anything except a-z0-9_ is not allowed
if echo "$migration_name" | grep -q '[^a-z0-9_]'; then
    echo "Migration name can only contain lowercase alphabets, numbers, and underscores"
    exit 1
fi

# check if atlas is installed
if ! command -v atlas > /dev/null; then
    echo "atlas is not installed"
    echo "Run \`curl -sSf https://atlasgo.sh | sh\` to install atlas or visit https://atlasgo.io/getting-started/ to learn more"
    exit 1
fi

# create a new migration file
atlas migrate diff --env gorm $migration_name