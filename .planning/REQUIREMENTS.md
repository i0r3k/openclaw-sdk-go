# Requirements: OpenClaw SDK Go

**Defined:** 2026-03-28
**Core Value:** Go developers can integrate the OpenClaw platform in under 10 lines of code

## v1 Requirements

### Foundation (P1 — must fix before production use)

- [x] **FOUND-01**: Client-side rate limiting — Add `RequestRateLimiter` interface with `WithRateLimit()` option to prevent server rejection under load
- [x] **FOUND-02**: Retry budget — Add `MaxRetries` field to `ReconnectConfig`; replace unlimited (MaxRetries=0) with sensible default of 10
- [x] **FOUND-03**: TLS CRL validation — Implement actual `CheckCertificateRevocation` or mark stub with explicit comment explaining limitation
- [x] **FOUND-04**: Unbounded pending requests limit — Add max pending requests limit with `ErrTooManyPendingRequests` error
- [x] **FOUND-05**: InsecureSkipVerify warning — Add warning log when TLS InsecureSkipVerify is enabled

### Observability (P2 — v1.x)

- [x] **OBS-01**: Connection health metrics — `ConnectionMetrics` struct with Latency, LastTickAge, ReconnectCount; expose via `GetMetrics()` method
- [ ] **OBS-02**: Per-request timeout — Allow different timeout per request via `SendRequest(ctx, req, opts...)` with timeout option
- [ ] **OBS-03**: Graceful degradation — Event priority levels; when EventChannel is full, drop low-priority events first
- [ ] **OBS-04**: Event buffer configuration — Configurable `EventBufferSize` via client option

### API Stability (P2 — v1.x)

- [x] **API-01**: Client struct refactor — Group oversized `client` struct into logical sub-structs (`core`, `protocol`, `health`, `api`)
- [x] **API-02**: Close/Disconnect disambiguation — Merge or clearly document `Close()` vs `Disconnect()` semantics in ConnectionManager

### Testing & Quality (P2 — v1.x)

- [ ] **TEST-01**: Hot-path benchmarks — Add `*_bench_test.go` files for transport write/read, event dispatch, request correlation using `b.Loop()` (Go 1.24+)
- [ ] **TEST-02**: Fuzz test depth — Add round-trip correctness assertions to existing fuzz tests; add corpus files in `testdata/fuzz/`
- [x] **TEST-03**: Benchmark CI integration — Add `benchstat` to CI for regression detection on hot paths

### Release Infrastructure (P2 — v1.x)

- [ ] **REL-01**: GoReleaser library mode — Complete `.goreleaser.yaml` with `blobs: true` and `gomod.proxy: true` for library distribution
- [ ] **REL-02**: Semantic versioning tags — Create git tag `v0.1.0` progression to `v1.0.0`; configure `goreleaser.yml` with `versioning: semver`
- [ ] **REL-03**: CHANGELOG automation — Configure `git-cliff` or equivalent for automatic changelog generation from conventional commits

## v2 Requirements (Future)

### Advanced Features

- **ADV-01**: Middleware/interceptors — Request/response hooks for logging, metrics, auth refresh
- **ADV-02**: OpenTelemetry tracing — Span propagation for WebSocket frames
- **ADV-03**: Circuit breaker — Prevent cascade failures when gateway is unhealthy
- **ADV-04**: Request deduplication — Idempotency key support for retry-safe requests
- **ADV-05**: WebSocket compression — Per-message deflate extension support

### Integration

- **INT-01**: Integration tests — Test against real OpenClaw gateway with live connection
- **INT-02**: Prometheus endpoint — Expose metrics in Prometheus format for scraping

## Out of Scope

| Feature | Reason |
|---------|--------|
| Connection pooling | Single connection per client; pooling is user responsibility |
| HTTP/REST fallback | WebSocket-only; gateway does not support REST |
| Binary protocol | JSON is the gateway wire format; performance not critical path |
| Built-in retry for all errors | Only reconnect on disconnect; application-level retry is user responsibility |
| Multiple simultaneous connections | Single connection architecture; multiplex at application level |

## Active Anti-Patterns (from Research)

| Anti-Pattern | Prevention |
|--------------|------------|
| Concurrent write corruption | All writes serialized behind mutex; `go test -race` must pass |
| Failing to read connection | `readLoop` runs until `Close()`, never exits early on error |
| Write deadline corruption | After write timeout, close and reconnect; never reuse connection |
| Channel + lock deadlock | Rule: never send to channel while holding lock; release lock before send |
| Timer leak in reconnect | Always `defer timer.Stop()` in reconnect loop |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | Phase 1 | Complete |
| FOUND-02 | Phase 1 | Complete |
| FOUND-03 | Phase 1 | Complete |
| FOUND-04 | Phase 1 | Complete |
| FOUND-05 | Phase 1 | Complete |
| OBS-01 | Phase 2 | Complete |
| OBS-02 | Phase 2 | Pending |
| OBS-03 | Phase 2 | Pending |
| OBS-04 | Phase 2 | Pending |
| API-01 | Phase 3 | Complete |
| API-02 | Phase 3 | Complete |
| TEST-01 | Phase 4 | Pending |
| TEST-02 | Phase 4 | Pending |
| TEST-03 | Phase 4 | Complete |
| REL-01 | Phase 5 | Pending |
| REL-02 | Phase 5 | Pending |
| REL-03 | Phase 5 | Pending |
| ADV-01–ADV-05 | Future | Deferred |
| INT-01–INT-02 | Future | Deferred |

**Coverage:**
- v1 requirements: 17 total
- Mapped to phases: 17
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-28*
*Last updated: 2026-03-28 after research synthesis*
