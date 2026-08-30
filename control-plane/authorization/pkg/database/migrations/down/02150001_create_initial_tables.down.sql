drop index if exists idx_subject_role_resource;
drop index if exists idx_access_rules_tenant_id;
drop index if exists idx_access_rules_cluster_uuid;
drop index if exists idx_access_rules_deleted_at;
drop table if exists access_rules;
drop index if exists idx_permissions_resource_type;
drop table if exists permissions;
drop table if exists roles;