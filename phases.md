# Performance Pass Phases

## Scope

Optimize the highest-cost developer and CI feedback loops first, using repository-native entrypoints and measured evidence.

This repo is currently classified as a **Go library repository** with:
- root module: `go.mod`
- CI entrypoints: `.github/workflows/go.yml`, `.github/workflows/bump.yml`
- lint entrypoint: `.golangci.yml`
- additional nested modules under `ext/*` and `examples/middleware`

## Source-of-truth entrypoints

- Lint: `golangci-lint run`
- Build: `go build -v ./...`
- Test: `go test -v ./...`
- Vulnerability scan: `govulncheck ./...`
- Dependency bump verification: `.github/workflows/bump.yml`

---

## Phase 1 — Repository discovery and baseline definition

### Status
Completed

### Findings
- The primary validated workflow lives in `.github/workflows/go.yml`.
- The bump flow lives in `.github/workflows/bump.yml`.
- The repo is not using `just`, `make`, Docker, or Nix as the primary build entrypoint.
- Root module currently expands to **32 packages**.

### Baseline measurement set

Measured commands:

- `golangci-lint run` → **47.837s**
- `go build -v ./...` → **39.678s** warm / **63.626s** cold
- `go test -v ./...` → **43.419s** warm / **44.068s** cold
- `govulncheck ./...` → **16.143s** warm / **10.084s** isolated

Simulated old CI critical lane:

- `go build -v ./... && go test -v ./... && govulncheck ./...` → **96.591s** cold

Simulated old dependency bump verify step:

- `go build ./... && go test ./...` → **82.822s** cold

---

## Phase 2 — Highest-cost feedback loop optimization

### Status
Completed

### Bottlenecks identified
1. The old CI workflow serialized `build`, `test`, and `govulncheck` in one lane.
2. The bump workflow re-ran a full `go build` before `go test`, adding cost with limited extra signal.
3. The prior CI matrix referenced `386`, while local evidence showed the path was already broken.

### Changes applied

#### `.github/workflows/go.yml`
- kept `lint` as its own job
- split the old combined lane into separate jobs:
  - `build`
  - `test`
  - `vulncheck`
- added Go cache restore for `build`, `test`, and `vulncheck`
- removed stale matrix/`GOARCH` coupling from lint/build flow
- fixed shell output grouping for `actionlint` compliance

#### `.github/workflows/bump.yml`
- removed redundant `go build ./...`
- kept `go test ./...` as the verification step

### Validation
- `actionlint .github/workflows/go.yml .github/workflows/bump.yml` → passed

### Post-change measurement set

- `golangci-lint run` → **47.837s**
- `go build -v ./...` → **63.626s**
- `go test -v ./...` → **44.068s**
- `govulncheck ./...` → **10.084s**
- simulated new bump verify: `go test ./...` → **70.522s**

### Delta

#### CI critical path
- Before: **96.591s**
- After: **63.626s**
- Absolute delta: **-32.965s**
- Percent delta: **-34.13%**

#### Bump verify step
- Before: **82.822s**
- After: **70.522s**
- Absolute delta: **-12.300s**
- Percent delta: **-14.85%**

---

## Phase 3 — Clear existing validation blockers

### Status
Completed

### Blockers found
- `golangci-lint run` initially failed on:
  - `dispatcher/handlers/commands.go:90` (`gosec` G115, rune-to-byte conversion)
  - `adapter/context_resolve_test.go:46` (unused helper)

### Changes applied
- `dispatcher/handlers/commands.go`
  - replaced unsafe `byte(prefix)` comparison with rune decoding via `utf8.DecodeRuneInString`
- `adapter/context_resolve_test.go`
  - removed unused `addTestChat` helper

### Validation
- `go test ./dispatcher/handlers ./adapter` → passed
- `golangci-lint run ./dispatcher/handlers ./adapter` → **0 issues**
- `golangci-lint run` → **0 issues**
- `go test ./...` → passed

---

## Phase 4 — Remove broken or wasted breadth

### Status
Completed

### Evidence
Local validation of the former 32-bit path failed:

- `GOARCH=386 go build ./...` → failed in **5.336s**

Observed failure causes:
- missing 32-bit libc headers: `gnu/stubs-32.h`
- `github.com/bytedance/sonic` does not support 32-bit arch

### Outcome
- current workflow files no longer contain `386`, `GOARCH`, or a build matrix for this path

---

## Phase 5 — Remaining work and follow-ups

### Status
Ready

### Still open
1. **Toolchain vuln noise remains**
   - `govulncheck -show verbose ./...` reports standard library vulnerabilities from **Go 1.26.1**
   - fixes are reported in **Go 1.26.2**
   - this is currently a toolchain/environment issue, not a code-level regression from this pass

2. **Hosted-runner validation still pending**
   - local timings simulate the CI job shape, but do not include GitHub scheduling overhead

### Recommended next sequence
1. Keep the current CI workflow split and cache changes.
2. Upgrade CI/local Go to a patched version that clears current stdlib vulnerability findings.
3. Use the nested-module measurements in Phase 6 before expanding CI scope.
4. If requested, prepare a commit after final review.

---

## Phase 6 — Nested module discovery and measurement

### Status
Completed

### CI inclusion check
- Current workflow files do **not** validate nested modules directly.
- `.github/workflows/go.yml` only runs root-module `go build -v ./...` and `go test -v ./...`.
- `.github/workflows/bump.yml` only runs root-module `go test ./...`.

### Nested module package counts
- `ext/gorm` → **12 packages**
- `ext/pgx` → **1 package**
- `ext/redis` → **1 package**
- `ext/mongodb` → **1 package**
- `examples/middleware` → **1 package**

### Cold-cache measurements

#### Build
- `ext/gorm`: `go build ./...` → **51.350s**
- `ext/pgx`: `go build ./...` → **29.313s**
- `ext/redis`: `go build ./...` → **29.717s**
- `ext/mongodb`: `go build ./...` → **28.627s**
- `examples/middleware`: `go build ./...` → **36.529s**

#### Test
- `ext/gorm`: `go test ./...` → **45.964s**
- `ext/pgx`: `go test ./...` → **35.325s**
- `ext/redis`: `go test ./...` → **35.576s**
- `ext/mongodb`: `go test ./...` → **34.600s**
- `examples/middleware`: `go test ./...` → **33.034s**

### Findings
1. `ext/gorm` is the only nested module with substantial breadth.
2. Most nested-module test time is compile/setup cost; several modules have **no test files**.
3. `ext/gorm` test breadth is dominated by example packages that compile but do not execute tests.
4. Because these modules are **not currently part of active CI**, expanding validation would increase CI cost rather than reduce it.

### Decision
- No workflow changes were applied for nested modules in this pass.
- The current evidence supports keeping CI scope focused on the root module unless nested modules are explicitly added to the supported CI surface.

### Deferred low-risk follow-up
- If nested modules need CI coverage later, start with `ext/gorm` first, because it is the only nested module large enough to materially affect validation time.
- If that happens, prefer a dedicated workflow or targeted package selection over folding all nested modules into the root critical path.

---

## Phase 7 — `ext/gorm` validation strategy analysis

### Status
Completed

### Question
If `ext/gorm` were added to CI, is there a materially cheaper narrow validation command than validating the full module?

### Evidence

Cold-cache timings:

- `cd ext/gorm && go build .` → **45.222s**
- `cd ext/gorm && go test .` → **45.992s**
- `cd ext/gorm && go test ./...` → **46.145s**
- `cd ext/gorm && go test ./examples/...` → **35.474s**

Additional structure evidence:
- `ext/gorm` has **no `_test.go` files**.
- The module contains **11 example packages** plus the root package.
- `go test .` is nearly the same cost as `go test ./...`, so narrowing to the root package does **not** produce a meaningful time reduction.

### Findings
1. `ext/gorm` validation cost is dominated by compile/setup work rather than executing tests.
2. The example subtree is non-trivial, but excluding it would only save part of the time while also reducing coverage of the public example surface.
3. There is no evidence-backed low-risk command change that meaningfully improves `ext/gorm` validation time in isolation.

### Decision
- No `ext/gorm`-specific CI change was applied in this pass.
- If `ext/gorm` is promoted into active CI later, the safest default is still `go test ./...` within `ext/gorm` unless coverage requirements are explicitly narrowed.

---

## Files changed in this pass

- `.github/workflows/go.yml`
- `.github/workflows/bump.yml`
- `dispatcher/handlers/commands.go`
- `adapter/context_resolve_test.go`

## Validation summary

- `actionlint .github/workflows/go.yml .github/workflows/bump.yml` → passed
- `golangci-lint run` → passed
- `go test ./...` → passed

## Start state

This phased plan has already been started and phases 1–6 are complete.
The current state is ready for review, optional commit, or a separate decision on nested-module CI expansion.
