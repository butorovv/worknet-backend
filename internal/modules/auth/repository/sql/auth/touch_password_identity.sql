UPDATE auth_identities
SET last_used_at = $2,
    updated_at = now()
WHERE user_id = $1
  AND provider = $3
