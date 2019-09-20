PACKAGES := $(shell go list ./...)
COMMIT = $$(git describe --tags --always)
DATE = $$(date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: all
all: test

.PHONY: build
build:
	go build

.PHONY: test
test:
	go test -v -parallel=4 ./...

.PHONY: devel-deps
devel-deps:
	@go get -v -u github.com/Songmu/ghch/cmd/ghch
	@go get -v -u github.com/Songmu/goxz/cmd/goxz
	@go get -v -u github.com/git-chglog/git-chglog/cmd/git-chglog
	@go get -v -u github.com/golang/dep/cmd/dep
	@go get -v -u golang.org/x/lint/golint
	@go get -v -u github.com/haya14busa/goverage
	@go get -v -u github.com/haya14busa/reviewdog/cmd/reviewdog
	@go get -v -u github.com/motemen/gobump/cmd/gobump
	@go get -v -u github.com/tcnksm/ghr

.PHONY: dep
dep: devel-deps
	dep ensure -v

.PHONY: reviewdog
reviewdog: devel-deps
	reviewdog -reporter="github-pr-review"

.PHONY: coverage
coverage: devel-deps
	goverage -v -covermode=atomic -coverprofile=coverage.txt $(PACKAGES)

.PHONY: release
release:
	@bash ./misc/scripts/release.sh
