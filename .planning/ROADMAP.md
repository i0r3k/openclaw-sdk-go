# Roadmap: OpenClaw SDK Go

## Overview

OpenClaw SDK Go is a production-grade WebSocket client library migrated from TypeScript. This roadmap delivers a production-safe v1.0 by hardening the foundation (rate limiting, retry budgets, TLS), adding observability (metrics, graceful degradation), refactoring the oversized client struct, establishing performance benchmarks with fuzz testing, and completing release infrastructure (GoReleaser, semantic versioning, changelog automation).

## Phases

- [x] **Phase 1: Foundation Hardening** - Production-safe core with rate limiting, retry budgets, TLS CRL, pending request limits, and InsecureSkipVerify warning (completed 2026-03-28)
- [ ] **Phase 2: Observability** - Connection health metrics, per-request timeouts, event priority levels, configurable event buffer
- [ ] **Phase 3: Client Struct Refactor** - Group oversized client struct into logical sub-structs; clarify Close vs Disconnect
- [ ] **Phase 4: Benchmarking and Fuzz Testing** - Hot-path benchmarks, fuzz test depth with corpus, benchstat CI integration
- [ ] **Phase 5: Release Infrastructure** - GoReleaser library mode, semantic versioning tags, changelog automation

## Phase Details

### Phase 1: Foundation Hardening

**Goal**: Production-safe core SDK with resource limits and TLS hardening
**Depends on**: Nothing (first phase)
**Requirements**: FOUND-01, FOUND-02, FOUND-03, FOUND-04, FOUND-05
**Success Criteria** (what must be TRUE):
  1. SDK enforces client-side rate limiting -- server rejects requests only when rate limit is exceeded, not before
  2. Reconnect attempts are bounded -- after MaxRetries=10, reconnect stops and returns ErrMaxRetriesExceeded
  3. TLS CRL validation either works or is explicitly marked as stub with documentation explaining limitation
  4. Pending request map has a hard limit -- when limit is reached, SendRequest returns ErrTooManyPendingRequests immediately
  5. When InsecureSkipVerify is enabled, a warning is logged at connection time through Logger (not stderr)

**Plans**: 3 plans

Plans:
- [x] 01-foundation-01-PLAN.md -- Define shared type contracts (RequestRateLimiter interface, TokenBucketLimiter, MaxRetries field with precedence docs, typed errors integrating existing hierarchy)
- [x] 01-foundation-02-PLAN.md -- Fix channel ownership, reduce client mutex scope, implement rate limiting and configurable pending request limits
- [x] 01-foundation-03-PLAN.md -- Fix reconnect triggering, enforce MaxRetries budget, wire TLS/Logger through actual connection path (connection.go, websocket.go), CRL stub docs

### Phase 2: Observability

**Goal**: Connection health visibility and graceful degradation under load
**Depends on**: Phase 1
**Requirements**: OBS-01, OBS-02, OBS-03, OBS-04
**Success Criteria** (what must be TRUE):
  1. User can retrieve connection metrics -- GetMetrics() returns Latency, LastTickAge, ReconnectCount
  2. User can set per-request timeout -- SendRequest accepts timeout option different from default
  3. Events have priority levels -- when EventChannel is full, low-priority events are dropped before high-priority
  4. User can configure event buffer size -- EventBufferSize is settable via client option

**Plans**: 3 plans

Plans:
- [x] 02-observability-01-PLAN.md -- OBS-01: ConnectionMetrics struct, GetTickIntervalMs/GetStaleMultiplier helpers, ReconnectManager.AttemptCount(), client.GetMetrics()
- [x] 02-observability-02-PLAN.md -- OBS-02: SendRequest variadic opts + WithRequestTimeout; OBS-03: Event priority levels with priority-based dispatcher and drop logic
- [ ] 02-observability-03-PLAN.md -- OBS-04: Verify EventBufferSize configurable via WithEventBufferSize (already implemented)

### Phase 3: Client Struct Refactor

**Goal**: Maintainable client struct with clear separation of concerns
**Depends on**: Phase 2
**Requirements**: API-01, API-02
**Success Criteria** (what must be TRUE):
  1. Client struct is organized into logical sub-structs -- core, protocol, health, api are distinct groups with clear boundaries
  2. Close vs Disconnect semantics are documented -- user can understand which to call and what each guarantees

**Plans**: TBD

### Phase 4: Benchmarking and Fuzz Testing

**Goal**: Performance validation and regression detection for hot paths
**Depends on**: Phase 3
**Requirements**: TEST-01, TEST-02, TEST-03
**Success Criteria** (what must be TRUE):
  1. Hot-path benchmarks exist -- benchmark files for transport write/read, event dispatch, request correlation using b.Loop()
  2. Fuzz tests validate correctness -- round-trip assertions pass; corpus files exist in testdata/fuzz/
  3. Benchstat runs in CI -- performance regressions on hot paths are detected and reported

**Plans**: TBD

### Phase 5: Release Infrastructure

**Goal**: Library distribution ready for v1.0 release
**Depends on**: Phase 4
**Requirements**: REL-01, REL-02, REL-03
**Success Criteria** (what must be TRUE):
  1. GoReleaser configured for library mode -- blobs: true and gomod.proxy: true are set; release builds without error
  2. Semantic version tags exist -- v0.1.0, v0.2.0, ... progression leading to v1.0.0
  3. Changelog automation configured -- git-cliff generates CHANGELOG.md from conventional commits

**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation Hardening | 3/3 | Complete    | 2026-03-28 |
| 2. Observability | 2/3 | In progress | - |
| 3. Client Struct Refactor | 0/2 | Not started | - |
| 4. Benchmarking and Fuzz Testing | 0/3 | Not started | - |
| 5. Release Infrastructure | 0/3 | Not started | - |
