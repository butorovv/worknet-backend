INSERT INTO users (username, email)
VALUES ($1, $2)
RETURNING id, username, email, status, created_at, updated_at
