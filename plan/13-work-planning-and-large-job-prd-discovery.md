# Hotline PRD Discovery — Team Plan, Monthly Plan, and งานระดมทีม

Date: 2026-05-09
Status: discovery only; do not implement until current performance/RBAC/monthly-plan stabilization is complete.

## Source

User clarified the next product direction after the performance-first replan:

- The system should not focus only on daily report/data recording like the old system.
- The system should focus on planning work so each team knows its own upcoming work and can prepare.
- The future multi-team large-work feature must be called **งานระดมทีม**.
- **งานระดมทีม is out of scope for now** and should not be discussed or implemented in the current planning pass.
- Current planning focus is:
  1. `monthly plan` for work outside the team’s responsible area that needs approval.
  2. `team plan` for work inside the team’s own responsible area that does not need approval.
  3. Calendar-style visibility so users can see where they are going on each day.
  4. Team/user contact directory so people can find name, position, and phone number of other teams.

## Product Direction

Hotline should evolve from a daily reporting system into an operations planning system with daily reporting as a downstream/supporting workflow.

The desired experience is similar to Google Calendar:

- Monthly calendar view shows which dates have planned work.
- Clicking a date shows the jobs planned for that day.
- Time is optional; it is acceptable to know only that a team has work on that day, at which point/location/electric area, and what rough work type.
- Users should be able to plan work and personal life more easily because upcoming team work is visible.

## Domain Terms

### monthly plan

A planned work item that goes **outside the team’s own responsible area** and therefore must be submitted as a monthly-plan document workflow.

Current understanding:

- Must be submitted because the team leaves its own responsible area.
- It remains primarily a **document submission** system.
- Users enter only enough details to describe where and when work will happen: location, work date, and end date/range.
- There is **no in-system approval flow** for monthly plan.
- Admin takes the submitted document outside the system for approval, then uploads the approved document back after approval is received.
- Should be shown on the same planning calendar.
- Work date and location from monthly plan should be reusable/prefilled into daily work reporting.

### team plan

A planned work item inside the team’s own responsible area.

Current understanding:

- Does not require approval.
- `user` and `team_lead` can both add team-plan work together.
- Anyone in the team can add a team plan.
- The user who created a team plan can edit their own item.
- Deleting/removing is restricted: only `team_lead` can delete.
- Intended for rough planning, not heavy reporting.
- Should sync or be reusable when entering the daily report/task data later.

### งานระดมทีม

The future multi-team large-work planning feature.

Current instruction:

- Remember this exact term.
- Do not discuss/plan/implement it yet.
- It will be a larger separate feature later.

### team/user contact directory

A directory of users/teams so staff can contact each other more easily.

Required visible fields from user requirement:

- Name
- Position
- Phone number
- Team information

## Current User Flow Draft

### Plan Type Selection Flow

1. User decides whether the work is inside or outside the team’s own responsible area.
2. If the work is inside the team’s own area, use `team plan`.
3. If the work is outside the team’s own area, use `monthly plan` and submit the required document/details.

### Team Plan Creation Flow

1. A team member or `team_lead` opens the planning calendar.
2. They add a planned work item for their own team/responsible area.
3. They enter rough planning details:
   - Work date
   - Location / where to work
   - End date or until when, if the work spans multiple days
   - Rough work description / work type
   - Electric area / PEA / feeder / station if applicable
   - Optional notes
   - Optional time; time is not required
4. The planned work appears on the monthly calendar view.
5. Clicking that day shows the planned work items for that day.
6. The creator can edit their own team-plan item.
7. If the plan needs removal/deletion, only `team_lead` can delete it.
8. When the actual work day arrives, daily work reporting should be able to sync/reuse this plan data instead of forcing users to re-enter everything.

### Monthly Plan Flow

1. A team plans work outside its own responsible area.
2. Because it leaves the responsible area, the user submits monthly-plan document/details.
3. Monthly-plan details only need to capture where the work is, the work date, and the end date/range.
4. There is no in-system approval. Admin processes the approval externally, then uploads the approved document back into the system.
5. The monthly plan should also appear on the same planning calendar.
6. On the work date, the actual daily report should sync/reuse the plan information where possible.

### Calendar Viewing Flow

1. User opens a monthly calendar view.
2. Each day shows whether there is planned work.
3. User clicks a date.
4. System shows all work planned for that date:
   - team plan items
   - monthly plan items
5. Time is optional; if no time is provided, the item still appears on that date.
6. The key information is: what work, where, which electric area/location, and which team.

### Team Contact Directory Flow

1. User opens team/user directory.
2. User can search or browse other teams/users.
3. User can see basic contact information:
   - name
   - position
   - phone number
   - team
4. Purpose: enable easier coordination/contact between teams.

## Role/Permission Draft

| Capability | super_admin | admin | team_lead | user |
|---|---:|---:|---:|---:|
| View planning calendar | yes | yes | yes | yes |
| Add team plan for own team | yes | yes / as configured | yes | yes |
| Edit own-created team plan | yes | yes | yes | yes |
| Delete team plan | yes | yes / as configured | own team only | no |
| Create monthly plan / outside-area request | yes | yes | own team only | no or needs confirmation |
| View monthly plan | yes | yes | yes | yes |
| Sync plan into daily report | yes | yes | own team | own assigned/own team |
| View team contact directory | yes | yes | yes | yes |
| Manage own personal/contact info | yes | yes | yes | yes |
| Manage other users contact info | yes | maybe admin scope needs confirmation | no | no |
| งานระดมทีม | future only | future only | future only | future only |

## Data Fields Draft

### Plan item common fields

- id
- plan type: `team_plan` or `monthly_plan`
- team id
- created by user id
- date / start date
- end date, optional
- time, optional
- location text
- station / feeder / PEA / operation center if applicable
- rough work title/type
- rough work details/notes
- status
- source link to monthly plan file/request if applicable
- approved document link if admin uploaded it back after external approval
- daily report/task link if synced later
- created/updated timestamps

### Contact directory fields

- user id
- display name
- position/title
- phone number
- team id/name
- active status

## Boundary With Current Work

Current active board `hotline-performance-rbac-2026` remains focused on:

1. Performance fixes.
2. RBAC split.
3. Monthly-plan yearly view and corrected upload/download permissions.

This discovery should not interrupt HP1 while performance work is running. It should feed a later PRD/implementation board after current stabilization.

## Confirmed Decisions — 2026-05-09

- Inside own responsible area → use `team plan`.
- Outside own responsible area → use `monthly plan`.
- `team plan` requires no approval.
- `monthly plan` is primarily document submission; it has no in-system approval. Admin handles approval outside the system and uploads the approved document back.
- `user` who created a `team plan` can edit their own item.
- Only `team_lead` can delete team-plan items.
- Daily report should be able to reuse/sync work date and location from both `monthly plan` and `team plan`.
- Everyone can edit all of their own personal/contact information because it is their own data.

## Open Questions For User

### Team Plan

1. `team plan` ใช้ได้เฉพาะในพื้นที่รับผิดชอบตัวเอง — ระบบรู้พื้นที่รับผิดชอบจากอะไร?
   - team → PEA/operation center
   - feeder/station mapping
   - manual text only
   - other rule

2. `team_lead` ลบ team plan ได้ทุกงานของทีมตัวเองใช่ไหม?

3. team plan ต้องมีสถานะไหม เช่น draft/published/cancelled/completed หรือแค่สร้างแล้วแสดงใน calendar พอ?

4. ถ้า plan เป็นหลายวัน ให้แสดงใน calendar ทุกวันที่อยู่ในช่วงนั้นใช่ไหม?

### Monthly Plan

1. monthly plan ควรแปลงเป็น calendar item อัตโนมัติหลังส่งเอกสารใช่ไหม หรือหลัง admin upload เอกสารอนุมัติกลับมาเท่านั้น?

2. Admin upload approved document กลับมาได้กี่ไฟล์/เวอร์ชัน และต้องแทนไฟล์เดิมหรือเก็บประวัติ?

### Daily Report Sync

1. เมื่อถึงวันทำงาน ต้องให้ daily report ดึงข้อมูลจาก `team plan`/`monthly plan` มา prefill ใช่ไหม?

2. ถ้าทำงานจริงไม่ตรงแผน ให้แก้ใน daily report ได้ไหมโดยไม่แก้ plan เดิม?

3. ต้องเก็บความต่างระหว่าง “แผน” กับ “ทำจริง” ไหม?

### Contact Directory

1. user ทุกคนเห็นเบอร์โทรทุกคนได้ไหม หรือจำกัดเฉพาะภายในองค์กร/ทีม?

2. ต้องมี search/filter ตามทีม/ตำแหน่ง/พื้นที่ไหม?

## Candidate Future Kanban Graph

Do not start this graph until user confirms discovery answers.

1. HN0 — PRD discovery for team plan, monthly plan calendar, daily report sync, contact directory.
2. HN1 — Data model design: plan item, plan source, calendar projection, contact fields.
3. HN2 — Backend API and RBAC policy tests for team plan CRUD and monthly-plan calendar projection.
4. HN3 — Frontend calendar UX prototype: monthly view + day detail drawer.
5. HN4 — Daily report sync/prefill design and implementation.
6. HN5 — Team/user contact directory backend/frontend.
7. HN6 — QA, smoke tests, migration safety, release docs.

## Initial Recommendation

Capture requirements now while context is fresh, but keep implementation gated behind current performance/RBAC/monthly-plan stabilization. The planning-calendar direction is important enough that it should influence the upcoming monthly-plan yearly UX, but it should not cause uncontrolled scope creep during HP1.
