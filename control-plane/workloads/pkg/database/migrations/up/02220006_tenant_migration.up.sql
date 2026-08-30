CREATE TABLE IF NOT EXISTS crud_migrations (
    "tenant_id" int4 NOT NULL,
    "assets_to_flat_templates" int,
    PRIMARY KEY ("tenant_id")
);
