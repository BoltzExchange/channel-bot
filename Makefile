PKG := github.com/BoltzExchange/channel-bot

GO_BIN := ${GOPATH}/bin

GOTEST := GO111MODULE=on go test -v
GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v
GOLIST := go list -deps $(PKG)/... | grep '$(PKG)'| grep -v '/vendor/'

COMMIT := $(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := -ldflags "-X $(PKG)/build.Commit=$(COMMIT)"

LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint
LINT_BIN := $(GO_BIN)/golangci-lint
LINT = $(LINT_BIN) run -v

XARGS := xargs -L 1

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

default: build

#
# Dependencies
#

$(LINT_BIN):
	@$(call print, "Fetching linter")
	go get $(LINT_PKG)

#
# Tests
#

unit:
	@$(call print, "Running unit tests")
	$(GOLIST) | $(XARGS) env $(GOTEST)

#
# Building
#

build:
	@$(call print, "Building channel-bot")
	$(GOBUILD) -o channel-bot $(LDFLAGS) $(PKG)

install:
	@$(call print, "Installing channel-bot")
	$(GOINSTALL) $(LDFLAGS) $(PKG)

#
# Utils
#

fmt:
	@$(call print, "Formatting source")
	gofmt -l -s -w .

lint: $(LINT_BIN)
	@$(call print, "Linting source.")
	$(LINT)

.PHONY: build
