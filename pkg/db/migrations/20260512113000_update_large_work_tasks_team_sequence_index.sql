-- +goose Up
DROP INDEX IF EXISTS idx_large_work_tasks_plan_sequence;
CREATE UNIQUE INDEX IF NOT EXISTS idx_large_work_tasks_plan_team_sequence
    ON large_work_tasks (large_work_item_id, assigned_team_id, sequence);

-- +goose Down
DROP INDEX IF EXISTS idx_large_work_tasks_plan_team_sequence;
CREATE UNIQUE INDEX IF NOT EXISTS idx_large_work_tasks_plan_sequence
    ON large_work_tasks (large_work_item_id, sequence);
