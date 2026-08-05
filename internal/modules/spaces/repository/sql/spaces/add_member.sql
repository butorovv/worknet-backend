INSERT INTO space_members (
    space_id,
    user_id,
    role,
    status,
    title,
    joined_at
)
VALUES ($1, $2, $3, 'active', $4, now())
ON CONFLICT (space_id, user_id) DO UPDATE
SET
    role = EXCLUDED.role,
    status = 'active',
    title = EXCLUDED.title,
    joined_at = COALESCE(space_members.joined_at, now());
