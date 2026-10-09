SELECT id, space_id, name, created_at
FROM space_stack_items
WHERE space_id = $1
ORDER BY id;
