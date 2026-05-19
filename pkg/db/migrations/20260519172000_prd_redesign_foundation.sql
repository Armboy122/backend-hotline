-- Hotline PRD redesign foundation.

-- Final role model removes admin. Existing admin users become team_lead so
-- team-scoped operational ownership is preserved.
UPDATE "User" SET role = 'team_lead' WHERE role = 'admin';

-- Planning Board backlog: draft work may exist before a date is chosen.
ALTER TABLE team_plans ALTER COLUMN start_date DROP NOT NULL;
ALTER TABLE team_plans ALTER COLUMN status SET DEFAULT 'draft';

-- Round-1 capability model.
CREATE TABLE IF NOT EXISTS user_capabilities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "User"(id),
    code TEXT NOT NULL,
    granted_by_user_id BIGINT REFERENCES "User"(id),
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ(6)
);

CREATE UNIQUE INDEX IF NOT EXISTS user_capabilities_active_user_code_idx
    ON user_capabilities (user_id, code)
    WHERE revoked_at IS NULL;

-- Daily Report source attribution for Work Report and Large Work idempotency.
ALTER TABLE "TaskDaily" ADD COLUMN IF NOT EXISTS source_type TEXT;
ALTER TABLE "TaskDaily" ADD COLUMN IF NOT EXISTS source_id BIGINT;
ALTER TABLE "TaskDaily" ADD COLUMN IF NOT EXISTS large_work_task_id BIGINT REFERENCES large_work_tasks(id);

CREATE UNIQUE INDEX IF NOT EXISTS taskdaily_large_work_task_unique_idx
    ON "TaskDaily" (large_work_task_id)
    WHERE large_work_task_id IS NOT NULL AND deletedat IS NULL;
