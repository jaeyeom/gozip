# Makefile for gozip

BINDIR      := bin
GOZIP       := $(BINDIR)/gozip
GOUNZIP     := $(BINDIR)/gounzip
COVERAGE    := coverage.out

# All source files (for staleness checks)
GO_FILES    := $(shell find . -name '*.go' -not -path './.omc/*')

.PHONY: all check format check-format lint fix test build coverage coverage-html coverage-report clean

# ── Full local workflow ────────────────────────────────────────────────

all: format fix test build

# ── CI-friendly checks (no mutation) ──────────────────────────────────

check: check-format lint test build

# ── Format / check-format ─────────────────────────────────────────────

format:
	@gofmt -w .

check-format:
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "gofmt: files need formatting"; exit 1; }

# ── Lint / fix ────────────────────────────────────────────────────────

lint:
	@go vet ./...

fix: format

# ── Test ──────────────────────────────────────────────────────────────

test:
	@go test ./...

# ── Build ─────────────────────────────────────────────────────────────

build: $(GOZIP) $(GOUNZIP)

$(GOZIP): $(GO_FILES)
	@go build -o $@ ./cmd/gozip

$(GOUNZIP): $(GO_FILES)
	@go build -o $@ ./cmd/gounzip

# ── Coverage ──────────────────────────────────────────────────────────

coverage:
	@go test -coverprofile=$(COVERAGE) ./...

coverage-html: coverage
	@go tool cover -html=$(COVERAGE) -o coverage.html

coverage-report: coverage
	@go tool cover -func=$(COVERAGE)

# ── Clean ─────────────────────────────────────────────────────────────

clean:
	@rm -rf $(BINDIR) $(COVERAGE) coverage.html
