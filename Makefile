PKG := github.com/BoltzExchange/channel-bot

GO_BIN := ${GOPATH}/bin
GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

COMMIT := $(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := -ldflags "-X $(PKG)/build.Commit=$(COMMIT)"

LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint

LINT_BIN := $(GO_BIN)/golangci-lint

LINT = $(LINT_BIN) run -v

default: build

#
# Dependencies
#

$(LINT_BIN):
	@$(call print, "Fetching linter")
	go get -u $(LINT_PKG)


GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

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
