# Hotline Redesign Implementation Plan — 2026-05-19

Status: active
Board: `hotline-redesign-2026`

## Goal

Implement the finalized Hotline redesign requirements from the LLM/wiki source of truth.

## Source of truth

Wiki files:

1. `/Users/sakdithat/Desktop/myproject/wiki/queries/hotline-redesign-requirement-a-role-permission-2026-05-15.md`
2. `/Users/sakdithat/Desktop/myproject/wiki/queries/hotline-redesign-requirement-b-navigation-ia-2026-05-16.md`
3. `/Users/sakdithat/Desktop/myproject/wiki/queries/hotline-redesign-requirement-c-final-page-by-page-spec-2026-05-18.md`
4. `/Users/sakdithat/Desktop/myproject/wiki/queries/hotline-redesign-requirement-d-design-system-2026-05-19.md`
5. `/Users/sakdithat/Desktop/myproject/wiki/queries/hotline-redesign-implementation-plan-2026-05-19.md`

## Repos

- Backend/API/schema owner: `/Users/sakdithat/Desktop/myproject/hotline/backend-hotline`
- Frontend/API-only client: `/Users/sakdithat/Desktop/myproject/hotline/hotlines3`

## Worker overrides

- Existing frontend `CLAUDE.md` has older green/glassmorphism design rules. For this redesign, Requirement D overrides those old visual rules.
- Use official/deep blue primary, light neutral/white background, team-planning source blue, monthly-plan source teal.
- No Dashboard.
- Use `/work-report` with Thai label `รายงานการปฏิบัติงาน`; do not bring back old `/list` concept.
- `viewer` remains read-only and must not see write/download/export actions except allowed preview/call/copy.
- Frontend remains API-only; backend owns DB schema/migrations/API.

## Kanban task graph

Ready wave:

- `HRD-UX0` UX/UI component audit and implementation handoff — `dev-uxui`
- `HRD-B0` backend/API gap audit and safe contract fixes — `dev-backend-1`

Foundation wave:

- `HRD-F0` frontend design-system and app-shell foundation — `dev-frontend-1`, depends on UX0

Page implementation wave:

- `HRD-F1` `/planning` redesign — `dev-frontend-1`, depends on F0 + B0
- `HRD-F2` `/monthly-plan` redesign — `dev-frontend-2`, depends on F0 + B0
- `HRD-F3` `/daily-report` and `/work-report` redesign — `dev-frontend-1`, depends on F0 + B0
- `HRD-F4` `/contacts` and `/admin` redesign — `dev-frontend-2`, depends on F0 + B0

QA/release wave:

- `HRD-QA1` role/RBAC and responsive QA — `dev-qa`, depends on page implementation wave
- `HRD-PW1` browser/mobile smoke verification — `dev-playwright`, depends on page implementation wave
- `HRD-FINAL` final integration gate and handoff — `scc`, depends on QA/PW

## Quality gates

Frontend:

```bash
npm run lint
npx tsc --noEmit
npm run build
```

Backend:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
```
