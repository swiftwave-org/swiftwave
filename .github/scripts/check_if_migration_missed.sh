#!/usr/bin/env sh

# get output of atlas migrate diff --env gorm
output=$(atlas migrate diff --env gorm --dev-url postgres://postgres:pass@localhost:5432/dev?sslmode=disable)

# check if output contains `no changes to be made`
if echo "$output" | grep -q 'no changes to be made'; then
    echo "No migration changes detected"
    exit 0
else
    echo "Migration changes detected and need to be committed"
    echo "Please run ./generate_migration_records.sh <name> to generate new migration files"
    exit 1
fi