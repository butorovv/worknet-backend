BEGIN;

DROP TRIGGER IF EXISTS trg_space_links_touch_updated_at ON space_links;

DROP TABLE IF EXISTS space_links;
DROP TABLE IF EXISTS space_stack_items;

ALTER TABLE space_members
    DROP CONSTRAINT IF EXISTS ck_space_members_title_length,
    DROP CONSTRAINT IF EXISTS ck_space_members_role;

UPDATE space_members
SET role = 'viewer'
WHERE role = 'member';

-- Owner memberships inserted or promoted by 013 intentionally remain after rollback.
-- The owner role and active status are valid in the legacy schema and preserve access.
ALTER TABLE space_members
    ADD CONSTRAINT ck_space_members_role CHECK (
        role IN ('owner', 'admin', 'contributor', 'viewer')
    ),
    DROP COLUMN IF EXISTS title;

DROP INDEX IF EXISTS idx_spaces_visibility_status_created;

ALTER INDEX IF EXISTS ux_spaces_slug
    RENAME TO ux_spaces_slug_active;

ALTER TABLE spaces
    DROP CONSTRAINT IF EXISTS ck_spaces_archive_state,
    DROP CONSTRAINT IF EXISTS ck_spaces_accent_color,
    DROP CONSTRAINT IF EXISTS ck_spaces_cover_icon,
    DROP CONSTRAINT IF EXISTS ck_spaces_stage,
    DROP CONSTRAINT IF EXISTS ck_spaces_visibility,
    DROP CONSTRAINT IF EXISTS ck_spaces_status,
    DROP CONSTRAINT IF EXISTS ck_spaces_why_it_matters_length,
    DROP CONSTRAINT IF EXISTS ck_spaces_description_length,
    DROP CONSTRAINT IF EXISTS ck_spaces_short_description_length,
    DROP CONSTRAINT IF EXISTS ck_spaces_title_length,
    DROP CONSTRAINT IF EXISTS ck_spaces_slug_format;

-- Rollback is deterministic but intentionally lossy:
-- v0.1 paused combines legacy draft and in_review, so every paused row becomes draft.
-- v0.1 private combines legacy private, registered and members, so it remains private.
UPDATE spaces
SET
    status = CASE status
        WHEN 'paused' THEN 'draft'
        WHEN 'archived' THEN 'archived'
        ELSE 'active'
    END,
    visibility = CASE visibility
        WHEN 'public' THEN 'public'
        ELSE 'private'
    END;

ALTER TABLE spaces
    ALTER COLUMN status SET DEFAULT 'draft',
    ADD CONSTRAINT ck_spaces_slug_format CHECK (
        slug ~ '^[a-z0-9][a-z0-9-]{2,80}$'
    ),
    ADD CONSTRAINT ck_spaces_name_not_empty CHECK (
        char_length(trim(title)) > 1
    ),
    ADD CONSTRAINT ck_spaces_status CHECK (
        status IN ('draft', 'active', 'in_review', 'archived')
    ),
    ADD CONSTRAINT ck_spaces_visibility CHECK (
        visibility IN ('public', 'registered', 'members', 'private')
    );

ALTER TABLE spaces
    DROP COLUMN IF EXISTS accent_color,
    DROP COLUMN IF EXISTS cover_icon,
    DROP COLUMN IF EXISTS stage,
    DROP COLUMN IF EXISTS why_it_matters;

ALTER TABLE spaces
    RENAME COLUMN title TO name;

COMMIT;
