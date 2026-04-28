# Phase A - Wave 1 Task Board

## Mission
Start the safest first wave for `backend-hotline` without breaking existing `/v1` behavior. This wave is focused on guardrails, test coverage, and low-risk fixes.

## Agent split

### Agent α — Sonnet 4.6
- ID: A1
- Status: completed
- Task: Architecture boundary tests
- Goal: lock layering rules so future work cannot import the wrong dependencies
- Files: `internal/architecture/architecture_test.go`
- Session: `proc_7b09ef3b0c90`
- Result: `go test ./internal/architecture -v` passed; `go test ./...` passed after fixing a task handler import alias
- Notes: hardest test task in Wave 1; requires careful import scanning logic

### Agent β — Sonnet 4.6
- ID: A4
- Status: completed
- Task: TaskDaily list behavior tests
- Goal: lock current TaskDaily behavior before any deeper refactor
- Files: `internal/app/task/usecase/list_tasks_test.go`, `internal/handlers/v1/error_mapping_test.go`
- Session: `proc_23f72db25647`
- Result: list pagination normalization, filter passthrough, response compatibility, and usecase error propagation are covered; targeted package tests and `go test ./...` passed
- Notes: Claude session hit the org monthly usage limit, so the remaining work was finished locally and verified with Go tests

### Agent γ — Lightweight model
- ID: A2
- Status: completed
- Task: Config and middleware tests
- Goal: stabilize config, auth middleware, timeout, recovery, and cache behavior
- Files: `internal/config/config_test.go`, `internal/middleware/middleware_test.go`
- Result: config expansion, cache headers, recovery, timeout, and auth guardrails are covered; `go test ./internal/config ./internal/middleware -v` passed
- Notes: routine test work, good fit for a lighter agent

### Agent δ — Lightweight model
- ID: A3
- Status: completed
- Task: Standard response and error mapping tests
- Goal: prevent response envelope drift
- Files: `internal/dto/response_test.go`, `internal/handlers/v1/error_mapping_test.go`
- Result: envelope serialization and task list response/error mapping are covered; `go test ./internal/dto ./internal/handlers/v1 -v` passed
- Notes: mostly shape assertions and regression coverage

### Agent ε — Lightweight model
- ID: A5
- Status: completed
- Task: Smoke script skeleton
- Goal: create a simple script that can verify representative `/v1` endpoints
- Files: `scripts/smoke.sh`
- Result: shell smoke script added and syntax-checked with `bash -n scripts/smoke.sh`
- Notes: script-only task, low risk

## Execution order
1. Start α and β immediately in parallel.
2. Start γ, δ, and ε in parallel as soon as the task board is approved.
3. Do not start TaskDaily refactor or MonthlyPlan work until Wave 1 tests are in place.

## Success criteria
- `go test ./...` remains green or gets closer to green without widening scope.
- No `/v1` contract changes.
- New tests clearly capture current behavior.
- Repo notes and Obsidian notes reflect the work completed.
