PACKAGES := $(shell go list ./...)
COMMIT = $$(git describe --tags --always)
DATE = $$(date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)
RELEASE_BUILD_LDFLAGS = -s -w $(BUILD_LDFLAGS)

.PHONY: all
all: test

.PHONY: build
build:
	go build

.PHONY: crossbuild
crossbuild:
	$(eval version = $(shell gobump show -r))
	goxz -pv=v$(version) -os=linux,darwin -arch=386,amd64 -build-ldflags="$(RELEASE_BUILD_LDFLAGS)" \
	  -d=./dist/v$(version)

.PHONY: test
test:
	go test -v -parallel=4 ./...

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
release: devel-deps
	@./misc/scripts/bump-and-chglog.sh
	@./misc/scripts/upload-artifacts.sh

.PHONY: devel-deps
devel-deps:
	@./misc/scripts/install-devel-deps.sh
