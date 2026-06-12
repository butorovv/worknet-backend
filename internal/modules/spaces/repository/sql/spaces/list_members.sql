SELECT
    sm.space_id,
    sm.user_id,
    u.username,
    sm.role,
    sm.title,
    sm.created_at
FROM space_members sm
JOIN users u ON u.id = sm.user_id
WHERE sm.space_id = $1
  AND sm.status = 'active'
  AND u.deleted_at IS NULL
ORDER BY
    CASE sm.role
        WHEN 'owner' THEN 0
        WHEN 'admin' THEN 1
        WHEN 'member' THEN 2
        ELSE 3
    END,
    sm.created_at,
    sm.user_id;
