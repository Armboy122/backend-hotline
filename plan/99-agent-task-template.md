# Agent Task Template

Copy this when assigning one task to Codex, Claude Code, GLM, or another worker.

```markdown
You are working in:
`/Users/sakdithat/Desktop/myproject/backend-hotline`

Branch: `[current branch]`

Read first:
- `plan/README.md`
- `plan/00-backend-architecture.md`
- `plan/01-current-state-and-done-checklist.md`
- [phase file for the assigned task]

Assigned task:
- ID: [A1.1 / B2.3 / C4.2 / etc.]
- Objective: [one sentence]

Hard rules:
- Preserve existing `/v1` API routes and response envelopes unless the task explicitly says otherwise.
- Use TDD where practical: write a failing test first, confirm failure, then implement.
- Keep domain/usecase code free of Gin, GORM, Viper, and AWS SDK imports.
- Do not add business logic to `internal/router`.
- Do not return `internal/models` types from repository interfaces.
- Do not rename DB columns/tables during architecture refactor tasks.
- Do not expose password hashes.
- Keep changes scoped to the assigned task only.

Expected files:
- Create: [...]
- Modify: [...]
- Tests: [...]

Verification commands:
```bash
go test ./path/to/package -run SpecificTest -v
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```

Completion report format:
1. Tests added
2. Implementation changed
3. Commands run and result
4. Any follow-up risks/open decisions
```

## Review checklist for completed agent work

- [ ] Does the work satisfy the exact assigned task only?
- [ ] Are tests meaningful for the risk of the change?
- [ ] Does `go test ./...` pass?
- [ ] Does `go vet ./...` pass if relevant?
- [ ] Is there any new DB/business logic in `internal/router`?
- [ ] Does any usecase import Gin?
- [ ] Does any domain package import GORM/Gin/Viper/AWS SDK?
- [ ] Are errors mapped to the standard response envelope?
- [ ] Are existing `/v1` routes and JSON field names preserved?
- [ ] Are password hashes and secrets kept out of responses/logs?
