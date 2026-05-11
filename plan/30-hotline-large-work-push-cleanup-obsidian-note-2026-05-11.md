# Hotline Large Work — Push & Cleanup Note

Date: 2026-05-11
Project vault: `/Users/sakdithat/Desktop/myproject/backend-hotline`
Related flow: `งานระดมทีม / large-work execution`

## Goal

Record what was pushed, what was cleaned from the local project folders, and where the backup lives after the `งานระดมทีม` implementation and infographic work.

## Product requirement snapshot

Corrected `งานระดมทีม` flow:

1. `team_lead` creates and edits the main large-work plan.
2. `team_lead` splits the plan into point-by-point execution tasks.
3. `team_lead` assigns work to other teams without Admin approval.
4. All teams can view the overview/progress.
5. Workers see only their own-team todo queue.
6. Workers execute one point at a time:
   - start task
   - add before photo URL
   - add after photo URL
   - add completion note
   - complete task
7. Lead monitors blocked/not-done work and closes the plan.
8. Watchdog reports `why_not_done` when work is unfinished.

## Remote commits pushed

### Backend repo

- Repo: `git@github.com:Armboy122/backend-hotline.git`
- Branch: `main`
- Commit: `f2dca64 feat: add large-work execution flow`
- Verified remote SHA: `f2dca64ce03468122b195833ede8f89b791f0f4f`

Main pushed scope:

- Large-work backend policy/DTO/entity/service/repository/controller updates
- Large-work execution task storage migration
- Router wiring for overview/tasks/my-todos/start/photos/complete/block
- Backend tests for policy/entity/repository/service/controller/router/migrations
- Plan docs:
  - `plan/28-hotline-large-work-execution-replan-2026-05-11.md`
  - `plan/29-hotline-large-work-execution-qa-handoff-2026-05-11.md`

### Frontend repo

- Repo: `git@github.com:Armboy122/hotlines3.git`
- Branch: `main`
- Commit: `d70fbaf feat: add large-work execution UI`
- Verified remote SHA: `d70fbafe1c4f75303eb158638e5915d6239e6d65`

Main pushed scope:

- Planning page large-work execution UI
- Large-work overview panel
- Large-work task assignment dialog
- Worker todo queue UI
- Large-work service/hooks/types
- Role policy updates
- Frontend helper tests and type/service tests

## Verification before push

Backend verification in clean clone:

- PASS: `go test ./internal/feature/largework/... ./internal/router/... ./pkg/db/migrations/... -count=1`
- PASS: `go build -o /tmp/hotlines-api main.go`
- PASS: `git diff --check`

Frontend verification in clean clone:

- PASS: `npx --yes tsx src/lib/auth/role-policy.test.ts`
- PASS: `npx --yes tsx src/types/large-work.test.ts`
- PASS: `npx --yes tsx src/features/large-work/worker-todo-flow.test.ts`
- PASS: `npx tsc --noEmit`
- PASS: `npm run build`
- PASS: `git diff --check`

## Cleanup performed

The primary working trees had macOS/iCloud filesystem issues while indexing Git objects:

- `Resource deadlock avoided`
- failure while `git add` / `git fetch` / `.git/objects` writes

Cleanup approach:

1. Move dirty/corrupted working trees to backup.
2. Clone fresh copies from remote into the original project paths.
3. Remove temporary clean clones used for push.
4. Move the generated infographic HTML out of the project root.
5. Verify both repos are clean and point to the pushed commits.

Current clean project paths:

- Backend: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Frontend: `/Users/sakdithat/Desktop/myproject/hotlines3`

Backup path:

- `/Users/sakdithat/Desktop/hotline-cleanup-backup-20260511-150738`

Backup contains:

- old `backend-hotline` working tree
- old `hotlines3` working tree
- generated infographic HTML file: `hotline-largework-flow-infographic.html`

Temporary push clones removed:

- `/tmp/backend-hotline-push`
- `/tmp/hotlines3-push`

## Files intentionally not pushed

These were treated as local artifacts or unsafe to include in the commit:

- Generated infographic HTML in project root
- `dogfood-output/`
- dirty/problematic `CLAUDE.md` state from the old frontend working tree
- unrelated untracked/dirty files from previous sessions

## Infographic summary

A Thai infographic was generated to explain/check the new `งานระดมทีม` flow.

Included checklist:

- `team_lead` creates and edits plans
- `team_lead` assigns other teams
- all teams view overview
- worker sees own-team todo
- work is executed point-by-point
- completion requires before photo + after photo + note
- watchdog reports `why_not_done`
- Admin is not the workflow bottleneck

Generated screenshot path from the Hermes browser cache at creation time:

- `/Users/sakdithat/.hermes/profiles/scc/cache/screenshots/browser_screenshot_73d4811517dc4f6da9abc5d1641723db.png`

The editable HTML source was moved to backup:

- `/Users/sakdithat/Desktop/hotline-cleanup-backup-20260511-150738/hotline-largework-flow-infographic.html`

## Current status

- `backend-hotline` is clean and at remote `main` commit `f2dca64`.
- `hotlines3` is clean and at remote `main` commit `d70fbaf`.
- Kanban board `hotline-largework-execution` is complete: `8/8 done`.
- Watchdog job `0d212ef5aed2` is paused because the board drained.

## Related notes

- [[28-hotline-large-work-execution-replan-2026-05-11]]
- [[29-hotline-large-work-execution-qa-handoff-2026-05-11]]
