create table if not exists audit_logs (
    created_at    timestamp with time zone,
    deleted_at    timestamp with time zone,
    updated_at    timestamp with time zone,
    created_by    text,
    id            bigserial primary key,
    log_id       uuid not null,
    tenant_id     int not null,
    cluster_id   uuid not null,
    cluster_name text not null,
    request_id   uuid not null,
    timestamp     timestamp with time zone not null,
    subject       text not null,
    subject_type  text not null,
    source_ip     text not null,
    action        text not null,
    http_method   text not null,
    url           text not null,
    entity_type   text not null,
    entity_name   text not null,
    entity_id     text not null,
    http_status_code int not null,
    result        text not null,
    origin        text not null
);

create index audit_logs_tenant_time_index on audit_logs(tenant_id,timestamp);
create index audit_logs_tenant_entity_index on audit_logs(tenant_id,entity_type);
create index audit_logs_tenant_subject_index on audit_logs(tenant_id,subject);
