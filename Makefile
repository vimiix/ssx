VERSION := $(shell head -1 version)
GO      := GO111MODULE=on CGO_ENABLED=0 go
_COMMIT := $(shell git describe --no-match --always --dirty)
COMMIT  := $(if $(COMMIT),$(COMMIT),$(_COMMIT))
BUILDDATE  := $(shell date '+%Y-%m-%dT%H:%M:%S')
REPO    := github.com/vimiix/ssx
LDFLAGS := -X "$(REPO)/ssx/version.Version=$(VERSION)"
LDFLAGS += -X "$(REPO)/ssx/version.Revision=$(COMMIT)"
LDFLAGS += -X "$(REPO)/ssx/version.BuildDate=$(BUILDDATE)"
LDFLAGS += $(EXTRA_LDFLAGS)
FILES   := $$(find . -name "*.go")
TEST_FILES   := $$(go list ./...)

.PHONY: help
help: ## print help info
	@printf "%-30s %s\n" "Target" "Description"
	@printf "%-30s %s\n" "------" "-----------"
	@grep -E '^[ a-zA-Z1-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## run go mod tidy
	@echo "go mod tidy"
	$(GO) mod tidy

.PHONY: fmt
fmt: ## format source code
	@echo "gofmt (simplify)"
	@gofmt -s -l -w $(FILES) 2>&1
	@echo "goimports (if installed)"
	$(shell goimports -w $(FILES) 2>/dev/null)

.PHONY: lint
lint: tidy fmt ## lint code with golangci-lint
	$([[ command -v golangci-lint ]] || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1)
	@golangci-lint run -v

.PHONY: test
test: ## run all unit tests
	$(GO) test -gcflags=all=-l $(TEST_FILES) -coverprofile dist/cov.out -covermode count

.PHONY: ssx
ssx: ## build ssx binary
	$(GO) build -ldflags '$(LDFLAGS)' -gcflags '-N -l' -o dist/ssx ./cmd/ssx/main.go

.PHONY: tag
tag: ## make tag with version.txt
	git tag -a "$(VERSION)" -m "Release version $(VERSION)"

.PHONY: snapshot
snapshot: ## build ssx release snapshot
	goreleaser release --clean --snapshot