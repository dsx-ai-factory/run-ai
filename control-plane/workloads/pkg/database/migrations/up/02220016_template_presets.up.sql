
-- Put scope type in front of the key, to allow searching by scope type system with no tenant id
DROP INDEX IF EXISTS crud_templates_by_name;
CREATE UNIQUE INDEX IF NOT EXISTS crud_templates_by_name ON workloads.crud_templates USING btree (scope_type, tenant_id, scope_id, name);

CREATE TABLE IF NOT EXISTS crud_templates_presets (
    "id" uuid NOT NULL,

    "preset_template_id" uuid NOT NULL,
    "tenant_id" int4 NOT NULL,
    "is_deleted" boolean NOT NULL,
    "tenant_template_id" uuid,

    PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS crud_template_presets_by_preset ON workloads.crud_templates_presets
    USING btree (preset_template_id, tenant_id);

CREATE UNIQUE INDEX IF NOT EXISTS crud_template_presets_by_tenant ON workloads.crud_templates_presets
    USING btree (tenant_id, tenant_template_id);
