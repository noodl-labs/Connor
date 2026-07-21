# ConnorLLM — solo-founder DevOps Makefile
# Usage: make help
# Disable colors: make check NO_COLOR=1

.PHONY: help install build test test-race lint fmt tidy verify check ci smoke smoke-cli release-dry

RUNTIME_DIR := services/runtime
BINARY      := bin/connor
VERSION     ?= v0.1.0
LDFLAGS     := -X github.com/noodl-labs/ConnorLLM/services/runtime/internal/cli.Version=$(VERSION)

# ── Colors ───────────────────────────────────────────────────────────
ifeq ($(NO_COLOR),1)
  GREEN  :=
  RED    :=
  YELLOW :=
  BLUE   :=
  CYAN   :=
  DIM    :=
  RESET  :=
  BOLD   :=
else
  GREEN  := \033[0;32m
  RED    := \033[0;31m
  YELLOW := \033[0;33m
  BLUE   := \033[0;34m
  CYAN   := \033[0;36m
  DIM    := \033[2m
  RESET  := \033[0m
  BOLD   := \033[1m
endif

# ── Help ─────────────────────────────────────────────────────────────
help:
	@echo "$(BOLD)ConnorLLM$(RESET) — dev commands"
	@echo ""
	@echo "  $(GREEN)make check$(RESET)        → before commit (lint + tests + build)"
	@echo "  $(GREEN)make ci$(RESET)           → same gates as GitHub Actions"
	@echo "  $(BLUE)make test$(RESET)         → fast unit tests"
	@echo "  $(BLUE)make lint$(RESET)         → golangci-lint"
	@echo "  $(BLUE)make build$(RESET)        → compile bin/connor"
	@echo "  $(BLUE)make install$(RESET)      → install connor to \$$PATH"
	@echo "  $(YELLOW)make smoke$(RESET)        → live suite (requires CONNOR_* env)"
	@echo "  $(DIM)make fmt$(RESET)            → gofmt"
	@echo "  $(DIM)make tidy$(RESET)           → go mod tidy"
	@echo "  $(CYAN)make release-dry$(RESET)  → pre-release checks + tag instructions"
	@echo ""
	@echo "  $(DIM)NO_COLOR=1 make check$(RESET)  → disable colors"

# ── Daily dev ────────────────────────────────────────────────────────
check: lint test-race build
	@echo "$(GREEN)✓ check OK — safe to commit$(RESET)"

ci: verify test-race build smoke-cli
	@echo "$(GREEN)✓ ci OK — matches GitHub Actions$(RESET)"

# ── Go ───────────────────────────────────────────────────────────────
verify:
	@echo "$(YELLOW)→ verify modules$(RESET)"
	@cd $(RUNTIME_DIR) && go mod verify

test:
	@echo "$(YELLOW)→ test$(RESET)"
	@cd $(RUNTIME_DIR) && go test ./...

test-race:
	@echo "$(YELLOW)→ test (race)$(RESET)"
	@cd $(RUNTIME_DIR) && go test -race -count=1 ./...

lint:
	@echo "$(YELLOW)→ lint$(RESET)"
	@cd $(RUNTIME_DIR) && golangci-lint run ./...

fmt:
	@echo "$(YELLOW)→ fmt$(RESET)"
	@cd $(RUNTIME_DIR) && gofmt -w .

tidy:
	@echo "$(YELLOW)→ tidy$(RESET)"
	@cd $(RUNTIME_DIR) && go mod tidy

# ── Build ────────────────────────────────────────────────────────────
build:
	@echo "$(YELLOW)→ build$(RESET)"
	@cd $(RUNTIME_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../$(BINARY) ./cmd/connor

install:
	@echo "$(YELLOW)→ install$(RESET)"
	@cd $(RUNTIME_DIR) && go install -ldflags "$(LDFLAGS)" ./cmd/connor
	@echo "$(GREEN)✓ connor installed$(RESET)"

# ── Smoke (no network — same as CI) ──────────────────────────────────
smoke-cli: build
	@echo "$(YELLOW)→ smoke-cli$(RESET)"
	@./$(BINARY) --help
	@./$(BINARY) run --help

# ── Live smoke (optional — your machine, with .env) ───────────────────
smoke:
	@test -n "$$CONNOR_BASE_URL" || (echo "$(RED)export CONNOR_BASE_URL and CONNOR_API_KEY first$(RESET)" && exit 1)
	@echo "$(YELLOW)→ smoke (live)$(RESET)"
	@connor run benchmarks/examples/serving-smoke.yaml

# ── Release ──────────────────────────────────────────────────────────
release-dry: ci
	@echo ""
	@echo "$(CYAN)Release dry-run$(RESET)"
	@echo "  $(BLUE)Version:$(RESET) $(VERSION)"
	@echo "  $(DIM)Tag:     git tag -a $(VERSION) -m '$(VERSION)'$(RESET)"
	@echo "  $(DIM)Push:    git push origin $(VERSION)$(RESET)"
	@echo "$(GREEN)✓ ready to tag$(RESET)"