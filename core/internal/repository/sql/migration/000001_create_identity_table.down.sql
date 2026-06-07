DROP TABLE IF EXISTS identity;
DROP TABLE IF EXISTS session;
DROP TYPE IF EXISTS identity_type;
DROP INDEX IF EXISTS idx_identity_email;
DROP INDEX IF EXISTS idx_unique_session;
DROP INDEX IF EXISTS idx_session_identity_id;