# Hotline Redesign UI Component Audit — HRD-UX0B

Date: 2026-05-19
Task: HRD-UX0B UX/UI component audit and implementation handoff
Frontend repo: `/Users/sakdithat/Desktop/myproject/hotline/hotlines3`
Source of truth: Requirement A/B/C/D in `/Users/sakdithat/Desktop/myproject/wiki/queries/`

## Executive decision

Frontend foundation work should happen before page-by-page implementation. Current UI already has useful primitives and some redesign prep, but the visible app shell/pages still mix old HotlineS3 green/glassmorphism, legacy `/list`, legacy `admin` role semantics, and page-specific card/status code.

Important override: Requirement D is authoritative for this redesign. Ignore stale green/glassmorphism rules in frontend `CLAUDE.md` and legacy utilities. New surfaces must use official/deep blue primary, light neutral background, blue team-planning source, teal monthly-plan source, semantic status colors, Thai-first copy, and mobile card/list/sheet patterns.

## Requirement fit snapshot

| Area | Current implementation | Requirement gap | Foundation target |
|---|---|---|---|
| App shell/nav | `src/components/header.tsx`, `src/components/navbar.tsx`, `src/config/navigation.tsx`; fixed header + mobile bottom nav; still emerald/glass | Missing `/work-report`; still has `/list`; Admin visible to `admin`; labels not final; old brand gradients | Replace nav config with Requirement B routes and labels; only `super_admin` sees Admin; default post-login `/planning`; deep-blue active/current state |
| Page shell/header | `src/components/ui/page-shell.tsx` has `PageShell`, `PageHero`, `KpiCard`; many pages also define local hero/header blocks | `PageHero` is still decorative emerald gradient; Requirement D asks practical header/filter, not hero content | Create clean `AppPageShell` + `PageHeader` + `FilterBar` and migrate pages to it; avoid decorative hero gradients |
| Tokens | `src/app/globals.css` already has Requirement D CSS variables and clean utility classes; `src/shared/design-tokens.ts` exists | Many components still hard-code `emerald`, `amber`, `sky`, `card-glass`, gradients; token helpers are not wired broadly | Standardize token exports and component variants; new code uses tokens/helpers, not direct one-off Tailwind color strings |
| Badges | `src/components/ui/badge.tsx`; `src/shared/badge-utils.ts`; some page-local `statusBadgeClass()` | Status/source/role badges duplicated; source colors not consistently D2 blue/teal; labels differ from Requirement C | Add shared `StatusBadge`, `SourceBadge`, `RoleBadge`, `PermissionBadge`; map all statuses/sources to D labels/colors |
| Cards | `src/components/ui/card.tsx`; many page-local cards (`TeamPlanCard`, `LargeWorkCard`, contact rows, monthly file cards) | Cards contain useful fields, but styling/actions are inconsistent; source/status/permission action patterns vary | Build reusable `OperationalCard` pattern for work/report/contact/file with action slots and metadata rows |
| Tables -> mobile cards | Admin pages and contact/list pages still use tables/grids; some are responsive, some use horizontal scroll | Requirement C/D says mobile must not squeeze desktop tables | Add `ResponsiveDataView` or page-level convention: desktop table, mobile cards/list; avoid `overflow-x` as mobile solution except admin-dense fallback |
| Drawer/sheet/detail | `src/components/ui/drawer.tsx` (Vaul) and `dialog.tsx` exist; Planning has `DayDetailDrawer`; Contact uses Dialog | Desktop side drawer vs mobile bottom/full sheet is not a shared pattern; dialogs are used for forms/details | Create `DetailSurface` wrapper: desktop drawer, mobile sheet/dialog; confirm dialogs for risky actions only |
| States | `src/components/ui/skeletons.tsx`, ad-hoc loading/error/empty blocks across pages | State copy/classes are duplicated; some errors lack retry; no standard read-only/no-permission state | Add `PageState`, `SectionState`, `EmptyState`, `ErrorState`, `LoadingSkeleton`, `NoPermissionState` with Thai copy hooks |
| Permission UI | Role policy in `src/lib/auth/role-policy.ts`; actions hidden by page logic | Policy still includes legacy `admin`; no capability-driven model; no consistent disabled `ไม่มีสิทธิ์` pattern | Frontend needs capability-aware helpers from backend contract; meanwhile isolate permission UI helpers and document where actions hide vs disable |

## Current reusable assets to keep

- `src/components/ui/button.tsx`, `input.tsx`, `textarea.tsx`, `select.tsx`, `native-select.tsx`, `dialog.tsx`, `drawer.tsx`, `card.tsx`, `badge.tsx`, `skeletons.tsx`, `sonner.tsx` are the primitive base.
- `src/app/globals.css` already contains useful Requirement D clean utilities:
  - `.card-clean`, `.card-clean-hover`, `.section-clean`
  - `.badge-success`, `.badge-warning`, `.badge-info`, `.badge-error`, `.badge-neutral`
  - `.badge-source-team`, `.badge-source-monthly`
  - `.page-header`, `.filter-bar`, `.pb-safe`
- `src/shared/design-tokens.ts`, `src/shared/badge-utils.ts`, and `src/shared/badges.tsx` appear to be the right home for shared color/status/source/role badge logic. Frontend foundation should verify/export these from `src/shared/index.ts` and use them across pages.
- Planning calendar components already separate useful pieces under `src/features/planning-calendar/components/`: `CalendarGrid`, `CalendarFilterBar`, `CalendarMonthSelector`, `DayDetailDrawer`, `PlanningItemTypeBadge`.
- Monthly plan has reusable file-row/banner/upload pieces under `src/features/monthly-plan/components/`.
- Contact page already has search/filter/detail/edit concepts; it needs component extraction and Requirement D styling/permissions, not a full rewrite first.

## Foundation components to implement first (HRD-F0)

1. App shell and navigation
   - Files: `src/config/navigation.tsx`, `src/components/header.tsx`, `src/components/navbar.tsx`, `src/app/(main)/layout.tsx`, `src/lib/auth/auth-guard.tsx`, `src/lib/auth/role-policy.ts`.
   - Required routes/labels:
     - `/planning` — `ระบบวางแผนงาน`; default after login.
     - `/monthly-plan` — `แผนประจำเดือน`.
     - `/daily-report` — `บันทึกงาน`.
     - `/work-report` — `รายงานการปฏิบัติงาน`.
     - `/contacts` — `สมุดโทรศัพท์`.
     - `/admin` — `จัดการระบบ`, `super_admin` only.
   - Remove `/list` from main nav. Do not resurrect Dashboard nav. If legacy route remains temporarily, it must not be the redesign nav target.
   - Active nav should be clear with deep-blue/neutral styling, not emerald glass.

2. Page shell/header/filter foundation
   - Files: `src/components/ui/page-shell.tsx` or new `src/components/layout/*`.
   - Replace decorative `PageHero` usage with practical `PageHeader`:
     - title, short description, scoped context/team lock, date/month controls, active filters, primary action if permitted.
   - Filters stay near data on desktop; mobile can collapse into chips/sheet but active filters remain visible.
   - Use light neutral background and clean cards/sections.

3. Semantic badge/source/status system
   - Files: `src/shared/design-tokens.ts`, `src/shared/badge-utils.ts`, `src/shared/badges.tsx`, maybe `src/components/ui/badge.tsx`.
   - Required variants:
     - status: waiting/draft = amber, scheduled/in_progress = blue, completed = green, cancelled/destructive = red, read-only/neutral = slate/gray.
     - source: team-planning = blue, monthly-plan = teal.
     - permission: muted gray + `ไม่มีสิทธิ์` when disabled action is intentionally shown.
   - Color is never the only signal; badges include text labels and optional icons.

4. Operational cards and responsive data view
   - Files: likely new `src/components/patterns/operational-card.tsx`, `responsive-data-view.tsx`, then page usage.
   - Card minimum fields by Requirement D: title/name, date/time, location, team/owner, source badge, status badge, permission-gated actions.
   - Mobile-specific priorities:
     - Planning card: detail/edit/create-report/move-to-calendar depending permission.
     - Contact card: `โทร` is easiest and most visible.
     - File card: preview/download/upload rules follow role/capability.
     - Report card: detail first; edit/delete only when permitted.

5. Detail drawer/sheet and form surface
   - Files: `src/components/ui/drawer.tsx`, `src/components/ui/dialog.tsx`, likely new `src/components/patterns/detail-surface.tsx`.
   - Desktop details should use side drawer where practical.
   - Mobile details/forms should use bottom sheet or full-screen sheet.
   - Avoid nested modals.
   - Risky deletes/permission/admin changes use a named confirmation dialog, not only `window.confirm` long term.

6. State patterns
   - Files: `src/components/ui/skeletons.tsx` or new `src/components/patterns/page-state.tsx`.
   - Standardize:
     - loading skeleton that preserves page structure
     - empty with Requirement C Thai copy and permission-aware CTA
     - retryable error with `ลองใหม่`
     - success toast copy from Requirement C
     - read-only/no-permission state
   - Existing ad-hoc errors like `ไม่สามารถโหลดข้อมูลได้` should become page-specific confirmed copy, e.g. `โหลดข้อมูลงานไม่สำเร็จ`.

## Page-level implementation notes

### `/planning`

Current file: `src/app/(main)/planning/page.tsx` plus `src/features/planning-calendar/components/*` and `src/features/large-work/components/*`.

Findings:
- Has Calendar tab and rich Planning work already, but current tabs are `ปฏิทิน | แผนทีม | งานระดมทีม | คิวงานฉัน` instead of Requirement C `Calendar | Board` structure.
- Uses emerald/dark gradient hero and `card-glass`; must move to practical header/filter and light neutral shell.
- Calendar item type/source includes `team_plan`, `monthly_plan`, `large_work`; Requirement C focuses source badges `งานแผนของทีม` and `งานจาก monthly plan`. Treat `large_work` as existing extension and do not let it confuse source-color rules.
- Existing `DayDetailDrawer` is a good detail pattern but should align with shared `DetailSurface` and D2 badges.

Handoff:
- First align header/month/source/status filters and D2 color badges.
- Then reshape tabs to Requirement C: `Calendar` default + `Board` for unscheduled work. Existing large-work/worker todo UX may become sub-sections or separate extension after requirement confirmation.
- Add disabled `เพิ่มงาน` with `ไม่มีสิทธิ์` for normal user where Requirement C says to show it; hide create/edit for `viewer`.

### `/monthly-plan`

Current file: `src/app/(main)/monthly-plan/page.tsx`, components in `src/features/monthly-plan/components/*`.

Findings:
- Current page is yearly/monthly card layout and starts with `แผนงานประจำปี`; Requirement C wants selected month page: header/month selector, approved/master file first, then tabs/sections.
- Uses `PageHero`, emerald cards, `admin`/`isPrivilegedAdmin` permission model.
- File components already support banners/rows/upload and can be reused.

Handoff:
- Reframe page around selected month, not a 12-month overview first.
- Top content must be `ไฟล์อนุมัติประจำเดือน` always visible.
- Then tabs/sections: `แผนทีมของฉัน`, `แผนทีมอื่น / ภาพรวม`, `อัปโหลดไฟล์อนุมัติ` gated to `super_admin` or `can_upload_approved_monthly_plan`.
- Viewer can preview PDF/image only; no download. Normal user can download approved/master file but not team-other submissions.

### `/daily-report`

Current file: `src/app/(main)/daily-report/page.tsx` delegates to `src/features/task-daily/components/task-daily-form.tsx`.

Findings:
- Existing flow is the old daily task form with job types/details/feeders/teams.
- Requirement C wants a page with date selector, planned-work selector/prefill, one reusable form callable from Calendar card, Board card, and Daily Report page.

Handoff:
- Extract a shared `DailyReportFormSurface` that can be opened from Planning/Board cards and rendered on `/daily-report`.
- Add planned work prefill source fields and make clear that actual work details are still user-entered.
- Viewer read-only: no create/edit/delete actions.

### `/work-report`

Current route: missing. Legacy equivalent is `/list` via `src/app/(main)/list/page.tsx` and `src/components/pages/task-list-client.tsx`.

Findings:
- Navigation still points to `/list` label `รายการรายงาน`.
- Requirement B/C explicitly says do not resurrect old `/list`; use `/work-report` and Thai UI name `รายงานการปฏิบัติงาน`.

Handoff:
- Create `/work-report` route and migrate/use relevant list/report code there.
- Main layout: header/filter, summary cards, desktop table + detail drawer, mobile cards/list.
- Viewer has no download/export/write action.
- Remove `/list` from nav; optional redirect from `/list` to `/work-report` only if needed for compatibility.

### `/contacts`

Current file: `src/app/(main)/contacts/page.tsx`.

Findings:
- Has search/filter/edit dialog and table/card concepts; styling still emerald/glass and local logic.
- Requirement C wants search as primary control, mobile contact cards, detail drawer/bottom sheet, `โทร` easiest, viewer can call/copy/view but no add/edit/delete.

Handoff:
- Extract contact card/detail pattern into reusable operational card + detail surface.
- Keep call/copy quick actions visible, especially mobile.
- Add viewer read-only mode and no add/edit/delete actions.

### `/admin`

Current files: `src/app/(main)/admin/page.tsx`, admin subroutes, `src/lib/auth/admin-guard.tsx`, `role-policy.ts`.

Findings:
- Current admin console includes old master data pages and some monthly-plan/task-daily admin access for legacy `admin` role.
- Requirement C says `/admin` is `super_admin` only for this redesign. Admin tabs: users, teams, capability, audit. Requirement A says no audit log needed for current feature, but Requirement C/D still mention Audit tab/read-only UI; flag backend contract alignment before building full audit data.

Handoff:
- Restrict `/admin` nav and route guard to `super_admin`.
- New admin shell: summary cards + tabs `ผู้ใช้ | ทีม | สิทธิ์/Capability | Audit`.
- Risky actions require confirmation with Thai target/consequence copy.
- Do not show edit/delete on audit rows.

## Known risks before page work

- Role model mismatch: frontend still has legacy `admin`; Requirement A says no `admin_monthlyplan` and Admin menu is `super_admin` only. Capability model must come from backend before fully correct UI gating.
- Source/status mismatch: existing planning source includes `large_work` and status labels like `planned`; Requirement C/D labels are `รอวางแผน`, `กำหนดวันแล้ว`, `กำลังทำ`, `เสร็จแล้ว`, `ยกเลิก` plus monthly plan statuses.
- Existing `CLAUDE.md` instructs green/glassmorphism and forbids blue/teal; this is stale for redesign. Requirement D overrides it.
- Existing `/list` and dashboard/admin-dashboard files remain in repo. Workers must not use them as redesign IA source of truth.
- Some foundation files appear already started (`src/shared/*`, clean globals utilities) but are uncommitted in the frontend workspace. HRD-F0 should verify provenance and avoid overwriting other worker changes.
- Current page components are large and page-local. Refactor in small steps: foundation components first, then migrate one page at a time with lint/typecheck.

## Browser/mobile QA notes

- Test desktop widths around 1280px and 1024px for header/filter density and table-vs-card switches.
- Test mobile widths 390px and 360px: bottom nav safe area, Thai label wrapping, touch targets >= 44px, no horizontal table squeeze.
- Verify current-page nav state is obvious on desktop and mobile.
- Verify all active filters remain visible after opening/closing mobile filter sheets.
- Verify drawer/sheet behavior: no nested modals, Escape/backdrop close where safe, destructive confirmations do not close accidentally.
- Verify `viewer`: no write/download/export actions, except allowed preview/call/copy behavior.
- Verify normal user without capability sees `ไม่มีสิทธิ์` only where Requirement C expects disabled education; otherwise hidden actions.
- Verify source/status badges include text labels and are not color-only.
- Verify no Dashboard nav/route is introduced and login/default route lands on `/planning`.
- Verify no production UI credits/demo/MVP copy appears.

## Suggested HRD-F0 acceptance checklist

- [ ] Requirement D visual override documented in code comments or plan notes where old green/glass rules are likely to mislead.
- [ ] `src/config/navigation.tsx` matches Requirement B and no longer exposes `/list` or Dashboard as redesign nav.
- [ ] `src/components/header.tsx` and `src/components/navbar.tsx` use deep-blue/neutral active states and `super_admin`-only Admin visibility.
- [ ] Shared page header/filter/state components exist and are used by at least the first migrated page.
- [ ] Shared badges exist for status/source/role/permission and cover Requirement D labels/colors.
- [ ] Mobile card/list/sheet pattern exists for at least one table-heavy page before broad migration.
- [ ] Permission UI distinguishes hidden viewer actions from disabled `ไม่มีสิทธิ์` user actions.
- [ ] Lint/typecheck/build pass before downstream page tasks start.
