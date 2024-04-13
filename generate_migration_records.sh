#!/usr/bin/env sh

# first argument is the migration name
migration_name=$1

## ask for the migration name
if [ -z "$migration_name" ]; then
    read -p "Enter the migration name: " migration_name
fi

## if the migration name is empty, exit
if [ -z "$migration_name" ]; then
    echo "Migration name cannot be empty"
    exit 1
fi

# create a new migration file

atlas migrate diff --env gorm $migration_name