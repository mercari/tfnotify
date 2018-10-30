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

git-chglog -o CHANGELOG.md --next-tag "v$next_version"
git commit -am "Bump version $next_version and update changelog"
git tag "v$next_version"

git push && git push --tags
