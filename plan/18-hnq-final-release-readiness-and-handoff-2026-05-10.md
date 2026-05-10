# HNQ Final Release Readiness and Handoff

Date: 2026-05-10 10:11 +07
Status: automated backend/frontend gates are documented as passing for HNQ QA parents; production promotion is still manual-validation gated. Follow-up fix card `t_1bd8c697` for frontend large-work action RBAC was checked afterward: role-policy now matches backend MVP view-only policy for `team_lead`/`user`/`viewer`, and local verification passed with `npx --yes tsx src/lib/auth/role-policy.test.ts && npx tsc --noEmit`.

## Scope

This handoff covers the HNQ planning workstream across:

- Backend repository: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Frontend repository: `/Users/sakdithat/Desktop/myproject/hotlines3`
- Backend source docs read for this release handoff:
  - `plan/11-k0-decision-matrix.md`
  - `plan/12-performance-rbac-monthly-plan-replan.md`
  - `plan/13-work-planning-and-large-job-prd-discovery.md`
  - `plan/14-session-handoff-2026-05-09.md`
  - `plan/15-team-plan-largework-implementation-plan.md`
  - `plan/16-planning-domain-api-contract.md`
  - `plan/17-contact-directory-implementation.md`
  - `plan/17-planning-frontend-ux-contract.md`

## Implemented backend flows

### Monthly plan correction

- Monthly plan remains active.
- `admin` and `super_admin` can manage monthly-plan operations broadly.
- `team_lead` and `user` can upload/download only their own team plan.
- `team_lead` and `user` can view all team rows for awareness.
- Monthly-plan lock semantics use the previous-month deadline from `MonthlyPlanSetting.lockDay`.
- Planning calendar monthly-plan edit permissions were realigned with monthly-plan service lock semantics.

### Team plan

- Team plan represents own-area planning work with no approval workflow.
- `user` and `team_lead` can add own-team plan items.
- Creator can edit their own item.
- `team_lead` can delete own-team items.
- Partial update validation now validates the final merged start/end date range before saving.
- Team plan items participate in planning calendar/day detail and daily-report source selection.

### Contact directory

- Authenticated users can browse active contact records.
- Directory response includes user/team contact context needed by the frontend.
- Users can update their own contact/profile fields according to the implemented contact-directory policy.
- `admin` is not re-expanded into broad user management; broad user/role management remains `super_admin`-only where implemented.

### Planning calendar

- Planning calendar combines implemented planning sources into a date-range view.
- Monthly-plan lock/action state is exposed consistently with backend monthly-plan policy.
- Team-plan and large-work visibility is role/team scoped.
- Daily-report draft source endpoint can return implemented planning sources for prefill.

### งานระดมทีม / large work

- `งานระดมทีม` is implemented as the large multi-team planning feature.
- Large-work items require valid multi-team planning context and date range validation.
- Backend MVP write policy is privileged-only: `super_admin` and `admin` can create/update/cancel; `team_lead`, `user`, and `viewer` are view-only for affected teams.
- Edit/cancel is limited to allowed item states.
- Large-work list pagination was fixed so repository totals are preserved and service-side visibility filtering does not double-slice pages.

## Implemented frontend flows

- Monthly-plan UI reflects the latest own-team upload/download rule for both `team_lead` and `user`.
- Planning UI exposes calendar/day planning concepts for team plan, monthly plan awareness, contact directory, daily-report prefill, and large-work planning.
- Contact-directory UI is present under `hotlines3` with service/types/hooks support.
- Daily task form has planning-source prefill support.
- Unauthenticated browser smoke confirmed protected routes redirect to login and invalid login shows an error without JS console errors.
- HNQ-02 found stale large-work action RBAC in `src/lib/auth/role-policy.ts` and `src/app/(main)/planning/page.tsx`; follow-up fix card `t_1bd8c697` was checked after this handoff. Large-work create/edit/cancel controls now use backend MVP policy: only `admin`/`super_admin` can act, while `team_lead`/`user`/`viewer` are view-only for relevant work.

## Role matrix for release validation

| Capability | super_admin | admin | team_lead | user | viewer |
|---|---:|---:|---:|---:|---:|
| Login | yes | yes | yes | yes | yes |
| Full user/role/password administration | yes | no | no | no | no |
| Monthly-plan broad manage/upload/download | yes | yes | no | no | no |
| Monthly-plan own-team upload/download | yes | yes | yes | yes | no |
| Monthly-plan awareness rows for all teams | yes | yes | yes | yes | yes/read-only if enabled |
| Team-plan create for own team | yes | no by default | yes | yes | no |
| Team-plan edit own-created item | yes | no by default | yes | yes | no |
| Team-plan delete/cancel own-team item | yes | no by default | yes | no | no |
| Planning calendar view | yes | yes | yes | yes | yes/read-only if enabled |
| Contact directory view | yes | yes | yes | yes | yes/read-only if enabled |
| Edit own contact/profile fields | yes | yes | yes | yes | yes if account supports profile edit |
| Create/update/cancel งานระดมทีม | yes | yes | no in MVP | no | no |
| View งานระดมทีม affecting own/visible teams | yes | yes | yes | yes | yes/read-only if enabled |
| Daily-report prefill from plan source | yes | yes | own/visible scope | own/visible scope | no |

## Final gates referenced from QA parents

Backend QA parent `t_330db94f` recorded these gates as passing:

```bash
go test ./...
go vet ./...
go build -o /tmp/hotlines-api main.go
bash scripts/test_smoke.sh
bash -n scripts/smoke.sh scripts/measure_performance.sh
BASE_URL=http://localhost:18081 bash scripts/smoke.sh
```

Notes from backend QA:

- Live smoke passed unauthenticated route checks.
- Authenticated live checks were skipped because safe `USERNAME`/`PASSWORD` and refresh token credentials were unavailable.
- Smoke coverage was expanded for contact directory, team plans, planning calendar, large-work items, and daily-report draft sources.
- Four backend blocker fixes were created and later marked done in recent Kanban history: large-work RBAC/date/state validation, large-work pagination, team-plan partial-update date validation, and planning-calendar monthly-plan lock semantics.

Frontend QA parent `t_616976aa` recorded these gates as passing:

```bash
npx --yes tsx src/lib/auth/role-policy.test.ts
npx --yes tsx src/types/monthly-plan.test.ts
npx --yes tsx src/types/planning-calendar.test.ts
npx tsc --noEmit
npm run build
npm run test:performance
git diff --check
```

Notes from frontend QA:

- Browser smoke passed login page load, unauthenticated monthly-plan redirect, invalid-login error visibility, and zero JS console errors.
- Full authenticated browser smoke was blocked because no safe test credentials were available for the configured remote API.
- HNQ-02 filed `t_1bd8c697` for stale frontend large-work action RBAC; the local follow-up check now passes role-policy and TypeScript verification.

## Current status

- Backend HNQ automated gates: passed per `t_330db94f`.
- Frontend HNQ automated gates: passed per `t_616976aa`; large-work RBAC follow-up `t_1bd8c697` was subsequently checked with role-policy tests and TypeScript.
- Release documentation: updated in this file and indexed from `plan/README.md`.
- Production promotion: not automatic. Manual validation is still required; frontend large-work RBAC is no longer the known blocking item from this handoff.

## Remaining risks

1. Authenticated backend smoke and authenticated browser QA require safe seeded credentials. Without credentials, authenticated paths are skipped or blocked, not proven.
2. R2-backed monthly-plan upload/download needs the target Cloudflare R2 configuration to prove production file I/O.
3. Frontend large-work action controls now match backend MVP policy in local role-policy checks: only `admin`/`super_admin` may create/edit/cancel; `team_lead`/`user`/`viewer` remain view-only. Recheck in authenticated browser validation before production.
4. The working trees contain many unrelated dirty/untracked implementation files from the broader workstream; review diffs carefully before committing or promoting.
5. Admin remains operations/monthly-plan scoped. Do not reintroduce broad admin user/master-data privileges without a new explicit user decision.

## Manual production validation steps

1. Start backend against target environment and confirm health.
2. Run backend smoke with safe credentials:

```bash
cd /Users/sakdithat/Desktop/myproject/backend-hotline
BASE_URL=<target-api-url> USERNAME=<safe-test-login> PASSWORD=<redacted> TEAM_ID=<safe-team-id> bash scripts/smoke.sh
```

3. Run frontend production build against the same target backend:

```bash
cd /Users/sakdithat/Desktop/myproject/hotlines3
npx tsc --noEmit
npm run build
```

4. Browser smoke by role:
   - `super_admin`: full management works; can manage monthly plans and งานระดมทีม.
   - `admin`: monthly-plan operations and งานระดมทีม management work; no broad user/role/password administration.
   - `team_lead`: can view all monthly-plan rows, upload/download own-team monthly plan only, manage own-team team plans, and view affected งานระดมทีม without write controls.
   - `user`: can view all monthly-plan rows, upload/download own-team monthly plan only, create/edit own team-plan items, and view affected งานระดมทีม without write controls.
   - `viewer` if retained: read-only only.
5. Validate monthly-plan file upload/download with real R2 settings.
6. Validate calendar date range display: team plan, monthly plan, and งานระดมทีม multi-day items show on each covered date.
7. Validate contact directory search/list and own-contact edit.
8. Validate daily-report prefill from implemented plan sources without mutating original plan items.
9. Rerun frontend role-policy tests, typecheck, build, and authenticated browser smoke for large-work action buttons before production promotion.
10. Run `git diff --check` in both repositories before committing/release handoff.

## Suggested next handoff actions

1. Run credentialed backend smoke and authenticated browser walkthrough against the target backend.
2. Rerun full frontend release gates including `npm run build` if any frontend files change after this handoff.
3. Commit or bundle only reviewed HNQ workstream files; avoid accidentally committing local Obsidian/Hermes folders unless intentionally part of the project vault.
