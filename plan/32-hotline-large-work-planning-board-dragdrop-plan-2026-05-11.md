# Hotline Large Work Planning Board Drag/Drop Plan

> **For Hermes:** Use subagent-driven-development + test-driven-development. Build UX as a planning board first, with drag/drop as progressive enhancement and plain-form entry as a full fallback.

**Goal:** Turn `งานระดมทีม` from a small “add task rows” dialog into a full work-planning board where a team lead can see all participating teams, create job cards, drag cards to teams, and also enter everything manually without drag/drop.

**Architecture:** Keep backend task persistence as `large_work_tasks`, but introduce a richer frontend planning model: draft task cards, team lanes, unassigned pool, and validation before save. Drag/drop only changes draft assignment/order in the browser; saving still calls the existing `POST /v1/large-works/:id/tasks` endpoint with a plain tasks array.

**Tech Stack:** Next.js/React 19, Tailwind, TanStack Query, `@dnd-kit/core`, `@dnd-kit/sortable`, Go/Gin backend, Neon PostgreSQL.

---

## Product Requirement

User wants a “เปิดคอมวางแผนงานทั้งหมด” feel:

- Team lead opens a large planning workspace.
- All teams participating in the large work are visible as lanes.
- Team lead creates task/job cards describing what each point needs.
- Team lead can drag/drop cards into the team that should do the work.
- Team lead can reorder cards inside each team lane.
- Team lead can still use plain form input if drag/drop is not convenient.
- Each team ends up with its own clear work queue.
- Worker later sees only their team’s assigned cards in `คิวงานของฉัน`.

---

## Recommended UX

### Desktop layout: planning board

```text
┌──────────────────────────────────────────────────────────────┐
│ งานระดมทีม: เปลี่ยนลูกถ้วย                                  │
│ วันที่ / พื้นที่ / สถานะ / ทีมทั้งหมด / จุดทั้งหมด          │
├───────────────┬──────────────────────────────────────────────┤
│ งานที่ยังไม่  │ ทีมสตูล                                      │
│ มอบหมาย       │  [จุด P-001 เปลี่ยนลูกถ้วย]                  │
│               │  [จุด P-002 ตรวจแนวสาย]                     │
│ + สร้างการ์ด  ├──────────────────────────────────────────────┤
│ จากฟอร์ม      │ ทีมหาดใหญ่                                   │
│ + template    │  [จุด P-003 ตัดต้นไม้]                       │
│ + paste list  ├──────────────────────────────────────────────┤
│               │ ทีมสงขลา                                     │
│               │  ว่าง — ลากงานมาวางหรือกด + เพิ่มงานทีมนี้  │
└───────────────┴──────────────────────────────────────────────┘
```

### Mobile layout: accordion lanes

Drag/drop on mobile can be awkward. Keep it, but do not rely on it.

```text
[งานยังไม่มอบหมาย 3]
  + เพิ่มการ์ด

[ทีมสตูล 2]
  + เพิ่มงานทีมนี้
  การ์ด P-001
  การ์ด P-002

[ทีมหาดใหญ่ 1]
  + เพิ่มงานทีมนี้
```

Mobile primary flow:
- tap `+ เพิ่มงานทีมนี้`
- fill card form
- save draft
- optional: move team via dropdown inside card form

---

## Card Shape

Each card should answer “งานต้องแบบไหน/ยังไง” clearly.

Required:
- `assignedTeamId` once assigned
- `pointLabel` / ชื่อจุด เช่น `P-001`
- `workType` / ประเภทงาน เช่น `เปลี่ยนลูกถ้วย`, `ตัดต้นไม้`, `ตรวจแนวสาย`
- `workDetail` / รายละเอียดงาน

Optional but useful:
- `locationText` / พื้นที่หรือคำอธิบายตำแหน่ง
- `latitude`, `longitude`
- `pointCount`
- `treeCount`
- `itemCount`
- `notes`
- `priority`: low/normal/high/urgent (frontend draft first; backend later if needed)
- `estimatedMinutes` (frontend draft first; backend later if needed)

Execution fields remain worker-owned:
- before photos
- after photos
- result note
- status

---

## Interaction Model

### Create card methods

Support 4 ways:

1. Quick card
   - title/work type + detail only
   - starts in `งานยังไม่มอบหมาย`

2. Add under team
   - click `+ เพิ่มงานทีมนี้`
   - `assignedTeamId` prefilled

3. Drag from unassigned pool
   - create cards first
   - drag to a team lane

4. Plain table/form mode
   - toggle `กรอกแบบตาราง`
   - rows with team dropdown, point label, work type, detail
   - same validation and same save endpoint

### Save behavior

Draft state remains local until user clicks:

```text
บันทึกแผนจุดงาน
```

On save:
- validate every card has team + enough details
- assign sequence by lane order
- call `POST /v1/large-works/:id/tasks`
- invalidate `largeWorkTasks`, `largeWorkOverview`, `largeWorkMyTodos`

### Unsaved changes safety

If user closes board with unsaved changes:
- show confirm dialog
- optionally persist draft in `localStorage` keyed by `largeWorkItemId`

---

## Data Model Mapping

Frontend draft card:

```ts
interface LargeWorkPlanningCardDraft {
  clientId: string
  assignedTeamId: string
  pointLabel: string
  locationText: string
  latitude: string
  longitude: string
  workType: string
  workDetail: string
  pointCount: string
  treeCount: string
  itemCount: string
  notes: string
  priority?: 'low' | 'normal' | 'high' | 'urgent'
  estimatedMinutes?: string
}
```

Save payload maps to existing API:

```ts
{
  tasks: drafts.map((card, index) => ({
    assignedTeamId: Number(card.assignedTeamId),
    sequence: index + 1,
    pointLabel: card.pointLabel,
    locationText: card.locationText,
    latitude: numberOrUndefined(card.latitude),
    longitude: numberOrUndefined(card.longitude),
    workType: card.workType,
    workDetail: card.workDetail,
    pointCount: numberOrUndefined(card.pointCount),
    treeCount: numberOrUndefined(card.treeCount),
    itemCount: numberOrUndefined(card.itemCount),
    notes: card.notes,
    metadata: {
      priority: card.priority,
      estimatedMinutes: numberOrUndefined(card.estimatedMinutes),
    },
  }))
}
```

Backend can accept priority/estimate in `metadata` first to avoid schema churn.

---

## Dependencies

Current `hotlines3/package.json` has no DnD library.

Add:

```bash
npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

Reason:
- supports pointer + keyboard sensors
- maintained
- works with React 19 style components better than older `react-beautiful-dnd`
- mobile/touch can be enabled, but plain-form fallback remains required

---

## Implementation Tasks

### Task 1: Production DB prerequisite

Before UI work can truly function, `large_work_tasks` must exist in Neon.

See:
- `31-hotline-large-work-team-task-queue-fix-plan-2026-05-11.md`

### Task 2: Add planning-board helper tests

Files:
- Create/modify: `hotlines3/src/features/large-work/planning-board-helpers.ts`
- Create: `hotlines3/src/features/large-work/planning-board-helpers.test.ts`

Test behaviors:
- create empty card
- assign card to team
- move card between lanes
- reorder cards in one lane
- convert lanes to `LargeWorkAddTasksRequest`
- validation catches unassigned cards and missing work detail

### Task 3: Add planning board component

Files:
- Create: `hotlines3/src/features/large-work/components/LargeWorkPlanningBoard.tsx`

Component sections:
- Board header / plan summary
- Unassigned pool
- Team lanes
- Card editor drawer/dialog
- Plain table mode toggle
- Save bar with dirty-state count

### Task 4: Add drag/drop layer

Files:
- Modify: `LargeWorkPlanningBoard.tsx`

Use:
- `DndContext`
- `SortableContext`
- pointer sensor
- keyboard sensor

Rules:
- dragging into team lane sets `assignedTeamId`
- dragging into unassigned clears `assignedTeamId`
- reorder updates sequence only in draft
- save serializes lane order

### Task 5: Replace old tasks dialog entry point

Files:
- Modify: `hotlines3/src/app/(main)/planning/page.tsx`
- Deprecate or wrap: `hotlines3/src/features/large-work/components/LargeWorkTasksDialog.tsx`

Change CTA from:
- `จัดการจุดงาน`

to:
- `เปิดโต๊ะวางแผนงาน`

### Task 6: Plain-form mode

Inside `LargeWorkPlanningBoard`:
- table/list rows for users who do not want drag/drop
- each row has team dropdown
- row edit updates the same draft state as board mode
- save is identical

### Task 7: Worker queue integration test

Verify after saving:
- tasks grouped by team in overview
- `my-todos` loads only assigned team tasks
- no tasks from other teams appear

### Task 8: Mobile visual QA

Required viewports:
- 390x844
- 768x1024
- desktop 1440+

Checks:
- no horizontal overflow
- lane accordion readable
- tap targets >= 44px
- save bar not hidden by mobile bottom nav

---

## Acceptance Criteria

- Team lead opens a full-screen/large modal planning board from a large-work item.
- All participating teams appear as lanes.
- Team lead can create an unassigned task card.
- Team lead can drag the card to a team.
- Team lead can add a card directly under a team.
- Team lead can use plain-form/table mode without drag/drop.
- Save creates `large_work_tasks` records grouped by team.
- Worker queue shows only the current user’s team tasks.
- Overview shows per-team progress.
- UI clearly says why save is disabled if cards are incomplete.

---

## Recommended Delivery Order

1. Apply/verify `large_work_tasks` schema in Neon.
2. Build helper tests + draft state model.
3. Build non-DnD plain board first.
4. Add DnD as enhancement.
5. Replace old dialog CTA.
6. Run frontend tests/build.
7. Visual QA desktop + mobile.
8. Push backend note + frontend implementation.
