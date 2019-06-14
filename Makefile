NAME=openvpn_exporter
SOURCES=$(find . -name "*.go" -not -path "./vendor/*")

GO ?= go
VERSION ?= $(shell git describe --tags --always --dirty)
BUILDER ?= $(shell echo "`git config user.name` <`git config user.email`>")
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
GOVERSION=$(shell $(GO) version)
BUILDTIME ?= $(shell date)

LDFLAGS=-X 'github.com/prometheus/common/version.Version=$(VERSION)' \
				-X 'github.com/prometheus/common/version.Revision=$(VERSION)'\
				-X 'github.com/prometheus/common/version.Branch=$(BRANCH)'\
				-X 'github.com/prometheus/common/version.BuildDate=$(BUILDTIME)'\
				-X 'github.com/prometheus/common/version.BuildUser=$(BUILDER)'\
				-w -extldflags "-static"

.DEFAULT: help

.PHONY: help
help:
	@grep -E '^[a-z0-9A-Z_\/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## clean generated binaries
	rm -rf $(NAME)

$(NAME): $(SOURCES) ## generate the binary
	env CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ $^

.PHONY: run
run: $(NAME) ## run the binary locally
	./$(NAME) --log.level="info"

.PHONY: test
test: ## run tests
	$(GO) test -v ./...

PACKAGES=$(shell $(GO) list ./...)

.PHONY: coverage
coverage: ## run tests with coverage
	@echo "mode: set" > cover.out
	@echo Running coverage for $(PACKAGES)
	@for package in $(PACKAGES); do \
		go test -v -coverprofile=profile.out $${package}; \
		cat profile.out | grep -v "mode: set" >> cover.out; \
	done
	@-go tool cover -html=cover.out -o cover.html
