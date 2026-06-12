SELECT
    s.id,
    s.owner_id,
    s.slug,
    s.title,
    COALESCE(s.short_description, ''),
    COALESCE(s.description, ''),
    COALESCE(s.why_it_matters, ''),
    s.status,
    s.visibility,
    s.stage,
    s.cover_icon,
    s.accent_color,
    s.created_at,
    s.updated_at,
    s.archived_at
FROM spaces s
WHERE s.deleted_at IS NULL
  AND s.status = $1
  AND s.visibility = $2
  AND (
      s.visibility = 'public'
      OR (
          $3::bigint IS NOT NULL
          AND EXISTS (
              SELECT 1
              FROM space_members sm
              WHERE sm.space_id = s.id
                AND sm.user_id = $3
                AND sm.status = 'active'
          )
      )
  )
ORDER BY s.created_at DESC, s.id DESC
LIMIT $4 OFFSET $5;
