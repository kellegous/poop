GOLANGCI_LINT_VERSION := v2.15.0
GOIMPORTS_VERSION := v0.51.0

bin/golangci-lint: tools.mk
	GOBIN="$(CURDIR)/bin" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

bin/goimports: tools.mk
	GOBIN="$(CURDIR)/bin" go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
