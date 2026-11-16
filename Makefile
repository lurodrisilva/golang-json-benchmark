# golang-json-encoding-benchmark — task runner. Wraps go/k6; see docs/runbook.md.
SHELL       := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help
MAKEFLAGS   += --no-builtin-rules

BIN      ?= bin
RESULTS  ?= results
MATRIX   ?= loadtest/matrix/pilot.json
RUN      ?= $(shell ls -1d $(RESULTS)/*/ 2>/dev/null | tail -1)
COUNT    ?= 10
BENCHCTL := $(BIN)/benchctl
REPORT   := $(BIN)/report

.PHONY: help
help: ## List targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  %-16s %s\n", $$1, $$2}'

.PHONY: tools
tools: ## Build benchctl, report, fixturegen
	@mkdir -p $(BIN)
	cd tools && go build -trimpath -o ../$(BIN)/benchctl ./cmd/benchctl
	cd tools && go build -trimpath -o ../$(BIN)/report ./cmd/report
	cd tools && go build -trimpath -o ../$(BIN)/fixturegen ./cmd/fixturegen

.PHONY: build
build: tools ## Build all server arms (+minimal variants) with build.json manifest
	$(BENCHCTL) build -out $(BIN)
	$(BENCHCTL) build -out $(BIN) -tags minimal

.PHONY: fixtures
fixtures: tools ## Regenerate fixtures (deterministic; must produce no git diff)
	$(BIN)/fixturegen -out fixtures

.PHONY: fixtures-verify
fixtures-verify: ## Verify fixture checksums
	cd fixtures && shasum -a 256 -c SHA256SUMS

.PHONY: vet
vet: ## gofmt + go vet for both modules
	@test -z "$$(gofmt -l service tools)" || { gofmt -l service tools; exit 1; }
	cd service && go vet ./... && go vet -tags minimal ./...
	cd tools && go vet ./...

.PHONY: test
test: fixtures-verify ## Unit + integration + conformance tests (both GOEXPERIMENT settings)
	cd service && go test ./...
	cd service && GOEXPERIMENT=nojsonv2 go test ./...
	cd service && go test -tags minimal ./...
	cd tools && go test ./...

.PHONY: race
race: ## Race detector (never during performance runs)
	cd service && go test -race -count=1 ./...
	cd service && GOEXPERIMENT=nojsonv2 go test -race -count=1 ./...

.PHONY: conformance
conformance: ## Write conformance matrices (raw codec vs v1 spec) to reports/conformance/
	@mkdir -p reports/conformance
	cd service && CONFORMANCE_OUT=$(CURDIR)/reports/conformance/conformance-default.tsv go test -count=1 -run TestCorpusRaw ./internal/conformance/
	cd service && GOEXPERIMENT=nojsonv2 CONFORMANCE_OUT=$(CURDIR)/reports/conformance/conformance-nojsonv2.tsv go test -count=1 -run TestCorpusRaw ./internal/conformance/

.PHONY: fuzz
fuzz: ## Fuzz the shared policy scanner for 60s
	cd service && go test -run '^$$' -fuzz FuzzCheck -fuzztime 60s ./internal/policy/

.PHONY: microbench
microbench: ## Decoder-only benchmarks, COUNT reps each pair, benchstat summary
	@mkdir -p reports/bench
	cd service && go test -run '^$$' -bench . -benchmem -count=$(COUNT) ./bench/ | tee $(CURDIR)/reports/bench/default.txt
	cd service && GOEXPERIMENT=nojsonv2 go test -run '^$$' -bench . -benchmem -count=$(COUNT) ./bench/ | tee $(CURDIR)/reports/bench/nojsonv2.txt
	go run golang.org/x/perf/cmd/benchstat@v0.0.0-20261009192801-be2c69fb417e -col /codec reports/bench/default.txt > reports/bench/benchstat-default.txt || echo "benchstat unavailable; raw output kept"
	go run golang.org/x/perf/cmd/benchstat@v0.0.0-20261009192801-be2c69fb417e -col /codec reports/bench/nojsonv2.txt > reports/bench/benchstat-nojsonv2.txt || true

.PHONY: env
env: tools ## Print environment manifest
	$(BENCHCTL) env

.PHONY: smoke
smoke: build ## e2e smoke matrix (~11 min); not a result
	$(BENCHCTL) run -matrix loadtest/matrix/e2e.json -bin $(BIN) -results $(RESULTS)

.PHONY: plan
plan: tools ## Dry-run MATRIX: print trial order and time estimate
	$(BENCHCTL) run -matrix $(MATRIX) -bin $(BIN) -results $(RESULTS) -dry-run

.PHONY: run
run: build test ## Run MATRIX (pilot by default). Long-running; requires approval for the authoritative matrix
	$(BENCHCTL) run -matrix $(MATRIX) -bin $(BIN) -results $(RESULTS)

.PHONY: report
report: tools ## Generate report for RUN (default: latest results dir)
	@test -n "$(RUN)" || { echo "no run under $(RESULTS)"; exit 1; }
	$(REPORT) -run $(RUN) -out $(RUN)/report \
	  -bench reports/bench/benchstat-default.txt \
	  -conformance reports/conformance -pprof

.PHONY: clean
clean: ## Remove built binaries (never deletes results/)
	rm -rf $(BIN)
