# HNP-02 Planning Frontend UX Contract

Date: 2026-05-10
Status: frontend implementation contract; no production UI code written by this card
Scope: frontend routes, screens, components, role-visible actions, and API integration points for planning calendar, team plan, contact directory, งานระดมทีม, and corrected monthly-plan own-team policy.

Source docs: `plan/11-k0-decision-matrix.md`, `plan/12-performance-rbac-monthly-plan-replan.md`, `plan/13-work-planning-and-large-job-prd-discovery.md`, `plan/14-session-handoff-2026-05-09.md`, `plan/15-team-plan-largework-implementation-plan.md`, `plan/16-planning-domain-api-contract.md`.

Frontend repo inspected: `/Users/sakdithat/Desktop/myproject/hotlines3`.

---

## 1. Current frontend baseline

Current Next.js routes:

- `/` — daily report entry form.
- `/list` — daily report list and export.
- `/monthly-plan` — authenticated yearly monthly-plan page for พ.ศ. 2569 / ค.ศ. 2026.
- `/admin` — admin menu hub.
- `/admin/monthly-plan` — monthly-plan admin/settings entry.
- `/admin/dashboard`, `/admin/operation-centers`, `/admin/peas`, `/admin/stations`, `/admin/feeders`, `/admin/job-types`, `/admin/job-details`, `/admin/task-daily` — system-admin/master-data/report areas.
- `/login` — auth route.

Current navigation facts:

- Main navigation is configured in `src/config/navigation.tsx`.
- Route visibility is delegated to `canAccessMainNavigationItem()` in `src/lib/auth/role-policy.ts`.
- Admin route access is guarded by `AdminGuard` and `canAccessAdminRoute()`.
- Current main nav labels are daily report, list, monthly plan, and admin.

Current role-policy facts:

- Roles are `super_admin`, `admin`, `team_lead`, `user`, and `viewer`.
- `super_admin` is the only system-admin role.
- `admin` is a monthly-plan operations manager, not broad system admin.
- `team_lead` and `user` can access normal authenticated pages.
- Existing monthly-plan frontend policy must be corrected/kept aligned so both `team_lead` and `user` can upload/download own-team monthly-plan files only.

Current monthly-plan facts:

- `/monthly-plan` uses `useMonthlyPlanYearOverview(year)` and renders all 12 months in one page.
- Team rows are built with `buildMonthlyPlanTeamRows()`.
- Admin/super_admin are treated as plan managers through `isPrivilegedAdmin()`.
- The route already shows all team rows for awareness.
- Important correction for implementation: `canManageMonthlyPlanFile()` currently grants own-team file management only to `team_lead`; HNP-02+ implementation must update policy/tests so `user` receives the same own-team upload/download permission where lock settings allow it.

---

## 2. Route plan

### 2.1 Main authenticated routes

Add these routes under `src/app/(main)/`:

- `/planning` — default planning calendar page.
- `/planning/team-plans` — list/table view for team-plan items; can be a tab or subroute from calendar.
- `/planning/team-plans/[id]` — detail/edit view or modal-backed route for direct links.
- `/planning/large-work` — งานระดมทีม list/detail entry.
- `/contacts` — team/user contact directory.

Keep `/monthly-plan` as the yearly document-submission page. The planning calendar should link into `/monthly-plan` for monthly-plan file operations rather than duplicating the whole yearly upload UI.

### 2.2 Admin routes

Keep existing admin split:

- `super_admin` sees broad admin/master-data/dashboard routes.
- `admin` sees only monthly-plan operations by default.

Do not put normal team planning behind `/admin`. Team planning is operational work and must be available from the normal authenticated shell for `team_lead` and `user`.

### 2.3 Navigation changes

Main nav should become:

- `บันทึกข้อมูล` → `/`.
- `รายการ` → `/list`.
- `ปฏิทินแผนงาน` or `แผนงาน` → `/planning`.
- `แผนเดือน` → `/monthly-plan`, if screen space permits; otherwise expose from planning calendar and profile/menu.
- `รายชื่อทีม` → `/contacts`, if screen space permits; otherwise expose from planning calendar header/action.
- `จัดการข้อมูล` → `/admin`, visible only to `super_admin` and `admin` per current guard.

Mobile bottom nav should keep the highest-frequency items only: daily report, list, planning calendar, monthly plan or contacts depending on operator feedback, and admin only for roles that can access it.

---

## 3. Screen contracts

### 3.1 Planning calendar `/planning`

Purpose: Google-Calendar-like monthly planning overview.

Required UI:

- Month/year selector with default current month.
- Calendar grid that shows badges/counts per date for:
  - team plan
  - monthly plan
  - งานระดมทีม
- Day detail drawer/sheet on click/tap.
- Filter controls for plan type and team.
- Primary action button: add team plan for `user`/`team_lead`; add team plan/งานระดมทีม for permitted admin roles.
- Empty-state copy for dates with no work.

API/hooks:

- Add `planningCalendarService.getItems({ from, to, teamId?, type? })`.
- Add `usePlanningCalendarItems()` using one range request for the visible month, not one request per day.
- Calendar items should use a discriminated union: `team_plan`, `monthly_plan`, `large_work`.

Acceptance:

- Multi-day items render on every date in the inclusive range.
- Items without time still render on the date.
- Calendar does not create 30+ serial daily calls.
- Role filtering hides/disables actions, not awareness rows, unless backend policy removes a row.

### 3.2 Team plan

Routes: `/planning/team-plans` and `/planning/team-plans/[id]`, with modal/drawer create/edit allowed from `/planning`.

Required UI:

- List filters: date range, team, status.
- Create/edit form fields:
  - title/work type
  - start date
  - optional end date
  - optional time
  - location text
  - optional PEA / operation center / feeder / station
  - notes
- Detail view shows creator, team, status, date range, location, electric-area metadata, and daily-report prefill action if available.
- Cancel/delete should be a soft state transition, not a destructive default.

Components:

- `TeamPlanForm`.
- `TeamPlanList`.
- `TeamPlanCard`.
- `TeamPlanDetailDrawer`.
- `TeamPlanStatusBadge`.

Actions:

- `user`: create own-team team plan; edit own-created item; cannot cancel/delete unless later confirmed.
- `team_lead`: create own-team item; edit/cancel own-team items.
- `admin`: read for operations by default; no broad team-plan mutation unless later enabled.
- `super_admin`: manage all.
- `viewer`: read-only awareness.

### 3.3 Monthly plan integration

Keep `/monthly-plan` as the document/yearly workflow.

Required HNP-02 correction:

- `team_lead` and `user` can view all team rows.
- `team_lead` and `user` can upload/download only own-team monthly-plan files when not locked.
- `admin` and `super_admin` can manage broadly.
- `viewer` is read-only with no upload action.

Frontend implementation notes:

- Update role-policy tests before changing `canManageMonthlyPlanFile()` and any row-action helpers.
- Ensure `PlanFileRow.canDownload` allows own-team download for both `team_lead` and `user`.
- Ensure upload dialog receives only the actor's own team for non-admin roles.
- Calendar monthly-plan items should deep-link to `/monthly-plan` with year/month/team context if query params are added.

### 3.4 Contact directory `/contacts`

Purpose: searchable contact list for coordination.

Required UI:

- Search box.
- Filters by team and role/position if backend supports them.
- Contact cards/list rows showing:
  - display name
  - position/title
  - phone number
  - team
  - role label where useful
- Own profile edit action.

Components:

- `ContactDirectoryPage`.
- `ContactSearchFilters`.
- `ContactCard`.
- `EditOwnContactDialog`.

Actions:

- All authenticated roles can view active contacts.
- `super_admin` can edit any contact/user fields.
- `admin`, `team_lead`, and `user` can edit their own personal/contact information.
- Normal roles cannot edit another user's contact info.
- `viewer` is read-only unless product later confirms viewer self-edit.

### 3.5 งานระดมทีม `/planning/large-work`

Purpose: large multi-team work planning.

Required UI:

- Thai label must be exactly `งานระดมทีม` in visible nav/screen copy.
- List view grouped by status/date.
- Create/edit form fields:
  - title
  - owner team
  - participating teams; must include at least one team besides owner
  - date/start date
  - optional end date
  - optional time
  - location
  - work type/details
  - notes
  - optional attachments if backend implements them
- Detail view shows owner team, participating teams, status, date range, location, and contact shortcuts.
- Calendar badges should clearly label this item type as `งานระดมทีม`.

Actions:

- MVP creation/update: `super_admin` and `admin`.
- `team_lead`/`user`: read items involving their team plus awareness rows as backend permits; no create/update in MVP unless later confirmed.
- `viewer`: read-only if retained.

---

## 4. Frontend file layout

Use existing frontend conventions:

```text
src/types/planning-calendar.ts
src/types/team-plan.ts
src/types/contact-directory.ts
src/types/large-work.ts
src/lib/services/planning-calendar.service.ts
src/lib/services/team-plan.service.ts
src/lib/services/contact-directory.service.ts
src/lib/services/large-work.service.ts
src/hooks/useQueries.ts
src/hooks/mutations/useTeamPlanMutations.ts
src/hooks/mutations/useContactDirectoryMutations.ts
src/hooks/mutations/useLargeWorkMutations.ts
src/features/planning-calendar/components/*
src/features/team-plan/components/*
src/features/contact-directory/components/*
src/features/large-work/components/*
```

Keep DTOs and services thin. Put display/view-model helpers near the feature, similar to `src/features/monthly-plan/utils.ts`.

---

## 5. Role-visible actions matrix

- Planning calendar view:
  - `super_admin`: all rows/actions.
  - `admin`: all awareness; create/manage งานระดมทีม in MVP; monthly-plan management; team-plan read by default.
  - `team_lead`: all awareness; own-team team-plan create/edit/cancel; own-team monthly-plan upload/download; involved large-work read.
  - `user`: all awareness; own-team team-plan create; edit own-created team plan; own-team monthly-plan upload/download; involved large-work read.
  - `viewer`: read-only awareness.

- Team-plan item:
  - create: `super_admin` all, `team_lead` own team, `user` own team.
  - edit: `super_admin` all, `team_lead` own team, `user` creator only.
  - cancel/delete: `super_admin` all, `team_lead` own team, `user` no.

- Monthly-plan file:
  - upload/download own team: `team_lead`, `user` when lock policy allows.
  - upload/download/manage all teams: `super_admin`, `admin`.
  - settings: `super_admin`, `admin`.

- Contact directory:
  - view: all authenticated roles.
  - edit self: `super_admin`, `admin`, `team_lead`, `user`.
  - edit others: `super_admin` only by default.

- งานระดมทีม:
  - create/manage: `super_admin`, `admin` for MVP.
  - read: all authenticated roles for awareness/involved-team visibility as backend returns.

---

## 6. Testing and verification requirements

TDD order for implementation cards:

1. Add/update role-policy tests first:
   - `user` can manage own-team monthly-plan file when unlocked.
   - `user` cannot manage another team's monthly-plan file.
   - `team_lead` own-team behavior remains allowed.
   - `viewer` cannot upload/download.
2. Add type/view-model tests for calendar projection:
   - multi-day item expands across inclusive dates.
   - no-time item renders on the date.
   - badge counts group by item type.
3. Add component tests if a test framework is introduced; otherwise keep pure helper tests with `tsx` like current role-policy tests.
4. Run frontend gates after implementation:
   - `npx --yes tsx src/lib/auth/role-policy.test.ts`
   - feature helper tests, if added
   - `npx tsc --noEmit`
   - `npm run build`
   - `git diff --check`

Docs-only verification for this card:

- `git diff --check` in `/Users/sakdithat/Desktop/myproject/backend-hotline`.

---

## 7. Open implementation notes

- Do not duplicate the monthly-plan yearly UX inside calendar; link or deep-link back to `/monthly-plan`.
- Keep planning features out of `/admin` except admin-only settings/operations.
- Preserve mobile-first layout: calendar day drawer and create forms must work on 320px+ screens.
- Use existing green/yellow/white/gray design tokens; red only for errors/destructive warnings.
- Avoid explanatory/test-like UI copy in production screens; labels should be operational and Thai-first.
