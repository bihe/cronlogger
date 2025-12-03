PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y.%m.%d_%H%M`

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)


## set the default architecture should work for most Linux systems
ARCH := amd64

UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M), x86_64)
	ARCH = amd64
endif
ifeq ($(UNAME_M), arm64)
	ARCH = arm64
endif


.PHONY: all clean mod-update build test coverage compose-integration integration

all: help

# ---------------------------------------------------------------------------
# application tasks
# ---------------------------------------------------------------------------

clean: ## clean caches and build output
	@-$(MAKE) -s go-clean

mod-update: ## update to latest compatible packages (yes golang!)
	@-$(MAKE) -s go-update

build: ## compile the whole repo
	@-$(MAKE) -s go-build

test: ## unit-test the repo
	@-$(MAKE) -s go-test

coverage: ## print coverage results for the repo
	@-$(MAKE) -s go-coverage

# internal tasks

go-clean:
	@echo "  >  Cleaning build cache"
	go clean ./...
	rm -rf ./dist

go-update:
	@echo "  >  Go update dependencies ..."
	go get -u -t ./...
	go mod tidy -compat=1.25

go-build:
	@echo "  >  Building the repo ..."
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-w -s" -o ./dist/linux/amd64/cronlogger ./cmd/logger/main.go
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -ldflags="-w -s" -o ./dist/linux/arm64/cronlogger ./cmd/logger/main.go

	go tool templ generate && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux  go build -ldflags="-w -s -X main.Version=${BUILD} -X main.Build=${COMMIT}" -o ./dist/linux/amd64/cronlogger_server ./cmd/server/main.go
	go tool templ generate && CGO_ENABLED=0 GOARCH=arm64 GOOS=linux  go build -ldflags="-w -s -X main.Version=${BUILD} -X main.Build=${COMMIT}" -o ./dist/linux/arm64/cronlogger_server ./cmd/server/main.go

go-test:
	@echo "  >  Testing the repo ..."
	# tparse: https://github.com/mfridman/tparse
	# go install github.com/mfridman/tparse@latest
	go test -v -race -count=1 -json ./... | go tool tparse -all

go-coverage:
	@echo "  >  Testing the repo (coverage) ..."
	# tparse: https://github.com/mfridman/tparse
	go test -race -coverprofile="coverage.txt" -covermode atomic -count=1 -json ./... | go tool tparse -all


## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

