SELECT
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
    archived_at
FROM spaces
WHERE id = $1
  AND deleted_at IS NULL;
