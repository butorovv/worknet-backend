SELECT id, space_id, type, title, url, position, created_at, updated_at
FROM space_links
WHERE space_id = $1
ORDER BY position, id;
