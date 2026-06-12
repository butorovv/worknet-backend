UPDATE spaces
SET
    title = $2,
    short_description = $3,
    description = NULLIF($4, ''),
    why_it_matters = NULLIF($5, ''),
    status = $6,
    visibility = $7,
    stage = $8,
    cover_icon = $9,
    accent_color = $10
WHERE id = $1
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
