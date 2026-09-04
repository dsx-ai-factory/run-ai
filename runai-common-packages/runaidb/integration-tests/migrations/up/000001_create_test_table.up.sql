-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE SCHEMA IF NOT EXISTS test_schema;

CREATE TABLE IF NOT EXISTS test_schema.test_secrets (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    secret TEXT NOT NULL
);

-- Schema the read user shouldn't have access to
CREATE SCHEMA IF NOT EXISTS restricted_schema;
CREATE TABLE IF NOT EXISTS restricted_schema.restricted_table (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

DO $$
DECLARE
    db_name text := current_database();
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'read-integdb') THEN
        CREATE USER "read-integdb" WITH PASSWORD 'read-password';
    END IF;

    EXECUTE format('GRANT CONNECT ON DATABASE %I TO %I', db_name, 'read-integdb');
END
$$;

GRANT USAGE ON SCHEMA test_schema TO "read-integdb";
GRANT SELECT ON ALL TABLES IN SCHEMA test_schema TO "read-integdb";
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA test_schema TO "read-integdb";

ALTER DEFAULT PRIVILEGES IN SCHEMA test_schema GRANT SELECT ON TABLES TO "read-integdb";
ALTER DEFAULT PRIVILEGES IN SCHEMA test_schema GRANT USAGE, SELECT ON SEQUENCES TO "read-integdb";
