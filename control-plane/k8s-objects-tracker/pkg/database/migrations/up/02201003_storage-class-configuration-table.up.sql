CREATE TABLE IF NOT EXISTS storage_class_configurations (
       id            UUID DEFAULT gen_random_uuid() PRIMARY KEY,
       tenant_id     INT NOT NULL,
       name          TEXT NOT NULL,
       scope_type    TEXT NOT NULL,
       scope_id      TEXT NOT NULL,
       allowed       BOOLEAN
);

CREATE UNIQUE INDEX IF NOT EXISTS storage_class_scope ON storage_class_configurations (tenant_id, scope_type, scope_id, name);