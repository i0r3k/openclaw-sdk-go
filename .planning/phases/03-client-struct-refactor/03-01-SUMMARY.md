---
phase: 03-client-struct-refactor
plan: '01'
subsystem: pkg/client.go
tags: [refactor, client-struct, API-01]
dependency_graph:
  requires: []
  provides: [API-01]
  affects: [pkg/client.go, pkg/api/]
tech_stack:
  added: []
  patterns:
    - Sub-struct organization for logical field grouping
    - Immutable struct reorganization (backward compatible)
key_files:
  created: []
  modified:
    - pkg/client.go
decisions:
  - |
    D-01 (API namespace sub-struct): Grouped 15 API namespace fields under c.api sub-struct.
    Accessor methods unchanged - they delegate to c.api.chat, etc.
  - |
    D-02 (Remaining fields grouping): Grouped protocol fields (negotiator, policy, serverInfo, snapshot)
    under c.protocol and health fields (tickMonitor, gapDetector, tickHandlerUnsub) under c.health.
metrics:
  duration: ~3 minutes
  completed: '2026-03-29T03:59:00Z'
  tasks: 2
  files: 1
---

# Phase 03 Plan 01 Summary: Client Struct Reorganization

## One-liner

Refactored oversized `client` struct into four logical sub-struct groups (managers, api, protocol, health) following D-01 and D-02 decisions.

## What Was Done

### Task 1: Redefine client struct with sub-structs

Redefined the `client` struct (lines 396-437) with four sub-struct groups:
- `managers` - already existed, unchanged
- `api` - 15 API namespace fields (chat, agents, sessions, config, cron, nodes, skills, devicePairing, browser, channels, push, execApprovals, system, secrets, usage)
- `protocol` - negotiator, policy, serverInfo, snapshot
- `health` - tickMonitor, gapDetector, tickHandlerUnsub

Removed stale comment "// New fields for Phase 6.1".

### Task 2: Update all internal references

Updated all internal references throughout pkg/client.go to use new sub-struct paths:
- NewClient() initialization: `c.protocol.negotiator`, `c.protocol.policy`, `c.api.chat`, etc.
- Connect(): `c.health.tickMonitor`, `c.health.gapDetector`, `c.health.tickHandlerUnsub`
- Disconnect(): `c.health.tickHandlerUnsub`, `c.health.tickMonitor`, `c.health.gapDetector`, `c.protocol.negotiator`, `c.protocol.serverInfo`, `c.protocol.snapshot`
- SendRequest(): `c.protocol.policy`
- GetServerInfo(): `c.protocol.serverInfo`
- GetSnapshot(): `c.protocol.snapshot`
- GetPolicy(): `c.protocol.policy`
- GetTickMonitor(): `c.health.tickMonitor`
- GetGapDetector(): `c.health.gapDetector`
- GetMetrics(): `c.health.tickMonitor` (all references)
- processServerInfo(): `c.protocol.serverInfo`, `c.protocol.snapshot`, `c.protocol.negotiator`, `c.protocol.policy`, `c.health.gapDetector`

All 15 API accessor methods (Chat(), Agents(), etc.) updated to return `c.api.*` paths.

## Verification

- `go build ./pkg/...` - passed
- `go test ./pkg/... -race` - all 10 packages passed
- `go vet ./pkg/client.go` - passed
- `gofmt` formatting - applied

## Deviations from Plan

None - plan executed exactly as written.

## Deviation: None

## Files Modified

| File | Change |
|------|--------|
| pkg/client.go | 99 insertions, 95 deletions - struct reorganization |

## Commit

`a797aac` - refactor(03-01): reorganize client struct into logical sub-structs

## Self-Check: PASSED

- [x] client struct has 4 sub-struct groups
- [x] No flat protocol/health/api fields remain
- [x] All 15 API accessors return c.api.* paths
- [x] All tickMonitor references use c.health.tickMonitor
- [x] All gapDetector references use c.health.gapDetector
- [x] All protocol.* references use c.protocol.*
- [x] go build passes
- [x] go test -race passes
