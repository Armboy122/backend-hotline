# Hotline Operations PRD ล่าสุด + Kanban Scope

## Document Status

อัปเดตจากการตรวจ Obsidian/project-local vault และ repo จริงเมื่อ 2026-05-08

เอกสารนี้ใช้เป็น PRD ฉบับทำงานสำหรับโปรเจค Hotline หลังการ refactor ใหญ่ และเป็นฐานสำหรับแตกงานลง Kanban

## Project Locations

- Backend / Obsidian project-local vault: `/Users/sakdithat/Desktop/myproject/backend-hotline`
- Frontend ปัจจุบันที่ตรงกับ backend API: `/Users/sakdithat/Desktop/myproject/hotlines3`
- Frontend เก่า/คนละแนวทาง: `/Users/sakdithat/Desktop/myproject/hotline-2` — ผู้ใช้ยืนยันให้ลบ และลบออกจากเครื่องแล้วเมื่อ 2026-05-08

## Verified Current State

### Backend

- Repo: `backend-hotline`
- Branch ที่ตรวจ: `main`
- Module: `backend-hotlines3`
- Stack: Go, Gin, GORM/PostgreSQL, Viper, JWT, Cloudflare R2/S3-compatible upload
- Architecture ล่าสุด: Hinghoi-style feature-first package ภายใต้ `internal/feature/<feature>/...`
- Feature migration refactor เสร็จตาม plan เดิมแล้ว:
  - task daily
  - monthly plan
  - dashboard
  - master data
  - auth
  - user
  - upload
- Quality gate ที่รันแล้วผ่าน:
  - `go test ./...`
  - `go vet ./...`
  - `go build -o /tmp/hotlines-api main.go`

### Frontend

- Repo: `hotlines3`
- Branch ที่ตรวจ: `main`
- Stack: Next.js, React, TypeScript, React Query, axios, Tailwind/shadcn/ui, antd-mobile
- Architecture ล่าสุด: pure client-side SPA consuming Go backend through REST API/proxy
- Auth ใช้ JWT Bearer token ไม่ใช่ cookie middleware
- Quality gate ที่รันแล้วผ่าน:
  - `npx tsc --noEmit`

## Product Definition ล่าสุด

Hotline Operations คือระบบจัดการงาน Hotline/งานบำรุงรักษาภาคสนามที่รวม:

1. การบันทึกข้อมูลงานประจำวัน
2. การจัดการข้อมูลหลัก เช่น ทีม ประเภทงาน รายละเอียดงาน สถานี สายป้อน การไฟฟ้า จุดรวมงาน
3. การแนบหลักฐานก่อน/หลังทำงานและพิกัด
4. การจัดการ/ส่งไฟล์หรือข้อมูลแผนงานรายเดือน เริ่มใช้จริงรอบแผนเดือนมิถุนายน 2569/2026
5. Dashboard/รายงานสำหรับติดตามภาพรวม
6. การจัดการผู้ใช้และสิทธิ์แบบ role-based

## Requirements จากผู้ใช้รอบนี้

### R1 Daily Operation Data

ระบบต้องบันทึกข้อมูลประจำวันได้ครบถ้วน

สถานะปัจจุบัน:
- Backend มี `TaskDaily` และ endpoint `/v1/tasks*`
- Frontend มีหน้าบันทึกข้อมูลและรายการงาน
- ยังต้องตรวจ UX/permission ตาม requirement ใหม่

### R2 Monthly Team Planning Submission

หัวหน้าแต่ละทีม (`team_lead`) ต้องส่งแผนงานเพื่อวางแผนเดือนถัดไปผ่านระบบ `monthly-plan` แทนการส่งใน LINE

บริบทการทำงานจริง:
- มีการประชุมประจำหัวหน้าชุดก่อนวางแผนเดือนถัดไป
- กำหนดส่งเอกสารก่อนประมาณวันที่ 20 ของเดือน หรือตามค่าที่ admin/super_admin ตั้งค่าในระบบ
- แต่ละทีมต้องส่งข้อมูลแผน เช่น ไปทำงานที่ไหน ช่วงวันที่เริ่ม-สิ้นสุด แนบเอกสารอื่น ๆ และหมายเหตุ
- ไม่ต้องมี approval workflow เพิ่ม ใช้แนวคิด monthly plan เดิม: ส่งไฟล์/ข้อมูลแล้วนับเป็นสถานะการส่ง

สถานะปัจจุบัน:
- Backend/Frontend มี `monthly-plan` ในรูปแบบ upload file ต่อปี/เดือน และสถานะการส่งตามทีม
- PRD ยืนยันให้ใช้ monthly plan เป็นระบบวางแผนเดือนถัดไป ไม่ต้องทำ calendar approval workflow
- ต้องเพิ่ม/ปรับ field ให้รองรับช่วงวันที่ สถานที่/งานที่จะไปทำ เอกสารแนบ และหมายเหตุ เพื่อแยกงานของทีมได้ชัดเจน

### R3 Role And Admin Model

ต้องกำหนดได้ว่าใครเป็น admin และใครเป็น user ธรรมดา

Requirement ใหม่:
- admin มีหลายคนได้
- super admin มีได้แค่คนเดียว
- super admin ต้องสร้างรหัสใหม่/รีเซ็ตรหัสได้
- admin ปัจจุบันต้องทำงานได้ตาม scope ที่วางไว้

สถานะปัจจุบัน:
- Backend model `User.Role` เป็น string ค่า default `user`
- Frontend type รองรับ `admin | user | viewer`
- Backend route `/v1/users` จำกัดด้วย `RequireRole("admin")`
- Frontend `AdminGuard` อนุญาตเฉพาะ `user.role === 'admin'`
- ยังไม่มี `super_admin` ใน type/model/policy
- ยังไม่มี invariant “super admin มีได้แค่หนึ่งคน”
- ยังต้องแยก capability ของ admin vs super admin อย่างเป็นทางการ

## Proposed Role Model Draft

> ต้องให้ผู้ใช้ยืนยันก่อน implement

### super_admin

- ใช้ role ชื่อ `super_admin` ตรง ๆ ในระบบ
- มีได้แค่ 1 คนทั้งระบบ
- สร้างจาก local/CLI ได้เพียงครั้งเดียว โดยใช้ DB จริงจาก env ปัจจุบัน
- เป็น real system administrator: ทำงานได้เหมือนมี direct DB-backed CRUD ผ่าน application
- อ่านข้อมูลได้ทุกอย่าง
- CRUD ผู้ใช้/role ได้ทั้งหมด รวมถึง custom user และ promote/demote admin/team_lead/user/viewer
- reset password ให้ admin/user/team_lead/viewer ได้
- CRUD master data, operational data, dashboard/report, monthly plan ได้ทั้งหมด
- override/manage locked monthly plan period ได้ตาม policy
- ไม่ควรถูกลบหรือปิดใช้งานโดย admin ธรรมดา

### admin

- มีหลายคนได้
- ไม่ใช่ system administrator แบบเต็ม
- เป็น operational monthly-plan manager เป็นหลัก
- login และใช้งาน flow พื้นฐานได้ตาม scope
- upload/download/replace monthly master plan ให้ทีมใช้ได้ตามช่วงเวลาที่ระบบอนุญาต รวมถึงจัดการรอบแผนเริ่มใช้จริงเดือนมิถุนายน 2569/2026
- ดูสถานะ monthly plan ที่จำเป็นต่อการปฏิบัติงานได้
- ไม่จัดการ user/role/password
- ไม่ CRUD master data/global data โดย default
- ไม่เห็นหรือใช้งาน super-admin-only console/actions
- สร้าง admin คนอื่นไม่ได้
- reset password ไม่ได้; การ reset password เป็นสิทธิ์ของ `super_admin` เท่านั้น

### team_lead

- เป็น role แยก ไม่ใช่ flag
- ผูกกับทีมผ่าน `teamId`
- เห็นรายการ monthly plan ได้ทุกทีมเพื่อ awareness
- upload/download ได้เฉพาะไฟล์/แผนของทีมตัวเองตาม policy
- ไม่ upload แผนของทีมอื่น; admin/super_admin เป็นผู้ manage ได้กว้างกว่า
- เห็น task/แผนของทีมตัวเองตาม policy

### user

- บันทึกงานประจำวันได้
- เห็น task เฉพาะทีมตัวเอง
- ไม่สามารถจัดการ admin/user หรือ reset password ผู้อื่นได้

### viewer (ถ้ายังต้องใช้)

- ดูข้อมูล/รายงานอย่างเดียว
- ไม่แก้ไขข้อมูล

## Open Questions ต้องถามผู้ใช้ก่อนแตก implementation ลึก

1. ต้องการ role ชื่อ `super_admin` ตรง ๆ ในระบบไหม หรือใช้ `admin` + flag เช่น `isSuperAdmin`?
2. “หัวหน้าแต่ละทีม” คือ role ใหม่ (`team_lead`) หรือเป็น user ที่มี flag/permission ในทีม?
3. หัวหน้าทีมส่งแผนเดือนถัดไปเป็นไฟล์แนบเหมือนระบบปัจจุบันพอไหม หรืออยากให้กรอกเป็นรายการงานล่วงหน้าแบบ structured plan?
4. แผนงานเดือนถัดไปต้องมี approval ไหม? ถ้ามี ใคร approve: admin, super_admin, หรือหัวหน้าระดับบน?
5. user ธรรมดาควรเห็น task เฉพาะทีมตัวเอง หรือเห็นทุกทีมแต่แก้ได้เฉพาะของตัวเอง/ทีมตัวเอง?
6. admin ธรรมดาควรสร้าง admin คนอื่นได้ไหม หรือเฉพาะ super_admin เท่านั้น?
7. admin ธรรมดาควร reset password ให้ user ได้ไหม? แล้ว reset ให้ admin คนอื่นได้หรือไม่?
8. super_admin คนแรกจะถูกสร้างอย่างไร: seed CLI, env var bootstrap, หรือผ่าน migration/one-time setup page?
9. ต้องเก็บ audit log สำหรับการเปลี่ยน role/reset password/upload monthly plan ไหม?
10. ต้องการเก็บ frontend เก่า `/Users/sakdithat/Desktop/myproject/hotline-2` ไว้ไหม หรือถือว่า `/Users/sakdithat/Desktop/myproject/hotlines3` เป็น frontend หลักเท่านั้น?

## Neon Schema Check 2026-05-08

ตรวจผ่าน Neon MCP project `hotlines3`:
- Project ID: `bitter-lake-05690037`
- Default branch: `production` / `br-snowy-thunder-a1teoes9`
- PostgreSQL version: 17
- Schema ใช้ PascalCase tables ใน `public`
- Tables สำคัญที่มีแล้ว: `User`, `Team`, `TaskDaily`, `MonthlyPlan`, `MonthlyPlanSetting`, `PlanFile`, `FileSizeLog`, master data tables
- `User.role` เป็น `text NOT NULL DEFAULT 'user'`; ยังไม่มี constraint/enforcement สำหรับ `super_admin` คนเดียว
- `User.teamId` nullable มี index แล้ว ใช้เป็นฐาน team scoping ได้
- `TaskDaily.teamId` เป็น required และมี index `TaskDaily_teamId_idx` + composite `TaskDaily_workdate_teamId_idx`; เหมาะกับ requirement user เห็นเฉพาะทีมตัวเอง
- `MonthlyPlanSetting.lockDay` default `23`, `reminderStartDay` default `20`; ตรงกับ requirement ส่งก่อนประมาณวันที่ 20 และล็อกหลังวันกำหนด
- `PlanFile` รองรับ `teamId`, `uploadedById`, `description`, file metadata, soft delete; ยังไม่มี field structured เช่น `workStartDate`, `workEndDate`, `destination/location`, `remarks` แยกจาก `description`

Implication:
- K1 ต้องเพิ่ม application-level + DB-level invariant สำหรับ `super_admin` คนเดียว โดยใช้ migration/test-first
- K2 ควรต่อยอด monthly plan เดิม ไม่สร้าง approval workflow ใหม่
- K2 ต้องออกแบบ migration เพิ่ม structured plan fields หรือ table ลูกสำหรับ plan items โดยคำนึงถึง zero-downtime เพราะใช้ Neon DB จริง

## Decision Log 2026-05-08

ผู้ใช้ยืนยัน decision สำหรับ implementation แล้ว:

1. ใช้ role ชื่อ `super_admin` ตรง ๆ
2. หัวหน้าทีมเป็น role แยกชื่อ `team_lead`
3. `monthly-plan` คือระบบวางแผนเดือนถัดไป: หลังประชุมหัวหน้าชุด แต่ละทีมส่งเอกสาร/ข้อมูลในระบบแทน LINE โดยระบุไปไหน ช่วงวันที่เริ่ม-สิ้นสุด เอกสารแนบ และหมายเหตุ
4. ไม่ต้องมี approval workflow เพิ่ม ใช้ flow monthly plan เดิม
5. `user` เห็นเฉพาะ task ของทีมตัวเอง
6. การสร้าง admin ใหม่ต้องเป็น `super_admin` เท่านั้น
7. การ reset password ของผู้อื่นต้องเป็น `super_admin` เท่านั้น; admin reset ไม่ได้
8. `super_admin` คนแรกสร้างจาก local/CLI ได้ครั้งเดียว โดยต่อ DB จริงจาก env; DB เป็น Neon และให้ตรวจ schema ผ่าน MCP ก่อนทำ migration/seed
9. ไม่ต้องทำ audit log สำหรับรอบนี้
10. Frontend หลักคือ `/Users/sakdithat/Desktop/myproject/hotlines3`; repo เก่า `/Users/sakdithat/Desktop/myproject/hotline-2` ผู้ใช้ยืนยันให้ลบ และลบออกจากเครื่องแล้ว

## Decision Log 2026-05-09 — Performance/RBAC/Monthly Plan Replan

ผู้ใช้ยืนยันเพิ่มเติมหลังทดสอบ frontend dev:

1. ต้องตรวจ performance และแก้ความหน่วงของหน้าบ้านก่อนงาน feature อื่น เพื่อไม่ให้กลายเป็น critical issue
2. `admin` ไม่ใช่ system administrator เต็มรูปแบบ; admin เป็นผู้จัดการ monthly plan operational data เท่านั้น
3. `super_admin` เป็นผู้ดูแลสูงสุด ทำได้ทั้งหมดแบบ application-level full CRUD/read access คล้ายการทำงานผ่าน DB โดยตรง แต่ผ่าน policy/API ของระบบ
4. `admin` ต้องไม่ทำ user/role/password management และไม่ใช้ super-admin-only actions
5. Correction ล่าสุด: ไม่ใช่ว่า monthly-plan ของทีมถูกยกเลิก; ระบบเริ่มใช้จริงรอบแผนเดือนมิถุนายน 2569/2026 และยังต้องส่ง/จัดการแผนตามเงื่อนไข lock
6. สิทธิ์ upload/manage monthly plan: `admin` และ `super_admin` ทำได้; `team_lead` upload ได้เฉพาะแผนทีมตัวเอง; `user` upload ไม่ได้แต่ดูได้
7. การมองเห็น monthly plan: team_lead/user เห็นรายการของทุกทีมได้; team_lead upload/download ได้เฉพาะทีมตัวเอง; user ดูได้แต่ upload ไม่ได้
8. `admin` แก้ monthly-plan settings ได้
9. หน้า monthly plan ต้องโชว์ทั้งปี 2569 / 2026 ไม่ fix เดือนย่อยหรือ Jan/Feb/Mar เท่านั้น
10. แผนเดือนมิถุนายน 2569/2026 ต้อง upload ได้ถึงวันที่ 23 พฤษภาคม ตาม `MonthlyPlanSetting.lockDay = 23` ใน DB
11. เพิ่ม replan doc: `plan/12-performance-rbac-monthly-plan-replan.md`

## Kanban Implementation Phases

### K0 Confirm PRD And Repo Baseline

Goal: ล็อก scope, roles, repo หลัก, และ acceptance criteria ก่อนเขียน code

Deliverables:
- PRD ที่ผู้ใช้ยืนยัน
- role matrix
- data ownership rules
- Definition of Done ของ MVP

### K1 Auth/RBAC Foundation

Goal: เพิ่ม super_admin/admin/team_lead/user permission model แบบ test-first

Deliverables:
- backend role constants/policy
- one-super-admin invariant
- password reset policy
- route guards
- frontend type/guard update
- tests ครอบ role edge cases

### K2 Team Leadership And Monthly Plan Workflow

Goal: ให้หัวหน้าทีมส่งแผนงานเดือนถัดไปได้ตามสิทธิ์

Deliverables ขึ้นกับคำตอบ:
- ถ้าเป็น file-based: ปรับ monthly plan permission/status/deadline UX
- ถ้าเป็น structured plan: เพิ่ม model/API/UI สำหรับ plan items
- tests สำหรับ own-team submission + admin/super_admin review

### K3 Daily Task Completion Hardening

Goal: ทำระบบบันทึกประจำวันให้สมบูรณ์ตาม workflow จริง

Deliverables:
- validate required fields
- team scoping
- image/location behavior
- edit/delete rules
- frontend UX checks
- regression tests

### K4 Admin Console Completion

Goal: admin ทำงานตาม scope ได้ครบ แต่ไม่ล้ำสิทธิ์ super_admin

Deliverables:
- user management UI policy
- master data permissions
- reset password UI/API
- audit-sensitive confirmation flows

### K5 Reporting/Operational Readiness

Goal: dashboard/report/runbook พร้อมใช้งานจริง

Deliverables:
- dashboard role scoping
- monthly submission status report
- smoke tests
- deployment/runbook update

## Current Recommendation

ยังไม่ควรเริ่ม implement ทันทีจนกว่าจะตอบ Open Questions อย่างน้อย Q1-Q8 เพราะ role model และ monthly plan shape จะกระทบ DB/API/UI หลายจุด

แต่สามารถสร้าง Kanban board ได้ทันทีโดยเริ่มจาก K0 และ task discovery ที่ไม่เปลี่ยน production code
