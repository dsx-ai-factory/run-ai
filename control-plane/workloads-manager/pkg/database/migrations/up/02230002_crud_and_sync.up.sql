-- we used this sequence to mark every update on the table with increasing number
-- which will be used as the timestamp for the sync.
CREATE SEQUENCE IF NOT EXISTS workloads_queue_sequence;

CREATE TABLE IF NOT EXISTS workloads_queue (
      "id" uuid NOT NULL,

      "created_by" text NOT NULL,
      "created_at" timestamptz NOT NULL,
      "updated_by" text NOT NULL,
      "updated_at" timestamptz NOT NULL,
      "deleted_by" text,
      "deleted_at" timestamptz,

      "tenant_id" int4 NOT NULL,
      "cluster_id" uuid NOT NULL,
      "department_id" text NOT NULL,
      "department_name" text NOT NULL,
      "project_id" text NOT NULL,
      "project_name" text NOT NULL,

      "name" text NOT NULL,

      "category" text,
      "priority_class" text,

      "resource_group" text NOT NULL,
      "resource_version" text NOT NULL,
      "resource_kind" text NOT NULL,

      "raw_manifest" jsonb NOT NULL,
      "effective_manifest" jsonb NOT NULL,

      "desired_phase" text NOT NULL,

      "update_count" int8 NOT NULL,

      PRIMARY KEY ("id")
);

-- Indices
CREATE INDEX IF NOT EXISTS workloads_queue_for_sync ON workloads_queue USING btree (tenant_id, cluster_id, update_count);
CREATE INDEX IF NOT EXISTS workloads_queue_by_project ON workloads_queue USING btree (tenant_id, project_id);

CREATE OR REPLACE FUNCTION workloads_queue_update_count()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.update_count := nextval('workloads_queue_sequence');
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
BEGIN
CREATE TRIGGER workloads_queue_update_count_trigger BEFORE INSERT OR UPDATE ON workloads_queue
                                                                        FOR EACH ROW
                                                                        EXECUTE PROCEDURE workloads_queue_update_count();
EXCEPTION
            WHEN duplicate_object THEN NULL;
END;
END $$;
