ALTER TABLE crud_workloads ADD COLUMN IF NOT EXISTS department_id text NOT NULL;

CREATE INDEX IF NOT EXISTS crud_workloads_by_project ON crud_workloads (tenant_id, project_id);
CREATE INDEX IF NOT EXISTS crud_workloads_by_department ON crud_workloads (tenant_id, department_id);
