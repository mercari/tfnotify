#!/bin/bash

set -e

current_version="$(gobump show -r)"

echo "current version: $current_version"
while true
do
    read -p "Specify [major | minor | patch]: " semver
    case "$semver" in
        major | minor | patch )
            gobump "$semver" -w
            next_version="$(gobump show -r)"
            break
            ;;
        *)
            echo "Invalid semver type" >&2
            continue
            ;;
    esac
    shift
done

git commit -am "Bump version $next_version"
git tag "v$next_version"

git-chglog -o CHANGELOG.md
git commit -am "Update changelog"

git push && git push --tags
