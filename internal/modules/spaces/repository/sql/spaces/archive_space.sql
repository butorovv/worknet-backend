UPDATE spaces
SET
    status = 'archived',
    archived_at = $2
WHERE slug = $1
  AND deleted_at IS NULL
  AND status <> 'archived'
RETURNING
    id,
    owner_id,
    slug,
    title,
    COALESCE(short_description, ''),
    COALESCE(description, ''),
    COALESCE(why_it_matters, ''),
    status,
    visibility,
    stage,
    cover_icon,
    accent_color,
    created_at,
    updated_at,
    archived_at;
