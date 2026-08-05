SELECT count(*)
FROM requests r
WHERE r.space_id = $1
  AND r.deleted_at IS NULL
  AND r.status = $2
  AND (
      $5::boolean
      OR ($3::bigint IS NOT NULL AND r.owner_id = $3::bigint)
      OR (
          r.status IN ('open', 'closed')
          AND (
              r.visibility = 'public'
              OR (r.visibility = 'registered' AND $3::bigint IS NOT NULL)
              OR (r.visibility = 'members' AND $4::boolean)
          )
      )
  );
