BEGIN;

DROP INDEX IF EXISTS idx_requests_fts;
DROP INDEX IF EXISTS idx_spaces_fts;
DROP INDEX IF EXISTS idx_profiles_fts;

COMMIT;
