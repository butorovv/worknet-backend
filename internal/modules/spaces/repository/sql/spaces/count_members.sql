SELECT count(*)
FROM space_members
WHERE space_id = $1
  AND status = 'active';
