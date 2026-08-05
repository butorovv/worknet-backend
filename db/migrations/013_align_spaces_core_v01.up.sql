BEGIN;

ALTER TABLE spaces
    RENAME COLUMN name TO title;

ALTER TABLE spaces
    ADD COLUMN why_it_matters text,
    ADD COLUMN stage text NOT NULL DEFAULT 'idea',
    ADD COLUMN cover_icon text NOT NULL DEFAULT 'other',
    ADD COLUMN accent_color text NOT NULL DEFAULT 'gray';

ALTER TABLE spaces
    ALTER COLUMN status SET DEFAULT 'active',
    DROP CONSTRAINT IF EXISTS ck_spaces_slug_format,
    DROP CONSTRAINT IF EXISTS ck_spaces_name_not_empty,
    DROP CONSTRAINT IF EXISTS ck_spaces_status,
    DROP CONSTRAINT IF EXISTS ck_spaces_visibility;

-- Normalize every legacy row before installing strict v0.1 constraints.
-- Invalid long slugs keep a readable prefix and receive a deterministic ID suffix.
DO $$
DECLARE
    current_space record;
    base_slug text;
    candidate text;
    suffix text;
    attempt integer;
BEGIN
    FOR current_space IN
        SELECT id, slug
        FROM spaces
        WHERE slug !~ '^[a-z0-9][a-z0-9-]{2,49}$'
        ORDER BY id
    LOOP
        base_slug := regexp_replace(lower(trim(current_space.slug)), '[^a-z0-9-]+', '-', 'g');
        base_slug := regexp_replace(base_slug, '^-+|-+$', '', 'g');
        IF base_slug = '' THEN
            base_slug := 'space';
        END IF;

        attempt := 0;
        LOOP
            suffix := '-' || current_space.id::text;
            IF attempt > 0 THEN
                suffix := suffix || '-' || attempt::text;
            END IF;
            candidate := regexp_replace(
                left(base_slug, greatest(1, 50 - char_length(suffix))),
                '-+$',
                ''
            ) || suffix;

            EXIT WHEN NOT EXISTS (
                SELECT 1
                FROM spaces other
                WHERE other.id <> current_space.id
                  AND other.deleted_at IS NULL
                  AND lower(other.slug) = lower(candidate)
            );
            attempt := attempt + 1;
        END LOOP;

        UPDATE spaces
        SET slug = candidate
        WHERE id = current_space.id;
    END LOOP;
END
$$;

UPDATE spaces
SET
    title = CASE
        WHEN char_length(trim(COALESCE(title, ''))) < 2 THEN 'Space ' || id::text
        ELSE left(trim(title), 80)
    END,
    short_description = CASE
        WHEN char_length(trim(COALESCE(short_description, ''))) BETWEEN 10 AND 240
            THEN trim(short_description)
        ELSE left('Project: ' || CASE
            WHEN char_length(trim(COALESCE(title, ''))) < 2 THEN 'Space ' || id::text
            ELSE left(trim(title), 80)
        END, 240)
    END,
    description = CASE
        WHEN description IS NULL THEN NULL
        ELSE left(description, 3000)
    END,
    why_it_matters = CASE
        WHEN why_it_matters IS NULL THEN NULL
        ELSE left(why_it_matters, 2000)
    END,
    status = CASE status
        WHEN 'draft' THEN 'paused'
        WHEN 'in_review' THEN 'paused'
        WHEN 'archived' THEN 'archived'
        ELSE 'active'
    END,
    visibility = CASE visibility
        WHEN 'public' THEN 'public'
        ELSE 'private'
    END,
    stage = CASE
        WHEN stage IN ('idea', 'prototype', 'mvp', 'beta', 'growth', 'scale') THEN stage
        ELSE 'idea'
    END,
    cover_icon = CASE
        WHEN cover_icon IN ('link', 'sprout', 'rocket', 'code', 'design', 'research', 'community', 'other') THEN cover_icon
        ELSE 'other'
    END,
    accent_color = CASE
        WHEN accent_color IN ('mint', 'blue', 'purple', 'orange', 'gray') THEN accent_color
        ELSE 'gray'
    END;

UPDATE spaces
SET archived_at = CASE
    WHEN status = 'archived' THEN COALESCE(archived_at, updated_at, created_at)
    ELSE NULL
END;

ALTER TABLE spaces
    ADD CONSTRAINT ck_spaces_slug_format CHECK (
        slug ~ '^[a-z0-9][a-z0-9-]{2,49}$'
    ),
    ADD CONSTRAINT ck_spaces_title_length CHECK (
        char_length(trim(title)) BETWEEN 2 AND 80
    ),
    ADD CONSTRAINT ck_spaces_short_description_length CHECK (
        short_description IS NOT NULL
        AND char_length(trim(short_description)) BETWEEN 10 AND 240
    ),
    ADD CONSTRAINT ck_spaces_description_length CHECK (
        description IS NULL OR char_length(description) <= 3000
    ),
    ADD CONSTRAINT ck_spaces_why_it_matters_length CHECK (
        why_it_matters IS NULL OR char_length(why_it_matters) <= 2000
    ),
    ADD CONSTRAINT ck_spaces_status CHECK (
        status IN ('active', 'paused', 'archived')
    ),
    ADD CONSTRAINT ck_spaces_visibility CHECK (
        visibility IN ('public', 'private')
    ),
    ADD CONSTRAINT ck_spaces_stage CHECK (
        stage IN ('idea', 'prototype', 'mvp', 'beta', 'growth', 'scale')
    ),
    ADD CONSTRAINT ck_spaces_cover_icon CHECK (
        cover_icon IN ('link', 'sprout', 'rocket', 'code', 'design', 'research', 'community', 'other')
    ),
    ADD CONSTRAINT ck_spaces_accent_color CHECK (
        accent_color IN ('mint', 'blue', 'purple', 'orange', 'gray')
    ),
    ADD CONSTRAINT ck_spaces_archive_state CHECK (
        (status = 'archived' AND archived_at IS NOT NULL)
        OR
        (status <> 'archived' AND archived_at IS NULL)
    );

ALTER INDEX IF EXISTS ux_spaces_slug_active
    RENAME TO ux_spaces_slug;

CREATE INDEX idx_spaces_visibility_status_created
    ON spaces (visibility, status, created_at DESC);

ALTER TABLE space_members
    ADD COLUMN title text,
    DROP CONSTRAINT IF EXISTS ck_space_members_role;

UPDATE space_members
SET role = 'member'
WHERE role = 'viewer';

-- Keep legacy memberships, but reserve the single active owner role for spaces.owner_id.
-- Any previous active owner remains an active admin instead of being deleted.
UPDATE space_members sm
SET role = 'admin'
FROM spaces s
WHERE sm.space_id = s.id
  AND sm.user_id <> s.owner_id
  AND sm.role = 'owner'
  AND sm.status = 'active';

INSERT INTO space_members (
    space_id,
    user_id,
    role,
    status,
    title,
    joined_at
)
SELECT
    s.id,
    s.owner_id,
    'owner',
    'active',
    NULL,
    now()
FROM spaces s
ON CONFLICT (space_id, user_id) DO UPDATE
SET
    role = 'owner',
    status = 'active',
    joined_at = COALESCE(space_members.joined_at, now());

ALTER TABLE space_members
    ADD CONSTRAINT ck_space_members_role CHECK (
        role IN ('owner', 'admin', 'member', 'contributor')
    ),
    ADD CONSTRAINT ck_space_members_title_length CHECK (
        title IS NULL OR char_length(trim(title)) BETWEEN 1 AND 80
    );

CREATE TABLE space_stack_items (
    id          bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    space_id    bigint NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT ck_space_stack_items_name_length CHECK (
        char_length(trim(name)) BETWEEN 1 AND 32
    )
);

CREATE UNIQUE INDEX ux_space_stack_items_space_name
    ON space_stack_items (space_id, lower(name));

CREATE INDEX idx_space_stack_items_space_id
    ON space_stack_items (space_id, id);

CREATE TABLE space_links (
    id          bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    space_id    bigint NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    type        text NOT NULL,
    title       text NOT NULL,
    url         text NOT NULL,
    position    integer NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT ck_space_links_type CHECK (
        type IN ('github', 'figma', 'notion', 'site', 'docs', 'other')
    ),
    CONSTRAINT ck_space_links_title_length CHECK (
        char_length(trim(title)) BETWEEN 1 AND 80
    ),
    CONSTRAINT ck_space_links_url_not_empty CHECK (
        char_length(trim(url)) > 0
    ),
    CONSTRAINT ck_space_links_position CHECK (
        position >= 0
    )
);

CREATE UNIQUE INDEX ux_space_links_space_position
    ON space_links (space_id, position);

CREATE INDEX idx_space_links_space_id
    ON space_links (space_id, position, id);

CREATE TRIGGER trg_space_links_touch_updated_at
BEFORE UPDATE ON space_links
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

COMMIT;
