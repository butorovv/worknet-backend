INSERT INTO spaces (
    owner_id,
    slug,
    title,
    short_description,
    description,
    why_it_matters,
    status,
    visibility,
    stage,
    cover_icon,
    accent_color
)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9, $10, $11)
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
