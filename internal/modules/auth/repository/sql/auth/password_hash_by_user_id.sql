SELECT password_hash
FROM auth_identities
WHERE user_id = $1
  AND provider = $2
