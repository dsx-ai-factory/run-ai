CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS policies (
    "id" UUID NOT NULL PRIMARY KEY,

    "name" text NOT NULL, CHECK (name <> ''),

    "workload_type" text NOT NULL, CHECK (workload_type <> ''),

    "scope" text NOT NULL, CHECK (scope <> ''),
    "tenant_id" integer,
    "department_id" text,
    "project_id" integer,

    "defaults" JSONB,
    "rules" JSONB,

    "created_by" text NOT NULL, CHECK (created_by <> ''),
    "created_at" timestamptz NOT NULL default NOW(),
    "updated_by" text NOT NULL, CHECK (updated_by <> ''),
    "updated_at" timestamptz NOT NULL default NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS by_name ON policies (tenant_id, workload_type, name);

DO $$
    BEGIN
        BEGIN
            CREATE TRIGGER set_timestamp
                BEFORE UPDATE ON policies
                FOR EACH ROW
            EXECUTE PROCEDURE trigger_set_timestamp();
        EXCEPTION
            WHEN duplicate_object THEN
                NULL;
        END;
    END $$;


CREATE TABLE IF NOT EXISTS effective_policies (
    "policy_id" UUID NOT NULL PRIMARY KEY,

    "defaults" JSONB,
    "rules" JSONB
);

