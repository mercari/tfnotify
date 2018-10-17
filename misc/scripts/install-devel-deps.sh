#!/bin/bash

set -e

go get -v -u github.com/Songmu/ghch/cmd/ghch
go get -v -u github.com/Songmu/goxz/cmd/goxz
go get -v -u github.com/git-chglog/git-chglog/cmd/git-chglog
go get -v -u github.com/golang/dep/cmd/dep
go get -v -u golang.org/x/lint/golint
go get -v -u github.com/haya14busa/goverage
go get -v -u github.com/haya14busa/reviewdog/cmd/reviewdog
go get -v -u github.com/motemen/gobump/cmd/gobump
go get -v -u github.com/tcnksm/ghr
