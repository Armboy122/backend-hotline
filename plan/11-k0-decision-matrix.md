# K0 Decision Matrix — Hotline PRD Open Questions

> Generated: 2026-05-09
> Source: `plan/10-hotline-prd-latest-and-kanban-scope.md` Q1-Q10
> Purpose: Record user-confirmed PRD decisions and unblock K1+ implementation tasks.

---

## Current State Summary

| Aspect | What exists now |
|---|---|
| Role model | `User.Role` = string, values `admin` \| `user` \| `viewer` |
| Route guard | `RequireRole("admin")` on admin routes only |
| Team lead | No concept in code; user confirmed new role `team_lead` |
| Super admin | No concept in code; user confirmed role string `super_admin` with single-instance invariant |
| Monthly plan | File upload per team per year/month, status tracking |
| Frontend | `hotlines3` (Next.js SPA), AdminGuard checks `role === 'admin'` |
| Old frontend | `hotline-2` deleted from workspace after user confirmation |

---

## User-Confirmed Decisions — 2026-05-08

| Q | Decision | Implementation implication |
|---|---|---|
| Q1 | Use role string `super_admin` directly | Update backend/frontend role constants and JWT/guards |
| Q2 | `team_lead` is a separate role | Add role policy and team-scoped monthly-plan permissions |
| Q3 | Monthly plan is the next-month planning submission system | Extend existing monthly plan flow to capture destination/location, date range, attachments, notes |
| Q4 | No additional approval workflow | Keep current monthly-plan status behavior; do not add approval state machine |
| Q5 | User sees only own team's tasks | Backend list/query filters and frontend screens must be team-scoped for users/team_leads |
| Q6 | Only `super_admin` can create/promote admins | Admin role management routes guarded by super_admin only |
| Q7 | Only `super_admin` can reset passwords | Admin cannot reset passwords |
| Q8 | First `super_admin` created locally once against real Neon DB from env | Add one-time local CLI/bootstrap command; no secrets committed |
| Q9 | No audit log for this round | Defer audit table/service |
| Q10 | Active frontend is `hotlines3`; delete `hotline-2` | Workspace cleaned; no code should target `hotline-2` |

Neon MCP check found project `hotlines3` (`bitter-lake-05690037`), default branch `production` (`br-snowy-thunder-a1teoes9`), PostgreSQL 17. Current schema already has `User.role`, `User.teamId`, `TaskDaily.teamId`, `MonthlyPlanSetting.reminderStartDay=20`, `MonthlyPlanSetting.lockDay=23`, and file-based `PlanFile` with `description`; it does not yet enforce one `super_admin` or store structured monthly-plan destination/date-range fields.

## Revision — 2026-05-09 Performance/RBAC/Monthly Plan

Supersedes parts of the 2026-05-08 role/monthly-plan interpretation:

- Performance investigation/fix must run before additional feature implementation.
- `super_admin` is the only full system administrator: full read/CRUD through application policy, user/role customization, password reset, master data, dashboard/report, and monthly-plan override.
- `admin` is not a broad administrator. Admin is an operational monthly-plan manager only: upload/download/replace monthly master plans and view monthly-plan status needed for operations.
- `admin` must not manage users, roles, passwords, broad master data, or super-admin-only global dashboard/CRUD actions.
- Correction: monthly-plan is still active for teams and starts with the June 2026 plan cycle; upload/manage permission belongs to admin/super_admin broadly; team_lead and user can upload only their own team plan.
- `team_lead`/`user` can see monthly-plan rows for every team. `team_lead`/`user` can upload/download only own-team files.
- `admin` can edit monthly-plan settings and can still upload/manage plans; `team_lead`/`user` can upload only own-team plans.
- Monthly-plan page target is full-year 2569 / 2026 view with all 12 months, role-aware actions, and lock rules based on `MonthlyPlanSetting.lockDay = 23` of the previous month.
- Detailed replan: `plan/12-performance-rbac-monthly-plan-replan.md`.

---

## Decision Matrix

### Q1 — Role naming: `super_admin` vs flag?

> ต้องการ role ชื่อ `super_admin` ตรง ๆ ในระบบไหม หรือใช้ `admin` + flag เช่น `isSuperAdmin`?

**Impact:** DB migration, JWT claims, route guard logic, frontend type/guard

| Option | Pros | Cons |
|---|---|---|
| A: `super_admin` as role string | Simple to query (`WHERE role = 'super_admin'`), clear in logs, easy JWT check | Changes `oneof` validation in DTO, needs data migration for existing admin |
| B: `admin` + `is_super_admin bool` flag | No DTO validation change, backward compatible | Extra column + composite check on every admin route guard, harder to audit |

**Decision:** Option A — use explicit `super_admin` role string.

**Blocks:** unblocked for K1

---

### Q2 — Team lead: new role vs flag?

> "หัวหน้าแต่ละทีม" คือ role ใหม่ (`team_lead`) หรือเป็น user ที่มี flag/permission ในทีม?

**Impact:** User model, team membership model, monthly plan submission policy, frontend team management UI

| Option | Pros | Cons |
|---|---|---|
| A: `team_lead` as a role | Simple role check for plan submission | A user can only lead one team (or needs multi-team membership), less flexible if someone is both lead and regular member in different teams |
| B: `is_team_lead bool` on a team membership/assignment | One user can be lead of Team A and member of Team B | Needs a team membership join table (currently Users have `TeamID` — single team) |
| C: `team_lead_id` on Team model | No new role, no membership table — just FK | A user leads exactly one team, simple query, but can't be lead of two teams |

**Current DB state:** `User.TeamID` is a single FK (one user, one team).

**Decision:** Option A — use `team_lead` as a separate role. Keep `teamId` as the team-scoping anchor.

**Blocks:** unblocked for K2

---

### Q3 — Monthly plan: file-only vs structured?

> หัวหน้าทีมส่งแผนเดือนถัดไปเป็นไฟล์แนบเหมือนระบบปัจจุบันพอไหม หรืออยากให้กรอกเป็นรายการงานล่วงหน้าแบบ structured plan?

**Impact:** New model + API + UI if structured; none if file-only

| Option | Pros | Cons |
|---|---|---|
| A: File-only (current) | No new model/API, ship faster | Can't query/filter/compare plan items, admin can't see what's planned without opening files |
| B: Structured plan items | Queryable, filterable, dashboard can compare plan vs actual, admin can review without downloading | Significant new feature: model, API, UI, likely 2+ phases of work |

**Decision:** Extend current monthly plan flow to accept planning information in-system: destination/location, start date, end date, attachments, and notes. Do not require LINE submission.

**Blocks:** unblocked for K2

---

### Q4 — Monthly plan approval workflow?

> แผนงานเดือนถัดไปต้องมี approval ไหม? ถ้ามี ใคร approve?

**Impact:** New approval model + state machine + UI if yes

| Option | Pros | Cons |
|---|---|---|
| A: No approval (current) | Ship faster, matches current behavior | Admin can't formally accept/reject plans |
| B: Simple admin approval | Accountability, clear status | New model: PlanApproval, new API endpoints, new UI states |

**Decision:** No additional approval workflow. Use existing monthly plan flow/status behavior.

**Blocks:** unblocked for K2

---

### Q5 — User task visibility scope?

> user ธรรมดาควรเห็น task เฉพาะทีมตัวเอง หรือเห็นทุกทีมแต่แก้ได้เฉพาะของตัวเอง/ทีมตัวเอง?

**Impact:** Backend filter logic, frontend list pages

| Option | Behavior | Pros | Cons |
|---|---|---|---|
| A: See own team only | User queries auto-filtered by `team_id` | Simple, no accidental cross-team edits | Can't see what other teams are doing |
| B: See all, edit own only | Read access to all teams, write scoped to own team | Cross-team visibility, better coordination | Need separate read/write scoping logic |
| C: See all, edit own only, admin sees all+edits all | Like B but admin has full access | Practical for small orgs | Slightly more complex permission logic |

**Decision:** Option A — normal `user` sees only tasks for their own team.

**Blocks:** unblocked for K3

---

### Q6 — Can admin create other admins?

> admin ธรรมดาควรสร้าง admin คนอื่นได้ไหม หรือเฉพาะ super_admin เท่านั้น?

**Impact:** User management policy in K1

| Option | Behavior |
|---|---|
| A: Only `super_admin` can create/promote to `admin` | Strict hierarchy, clear escalation path |
| B: `admin` can create `admin` | Flat admin pool, simpler but riskier |

**Decision:** Option A — only `super_admin` can create or promote admins.

**Blocks:** unblocked for K1

---

### Q7 — Can admin reset passwords? For whom?

> admin ธรรมดาควร reset password ให้ user ได้ไหม? แล้ว reset ให้ admin คนอื่นได้หรือไม่?

**Impact:** Password reset policy in K1/K4

| Option | Behavior |
|---|---|
| A: Admin resets `user` only, super_admin resets everyone | Clear boundary, admin can't lock out other admins |
| B: Admin resets `user` + `viewer`, super_admin resets everyone | Includes viewer role |
| C: Anyone resets anyone with same or lower role | Generic but complex |

**Decision:** Only `super_admin` can reset passwords. Admin cannot reset passwords.

**Blocks:** unblocked for K1

---

### Q8 — How is the first super_admin created?

> super_admin คนแรกจะถูกส้างอย่างไร?

| Option | Pros | Cons |
|---|---|---|
| A: DB seed / migration | Automatic, no manual step | Seed data in version control might leak |
| B: CLI command: `hotlines-api bootstrap --super-admin` | Explicit, one-time, no data in repo | Needs CLI flag handling |
| C: Config file / env var | Simple | Credential in config is risky |

**Decision:** Local one-time bootstrap against real Neon DB using env connection. Prefer CLI/local command that refuses to run if a `super_admin` already exists.

**Blocks:** unblocked for K1

---

### Q9 — Audit logging?

> ต้องเก็บ audit log สำหรับการเปลี่ยน role/reset password/upload monthly plan ไหม?

**Impact:** New model + service if yes

| Option | Pros | Cons |
|---|---|---|
| A: No audit log for MVP | Ship faster | No trail for security-sensitive actions |
| B: Simple audit log table (`who, what, whom, when`) | Accountability, lightweight | Extra table + write on every sensitive action |

**Decision:** No audit log for this round.

**Blocks:** unblocked; audit log deferred

---

### Q10 — Keep old frontend `hotline-2`?

> ต้องการเก็บ frontend เก่า `/Users/sakdithat/Desktop/myproject/hotline-2` ไว้ไหม?

**Impact:** Repo housekeeping only, no code impact

| Option | Behavior |
|---|---|
| A: Archive/delete `hotline-2` | Cleaner workspace, no confusion |
| B: Keep for reference | May need for reference, but shouldn't block anything |

**Decision:** Delete/archive old frontend. `hotline-2` has been removed from `/Users/sakdithat/Desktop/myproject`; `hotlines3` is the active frontend.

**Blocks:** Nothing

---

## Blocking Summary

All Q1-Q10 product decisions are answered and K1-K3 are unblocked.

### K1 Auth/RBAC now has enough scope to start
- `super_admin` role string
- exactly one `super_admin`
- only `super_admin` creates/promotes admins
- only `super_admin` resets passwords
- local one-time bootstrap against Neon env DB
- no audit log this round

### K2 Team Lead + Monthly Plan now has enough scope to start
- `team_lead` is a separate role
- monthly plan is next-month planning submission in the existing monthly-plan system
- capture destination/location, date range, attachments, and notes
- no additional approval workflow

### K3 Daily Task Hardening now has enough scope to start
- normal users see only own-team tasks
- team scoping should use existing `teamId` fields and indexes

### Repo housekeeping complete
- active frontend: `/Users/sakdithat/Desktop/myproject/hotlines3`
- old frontend `/Users/sakdithat/Desktop/myproject/hotline-2` removed from workspace

---

## Confirmed Role Matrix

| Capability | super_admin | admin | team_lead | user | viewer |
|---|---|---|---|---|---|
| Create/promote admin | YES | NO | NO | NO | NO |
| Reset password for others | YES | NO | NO | NO | NO |
| Manage master data | YES | YES | NO | NO | NO |
| Manage users (non-admin) | YES | YES (no password reset) | NO | NO | NO |
| View dashboard | YES | YES | scoped | scoped | scoped/read-only |
| View all teams' tasks | YES | YES | NO | NO | read-only if retained |
| View own team's tasks | YES | YES | YES | YES | YES |
| Edit own team's tasks | YES | YES | YES | YES | NO |
| Edit any team's tasks | YES | YES | NO | NO | NO |
| Submit monthly plan for own team | YES | YES | YES | NO | NO |
| Review monthly plan submissions/status | YES | YES | scoped own-team | NO | NO |
| Disable/delete users | YES | YES (non-admin only) | NO | NO | NO |
| Change own password | YES | YES | YES | YES | YES |

> `team_lead` is a separate role and is still scoped by `teamId`. Normal users only see their own team's tasks.

---

## MVP Definition of Done (proposed)

1. Role model implemented: `super_admin`, `admin`, `team_lead`, `user`, `viewer` with the matrix above
2. One-super-admin invariant enforced at DB + service level
3. Password reset policy enforced: only `super_admin` resets passwords
4. First super_admin bootstrap works locally/once against Neon env DB
5. All existing `/v1` API routes continue to work
6. Quality gates pass: `go test ./... && go vet ./... && go build -o /tmp/hotlines-api main.go`
7. Frontend `hotlines3` updated with new role types and guards
8. Frontend quality gate: `npx tsc --noEmit`
