BEGIN;

-- Full-text search indexes for MVP search.
-- They are expression indexes, so we do not store generated search_vector columns in the schema.

CREATE INDEX idx_profiles_fts
    ON profiles USING gin (
        to_tsvector(
            'simple',
            coalesce(full_name, '') || ' ' ||
            coalesce(headline, '') || ' ' ||
            coalesce(bio, '') || ' ' ||
            coalesce(organization, '') || ' ' ||
            coalesce(location, '')
        )
    )
    WHERE deleted_at IS NULL;

CREATE INDEX idx_spaces_fts
    ON spaces USING gin (
        to_tsvector(
            'simple',
            coalesce(name, '') || ' ' ||
            coalesce(short_description, '') || ' ' ||
            coalesce(description, '') || ' ' ||
            coalesce(type, '')
        )
    )
    WHERE deleted_at IS NULL;

CREATE INDEX idx_requests_fts
    ON requests USING gin (
        to_tsvector(
            'simple',
            coalesce(title, '') || ' ' ||
            coalesce(description, '') || ' ' ||
            coalesce(expected_commitment, '') || ' ' ||
            coalesce(type, '') || ' ' ||
            coalesce(format, '')
        )
    )
    WHERE deleted_at IS NULL;

COMMIT;
