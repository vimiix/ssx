SEMVER := $(shell head -1 version)
VERSION_MAJOR := $(shell echo $(SEMVER)|awk -F '.' '{ print $$1 }')
VERSION_MINOR := $(shell echo $(SEMVER)|awk -F '.' '{ print $$2 }')
VERSION_PATCH := $(shell echo $(SEMVER)|awk -F '.' '{ print $$3 }')

GO      := GO111MODULE=on CGO_ENABLED=0 GOPROXY="https://goproxy.cn,direct" go
_COMMIT := $(shell git describe --no-match --always --dirty)
COMMIT  := $(if $(COMMIT),$(COMMIT),$(_COMMIT))
BUILDDATE  := $(shell date '+%Y-%m-%dT%H:%M:%S')
REPO    := github.com/vimiix/ssx
LDFLAGS := -X "$(REPO)/internal/version.Major=$(VERSION_MAJOR)"
LDFLAGS += -X "$(REPO)/internal/version.Minor=$(VERSION_MINOR)"
LDFLAGS += -X "$(REPO)/internal/version.Patch=$(VERSION_PATCH)"
LDFLAGS += -X "$(REPO)/internal/version.Revision=$(COMMIT)"
LDFLAGS += -X "$(REPO)/internal/version.BuildDate=$(BUILDDATE)"
LDFLAGS += $(EXTRA_LDFLAGS)
FILES   := $$(find . -name "*.go")

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
lint: ## lint code with golangci-lint
	$([[ command -v golangci-lint ]] || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1)
	@golangci-lint run -v

.PHONY: ssx
ssx: tidy fmt lint ## build ssx binary
	$(GO) build -ldflags '$(LDFLAGS)' -gcflags '-N -l' -o dist/ssx ./cmd/ssx/main.go