# Hotline UI Polish Handoff — 2026-05-10

บันทึกเมื่อ: 2026-05-10 23:45:21 +07

## Goal
ปรับ UI โปรเจกต์ Hotline ให้เข้าธีมเดียวกับหน้า planning ล่าสุด โดยยึด mobile-first เป็นหลัก: clean, emerald/amber, soft glass, มีมิติ, อ่านง่าย และเหมาะกับงานภาคสนาม

## Key Decisions & Reasoning
- ใช้ mobile-first เป็น release gate เพราะผู้ใช้ย้ำว่าระบบนี้ต้องใช้งานบนมือถือเป็นหลัก ทั้งตอนทำ ตอน QA และตอนรีวิว
- ใช้ shared UI primitives (`PageShell`, `PageHero`, `KpiCard`) เพื่อให้หน้าต่าง ๆ มี visual language เดียวกัน แทนการปรับ style แยกหน้าแบบกระจัดกระจาย
- คง popup/detail drawer ของหน้า planning ไว้ ไม่รื้อ เพราะผู้ใช้ระบุว่าชอบส่วนนี้อยู่แล้ว และความเสี่ยง regression สูงกว่า benefit
- ยกระดับ tap target สำคัญเป็นประมาณ 44px (`min-h-11`, `h-11 w-11`) เพื่อให้ใช้งานบน touch device ได้ดีขึ้น
- ใช้ QA gate แบบต่อเนื่อง: TypeScript → ESLint → production build → browser/mobile QA → screenshot/report
- แยกงาน mobile contacts card layout เป็น task ต่อ เพราะ contacts table หลายคอลัมน์ยังมีความเสี่ยงบนมือถือ แม้ข้อมูลโหลดได้และ interaction หลักใช้งานได้

## What Changed
Frontend repo: `/Users/sakdithat/Desktop/myproject/hotlines3`

- `src/components/ui/page-shell.tsx` — เพิ่ม shared primitives `PageShell`, `PageHero`, `KpiCard`
- `src/app/(auth)/login/page.tsx` — ปรับหน้า login เป็น mobile-first hero/card shell และเพิ่ม tap target ปุ่ม password visibility
- `src/features/task-daily/components/task-daily-form.tsx` — ลด spacing/padding และปรับ hero/form ให้เหมาะมือถือ
- `src/features/task-daily/components/plan-prefill-picker.tsx` — ทำ toggle/action ให้แตะง่ายขึ้น
- `src/features/task-daily/components/location-picker.tsx` — เพิ่ม mobile-friendly tap target ให้ปุ่มตำแหน่งปัจจุบัน
- `src/app/(main)/planning/page.tsx` — ปรับปุ่ม/แท็บหลักให้ mobile-friendly โดยไม่แตะ drawer/popup
- `src/features/planning-calendar/components/CalendarMonthSelector.tsx` — เพิ่มขนาดปุ่มเดือนก่อน/ถัดไป
- `src/features/planning-calendar/components/CalendarFilterBar.tsx` — เพิ่มขนาด filter pills
- `src/components/header.tsx` — เพิ่มขนาด logout button สำหรับ touch device
- `src/app/(main)/contacts/page.tsx` — เพิ่ม touch-friendly search input และมีการปรับ phone/tel link จากรอบก่อน
- `src/app/(main)/monthly-plan/page.tsx` — เริ่ม unify เป็น shared shell/hero และปรับ action buttons
- `src/app/(main)/admin/page.tsx` — เริ่ม unify admin landing ด้วย shared shell/hero; มี patch ล่าสุดแก้ closing wrapper เป็น `</PageShell>`

Backend/Obsidian vault repo: `/Users/sakdithat/Desktop/myproject/backend-hotline`

- มีรายงานก่อนหน้า:
  - `plan/23-hotline-self-qa-mobile-calendar-contacts-2026-05-10.md`
  - `plan/24-hotline-planning-ui-polish-2026-05-10.md`
- บันทึก handoff นี้เพิ่มเป็น `plan/25-hotline-ui-polish-handoff-2026-05-10.md`

## Verification Status
ตรวจล่าสุดหลังตั้งค่า Hermes max turns เป็น 1000:

- `npx tsc --noEmit` — ผ่าน
- `npm run lint -- --quiet` — ผ่าน
- `npm run build` — ผ่าน

Build warning ที่ยังพบ:

```text
Failed to load env from .env Error: Unknown system error -11: Unknown system error -11, read
```

หมายเหตุ: warning นี้ไม่ทำให้ build fail; Next compile และ generate routes สำเร็จครบ

## Browser / Mobile QA Evidence
มี screenshot จาก QA mobile รอบก่อนอยู่ที่:

- `/tmp/hotline-mobile-shots/01-login.png`
- `/tmp/hotline-mobile-shots/02-daily.png`
- `/tmp/hotline-mobile-shots/02-planning.png`
- `/tmp/hotline-mobile-shots/02-contacts.png`
- `/tmp/hotline-mobile-shots/02-monthly-plan.png`

ผล mobile QA รอบที่บันทึกไว้:

- `/login` — ไม่มี horizontal overflow, tap targets ผ่าน
- `/` daily report — ไม่มี horizontal overflow, tap targets ผ่าน
- `/planning` — ไม่มี horizontal overflow, location summary และ `+N` แสดง, drawer ยังทำงาน
- `/contacts` — ข้อมูล mock contacts โหลดได้, search/filter ใช้งานได้; ยังควรทำ mobile card layout ต่อ
- `/monthly-plan` — ไม่มี horizontal overflow ใน QA รอบก่อน

## Current State
- ไม่มี blocker ที่ทำให้ไปต่อไม่ได้
- TypeScript/lint/build ผ่านแล้ว
- Hermes profile `scc` ถูกตั้ง `agent.max_turns: 1000` และ restart gateway แล้ว
- งาน UI polish ยังไม่จบ production-complete ทุกหน้า เพราะยังเหลือเก็บ admin/core pages และ QA browser final รอบเต็ม
- Git working tree มีไฟล์แก้เยอะมาก ทั้ง tracked และ untracked; session ใหม่ควรเริ่มจากตรวจ `git status` และแยก diff เป็นชุดก่อนทำต่อ

## Remaining Work / Next Steps
1. ตรวจ `git status` และ diff ใน `hotlines3` เพื่อจัดกลุ่มไฟล์งานจริง vs ไฟล์ชั่วคราว (`dogfood-output/`, `scripts/`, screenshot temp ฯลฯ)
2. ทำให้ admin landing ที่เพิ่ง patch ล่าสุด compile/QA ผ่านซ้ำใน browser เพราะ patch closing wrapper เพิ่งเกิดก่อน session ถูก interrupt
3. ปรับ/ตรวจหน้าที่ยังค้าง:
   - `/list`
   - `/admin/dashboard`
   - `/admin/monthly-plan`
   - master data pages ใต้ `/admin/*`
   - upload dialogs/drawers/modals บนมือถือ
4. รัน verify gate อีกครั้ง:
   - `npx tsc --noEmit`
   - `npm run lint -- --quiet`
   - `npm run build`
5. ทำ browser QA mobile viewport รอบ final พร้อม screenshot ใหม่หลายหน้า
6. อัปเดตรายงาน plan/Obsidian พร้อม screenshot paths และสรุป final ให้ผู้ใช้

## Notes for Next Session
- เริ่ม session ใหม่ได้เลยหลังบันทึกนี้ เพราะ context ปัจจุบันยาวมากและงานถัดไปควรเริ่มจาก checkpoint ที่สะอาด
- ให้โหลด context จากไฟล์นี้และรายงาน `plan/23`, `plan/24` ก่อนทำต่อ
- ย้ำ mobile-first เป็น criterion หลักทุกครั้ง ไม่ใช่แค่ desktop polish
- ห้ามบันทึกหรือเปิดเผย credential จริง; ใช้ `[REDACTED]` ในรายงานเท่านั้น

## Update — 2026-05-11 00:35:49 +0700

### Scope continued from this handoff
- Started from `git status`/diff in both repos.
- Removed local generated `.next-bad-*` build cache from `hotlines3` before continuing because it was untracked/generated and made git indexing noisy.
- Kept the remaining core/admin UI polish focused on `/admin`, `/admin/dashboard`, `/admin/monthly-plan`, `/list`, `/contacts`, `/planning`, and `/monthly-plan` with mobile viewport `390x844`.

### Additional changes
Frontend repo: `/Users/sakdithat/Desktop/myproject/hotlines3`

- `src/features/monthly-plan/utils.ts` — hardened `formatPeriodLabel` and `formatPeriodLabelFull` against missing/partial period data so UI no longer renders `undefined` or `NaN` in monthly plan copy.
- `src/types/monthly-plan.test.ts` — added assertions for normal Thai period labels and missing-period fallback behavior.
- `src/components/pages/task-list-client.tsx` — increased `/list` filter select triggers and search/download buttons to `h-12/min-h-12` so mobile tap-target QA passes.
- `src/features/monthly-plan/components/PlanFileRow.tsx` — increased icon-only row action buttons to `min-h-11 min-w-11`.
- `src/features/monthly-plan/components/MasterPlanBanner.tsx` — increased mobile download/delete action targets to at least 44px.

QA helper only:
- `/tmp/hotline_mock_api.mjs` — adjusted mock response shapes for contacts and monthly-plan routes so browser QA can render `/contacts`, `/admin/monthly-plan`, and `/monthly-plan` without depending on the real backend.

### Verification
All gates passed in `/Users/sakdithat/Desktop/myproject/hotlines3`:

- `npx tsx src/types/monthly-plan.test.ts` — passed
- `npx tsc --noEmit` — passed
- `npm run lint -- --quiet` — passed
- `NEXT_TELEMETRY_DISABLED=1 npm run build` — passed

Build still prints the known warning below, but exit code is 0 and all routes compile/generate successfully:

```text
Failed to load env from .env Error: Unknown system error -11: Unknown system error -11, read
```

### Mobile browser QA evidence
QA ran via headless Chrome CDP with viewport `390x844`, mock admin user `900003`, and screenshots saved in:

- `/tmp/hotline-mobile-shots-final/01-admin.png`
- `/tmp/hotline-mobile-shots-final/02-dashboard.png`
- `/tmp/hotline-mobile-shots-final/03-admin-monthly-plan.png`
- `/tmp/hotline-mobile-shots-final/04-list.png`
- `/tmp/hotline-mobile-shots-final/05-contacts.png`
- `/tmp/hotline-mobile-shots-final/06-planning.png`
- `/tmp/hotline-mobile-shots-final/07-monthly-plan.png`
- JSON report: `/tmp/hotline-mobile-shots-final/qa-report.json`

QA results:
- `/admin` — no horizontal overflow, no small tap targets, admin landing content renders.
- `/admin/dashboard` — no horizontal overflow, no small tap targets, dashboard filters/KPI render.
- `/admin/monthly-plan` — no horizontal overflow, no small tap targets, period text now renders `แผนงานพฤษภาคม 2569` instead of `undefined/NaN`.
- `/list` — no horizontal overflow, no small tap targets after increasing filter/action targets.
- `/contacts` — no horizontal overflow, no small tap targets, contact list renders with mock data instead of error page.
- `/planning` — no horizontal overflow, no small tap targets.
- `/monthly-plan` — no horizontal overflow, no small tap targets after action target polish.

### Remaining notes
- `hotlines3` still has a broad pre-existing working tree with many tracked/untracked files. Review/commit should be split carefully by feature area.
- `backend-hotline` still has pre-existing router/masterdata/migration/plan changes; this update only appended to this report.
