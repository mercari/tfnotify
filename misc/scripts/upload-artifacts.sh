#!/bin/bash

set -e

version="v$(gobump show -r)"
make crossbuild
ghr -username mercari -replace "$version" "dist/$version"
