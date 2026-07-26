-- +goose Up
-- Monthly team schedule authoring and immutable publication for Clinic Tool.

ALTER TABLE "Team" ADD COLUMN IF NOT EXISTS code TEXT;
ALTER TABLE "Team" ADD COLUMN IF NOT EXISTS base_area TEXT;
ALTER TABLE "Team" ADD COLUMN IF NOT EXISTS crew_type TEXT;
ALTER TABLE "Team" ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "Team" ADD COLUMN IF NOT EXISTS monthly_plan_visible BOOLEAN NOT NULL DEFAULT TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS team_code_unique_idx
    ON "Team" (code)
    WHERE code IS NOT NULL;

CREATE INDEX IF NOT EXISTS team_monthly_plan_order_idx
    ON "Team" (monthly_plan_visible, display_order);

CREATE TABLE IF NOT EXISTS monthly_plan_schedule_revisions (
    id BIGSERIAL PRIMARY KEY,
    monthly_plan_id BIGINT NOT NULL REFERENCES "MonthlyPlan"(id),
    revision_no INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_by_user_id BIGINT NOT NULL REFERENCES "User"(id),
    published_by_user_id BIGINT REFERENCES "User"(id),
    published_at TIMESTAMPTZ(6),
    checksum TEXT,
    projection JSONB,
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT monthly_plan_schedule_status_check
        CHECK (status IN ('draft', 'published', 'superseded')),
    CONSTRAINT monthly_plan_schedule_revision_no_positive
        CHECK (revision_no > 0),
    CONSTRAINT monthly_plan_schedule_publication_check
        CHECK (
            (status = 'draft' AND published_by_user_id IS NULL AND published_at IS NULL)
            OR
            (status IN ('published', 'superseded') AND published_by_user_id IS NOT NULL AND published_at IS NOT NULL)
        )
);

CREATE UNIQUE INDEX IF NOT EXISTS monthly_plan_schedule_revision_no_idx
    ON monthly_plan_schedule_revisions (monthly_plan_id, revision_no);

CREATE INDEX IF NOT EXISTS monthly_plan_schedule_status_idx
    ON monthly_plan_schedule_revisions (monthly_plan_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS monthly_plan_schedule_single_draft_idx
    ON monthly_plan_schedule_revisions (monthly_plan_id)
    WHERE status = 'draft';

CREATE UNIQUE INDEX IF NOT EXISTS monthly_plan_schedule_single_published_idx
    ON monthly_plan_schedule_revisions (monthly_plan_id)
    WHERE status = 'published';

CREATE TABLE IF NOT EXISTS monthly_plan_team_assignments (
    id BIGSERIAL PRIMARY KEY,
    revision_id BIGINT NOT NULL REFERENCES monthly_plan_schedule_revisions(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES "Team"(id),
    assignment_type TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    destination TEXT NOT NULL,
    note TEXT,
    source_type TEXT NOT NULL DEFAULT 'manual',
    source_id BIGINT,
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT monthly_plan_assignment_type_check
        CHECK (assignment_type IN ('field', 'remote', 'support', 'special')),
    CONSTRAINT monthly_plan_assignment_source_check
        CHECK (source_type IN ('manual', 'large_work', 'approved_file')),
    CONSTRAINT monthly_plan_assignment_date_order_check
        CHECK (end_date >= start_date),
    CONSTRAINT monthly_plan_assignment_destination_check
        CHECK (length(trim(destination)) > 0)
);

CREATE INDEX IF NOT EXISTS monthly_plan_assignment_team_date_idx
    ON monthly_plan_team_assignments (revision_id, team_id, start_date);

-- +goose Down
DROP INDEX IF EXISTS monthly_plan_assignment_team_date_idx;
DROP TABLE IF EXISTS monthly_plan_team_assignments;

DROP INDEX IF EXISTS monthly_plan_schedule_single_published_idx;
DROP INDEX IF EXISTS monthly_plan_schedule_single_draft_idx;
DROP INDEX IF EXISTS monthly_plan_schedule_status_idx;
DROP INDEX IF EXISTS monthly_plan_schedule_revision_no_idx;
DROP TABLE IF EXISTS monthly_plan_schedule_revisions;

DROP INDEX IF EXISTS team_monthly_plan_order_idx;
DROP INDEX IF EXISTS team_code_unique_idx;
ALTER TABLE "Team" DROP COLUMN IF EXISTS monthly_plan_visible;
ALTER TABLE "Team" DROP COLUMN IF EXISTS display_order;
ALTER TABLE "Team" DROP COLUMN IF EXISTS crew_type;
ALTER TABLE "Team" DROP COLUMN IF EXISTS base_area;
ALTER TABLE "Team" DROP COLUMN IF EXISTS code;
