SELECT
    r.id,
    r.space_id,
    r.owner_id,
    r.type,
    r.status,
    r.visibility,
    r.title,
    r.description,
    r.expected_commitment,
    r.format,
    COALESCE(response_counts.count, 0)::integer AS responses_count,
    r.published_at,
    r.closed_at,
    r.created_at,
    r.updated_at
FROM requests r
LEFT JOIN (
    SELECT rr.request_id, count(*) AS count
    FROM request_responses rr
    JOIN requests scoped_request ON scoped_request.id = rr.request_id
    WHERE scoped_request.space_id = $1
    GROUP BY rr.request_id
) response_counts ON response_counts.request_id = r.id
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
  )
ORDER BY r.created_at DESC, r.id DESC
LIMIT $6 OFFSET $7;
