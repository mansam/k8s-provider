GOPATH ?= $(HOME)/go
GOBIN ?= $(GOPATH)/bin
GOIMPORTS = $(GOBIN)/goimports

SERVE = -o bin/serve github.com/konveyor-ecosystem/k8s-provider/cmd/server
CLI = -o bin/cli github.com/konveyor-ecosystem/k8s-provider/cmd/cli

PKG = ./cmd/... \
      ./k8s/... \
      ./provider/...

PKGDIR = $(subst /...,,$(PKG))

RULESET_ARGS ?=

cmd: serve cli

serve: fmt vet
	go build $(SERVE)

cli: fmt vet
	go build $(CLI)

fmt: $(GOIMPORTS)
	$(GOIMPORTS) -w $(PKGDIR)

vet:
	go vet $(PKG)

# Ensure goimports installed.
$(GOIMPORTS):
	go install golang.org/x/tools/cmd/goimports@latest