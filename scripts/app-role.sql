\set ON_ERROR_STOP on
-- Run interactively as the existing database administrator AFTER migrations.
-- Re-running this script preserves the role and asks for a new password.
SELECT 'CREATE ROLE ferries_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname='ferries_app') \gexec
\password ferries_app
GRANT USAGE, CREATE ON SCHEMA public TO ferries_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ferries_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ferries_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ferries_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO ferries_app;
