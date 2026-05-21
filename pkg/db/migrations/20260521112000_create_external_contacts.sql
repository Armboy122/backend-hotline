-- +goose Up
-- Contact Directory external entries used by /v1/contact-directory.
-- Backend production disables auto_migrate, so this table must exist before
-- the repository joins user contacts with external/emergency/operation-center contacts.
CREATE TABLE IF NOT EXISTS external_contacts (
    id BIGSERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    display_name TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    organization TEXT,
    position TEXT,
    notes TEXT,
    team_id BIGINT REFERENCES "Team"(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ(6)
);

CREATE INDEX IF NOT EXISTS idx_external_contacts_type
    ON external_contacts (type);

CREATE INDEX IF NOT EXISTS idx_external_contacts_team
    ON external_contacts (team_id);

CREATE INDEX IF NOT EXISTS idx_external_contacts_active
    ON external_contacts (is_active)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_external_contacts_active;
DROP INDEX IF EXISTS idx_external_contacts_team;
DROP INDEX IF EXISTS idx_external_contacts_type;
DROP TABLE IF EXISTS external_contacts;
