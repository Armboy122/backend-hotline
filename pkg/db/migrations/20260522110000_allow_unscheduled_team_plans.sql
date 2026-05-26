-- +goose Up
-- Team plans may be captured before a work date is known. Those records stay
-- in the planning board backlog with status draft until a start_date is set.
ALTER TABLE team_plans
    ALTER COLUMN start_date DROP NOT NULL;

ALTER TABLE team_plans
    DROP CONSTRAINT IF EXISTS team_plans_status_check;

ALTER TABLE team_plans
    ADD CONSTRAINT team_plans_status_check
    CHECK (status IN ('draft', 'planned', 'cancelled', 'completed'));

-- +goose Down
UPDATE team_plans
SET
    start_date = COALESCE(start_date, CURRENT_DATE),
    status = CASE WHEN status = 'draft' THEN 'planned' ELSE status END
WHERE start_date IS NULL OR status = 'draft';

ALTER TABLE team_plans
    DROP CONSTRAINT IF EXISTS team_plans_status_check;

ALTER TABLE team_plans
    ALTER COLUMN start_date SET NOT NULL;

ALTER TABLE team_plans
    ADD CONSTRAINT team_plans_status_check
    CHECK (status IN ('planned', 'cancelled', 'completed'));
