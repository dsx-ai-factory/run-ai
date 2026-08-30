create table if not exists workloads (
     created_at    timestamp with time zone not null default now(),
     updated_at    timestamp with time zone not null default now(),
     deleted_at    timestamp with time zone,
     submitted_by    varchar not null default '',
     id            uuid primary key,
     name          varchar not null,
     source    varchar not null,
     priority_class_name   varchar,
     type   varchar not null,
     tenant_id   bigint not null,
     cluster_id   uuid not null,
     project_name   varchar not null default '',
     project_id   bigint not null,
     namespace   varchar not null,
     requested_gpu_limit   NUMERIC(1000, 2),
     requested_gpu_request   NUMERIC(1000, 2),
     requested_gpu_memory_limit   varchar,
     requested_gpu_memory_request   varchar,
     requested_cpu_limit   NUMERIC(1000, 2),
     requested_cpu_request   NUMERIC(1000, 2),
     requested_cpu_memory_limit   varchar,
     requested_cpu_memory_request   varchar,
     requested_mig_device   varchar array,
     support_delete boolean not null default false,
     support_suspend boolean not null default false,
     phase   varchar not null,
     phase_message   varchar not null default '',
     phase_updated_at   timestamp with time zone not null,
     k8s_phase   varchar not null,
     k8s_phase_updated_at   timestamp with time zone not null,
     requested_pods_number   bigint,
     requested_pods_min   bigint,
     requested_pods_max   bigint,
     requested_pods_parallelism   bigint,
     requested_pods_completions   bigint,
     requested_nodepools   varchar array,
     images   varchar array,
     conditions jsonb default null,
     datasources jsonb default null,
     environments jsonb default null,
     external_connections jsonb default null,
     distributed_framework varchar,
     urls   varchar array not null default array[]::varchar[],
     additional_fields jsonb not null default '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_workloads_tenant ON workloads (tenant_id);
CREATE INDEX IF NOT EXISTS idx_workloads_cluster ON workloads (cluster_id);
CREATE INDEX IF NOT EXISTS idx_workloads_project ON workloads (project_id);


create table if not exists workloads_phase_history (
                                                        created_at    timestamp with time zone,
                                                        workload_id uuid,
                                                        foreign key (workload_id) REFERENCES workloads(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
                                                        phase          varchar not null,
                                                        phase_message   varchar
);

create table if not exists workloads_children (
                                                  workload_id uuid,
                                                  foreign key (workload_id) REFERENCES workloads(id) ON DELETE CASCADE,
                                                  child_type          text,
                                                  child_id          uuid not null,
                                                      CONSTRAINT workload_children_unique_constraint UNIQUE (workload_id, child_id)
);

create table if not exists pods (
                                    created_at    timestamp with time zone,
                                    updated_at    timestamp with time zone,
                                    completed_at    timestamp with time zone,
                                    deleted_at    timestamp with time zone,
                                    id            uuid primary key,
                                    workload_id uuid,
                                    name          varchar not null,
                                    priority_class_name   varchar,
                                    tenant_id   bigint not null,
                                    cluster_id   uuid not null,
                                    project_id   bigint not null,
                                    node_name   varchar,
                                    current_nodepool   varchar,
                                    requested_nodepools   varchar array,
                                    requested_gpu_limit   NUMERIC(1000, 2),
                                    requested_gpu_request   NUMERIC(1000, 2),
                                    requested_gpu_memory_limit   varchar,
                                    requested_gpu_memory_request   varchar,
                                    requested_cpu_limit   NUMERIC(1000, 2),
                                    requested_cpu_request   NUMERIC(1000, 2),
                                    requested_cpu_memory_limit   varchar,
                                    requested_cpu_memory_request   varchar,
                                    requested_mig_device   varchar,
                                    allocated_gpu   NUMERIC(1000, 2),
                                    allocated_gpu_memory   varchar,
                                    allocated_cpu   NUMERIC(1000, 2),
                                    allocated_cpu_memory   varchar,
                                    allocated_mig_device   varchar,
                                    tolerations jsonb default null,
                                    phase   varchar not null,
                                    phase_updated_at   timestamp with time zone
);

CREATE INDEX IF NOT EXISTS idx_pods_tenant ON pods (tenant_id);
CREATE INDEX IF NOT EXISTS idx_pods_cluster ON pods (cluster_id);
CREATE INDEX IF NOT EXISTS idx_pods_project ON pods (project_id);
CREATE INDEX IF NOT EXISTS idx_pods_workload ON pods (workload_id);

create table if not exists containers (
                                          started_at    timestamp with time zone,
                                          cluster_id   uuid not null,
                                          pod_id uuid,
                                          foreign key (pod_id) REFERENCES pods(id) ON DELETE CASCADE,
                                          name          varchar not null,
                                          image   varchar not null
);

create table if not exists events (
                                      created_at    timestamp with time zone,
                                      id            uuid primary key,
                                      cluster_id   uuid not null,
                                      tenant_id   bigint not null,
                                      type          varchar not null,
                                      message   varchar,
                                      reason   varchar,
                                      source   varchar,
                                      involved_object_uid   uuid,
                                      involved_object_kind   varchar,
                                      involved_object_name   varchar,
                                      involved_object_namespace   varchar
);

drop view if exists workloads_view;
create view workloads_view AS (SELECT DISTINCT ON (w.id)
                                   w.created_at,
                                   w.updated_at,
                                   w.deleted_at,
                                   w.submitted_by,
                                   w.id,
                                   w.name,
                                   w.source,
                                   w.type,
                                   w.tenant_id,
                                   w.cluster_id,
                                   w.project_name,
                                   w.project_id,
                                   w.namespace,
                                   w.requested_gpu_limit as workload_requested_gpu_limit,
                                   w.requested_gpu_request as workload_requested_gpu_request,
                                   string_to_array(w.requested_gpu_memory_limit, '') as workload_requested_gpu_memory_limit,
                                   string_to_array(w.requested_gpu_memory_request, '') as workload_requested_gpu_memory_request,
                                   w.requested_cpu_limit as workload_requested_cpu_limit,
                                   w.requested_cpu_request as workload_requested_cpu_request,
                                   string_to_array(w.requested_cpu_memory_limit, '') as workload_requested_cpu_memory_limit,
                                   string_to_array(w.requested_cpu_memory_request, '') as workload_requested_cpu_memory_request,
                                   w.requested_mig_device as workload_requested_mig_device,
                                   p.requested_gpu_limit as pods_requested_gpu_limit,
                                   p.requested_gpu_request as pods_requested_gpu_request,
                                   p.requested_gpu_memory_limit as pods_requested_gpu_memory_limit,
                                   p.requested_gpu_memory_request as pods_requested_gpu_memory_request,
                                   p.requested_cpu_limit as pods_requested_cpu_limit,
                                   p.requested_cpu_request as pods_requested_cpu_request,
                                   p.requested_cpu_memory_limit as pods_requested_cpu_memory_limit,
                                   p.requested_cpu_memory_request as pods_requested_cpu_memory_request,
                                   p.requested_mig_devices as pods_requested_mig_device,
                                   w.support_delete,
                                   w.support_suspend,
                                   w.phase,
                                   w.phase_message,
                                   w.phase_updated_at,
                                   w.k8s_phase,
                                   w.k8s_phase_updated_at,
                                   w.requested_pods_number,
                                   w.requested_pods_min,
                                   w.requested_pods_max,
                                   w.requested_pods_parallelism,
                                   w.requested_pods_completions,
                                   coalesce(w.requested_nodepools, prnp.requested_nodepools) as requested_nodepools,
                                   w.priority_class_name,
                                   w.images,
                                   w.urls,
                                   w.conditions,
                                   w.datasources,
                                   w.environments,
                                   w.external_connections,
                                   w.distributed_framework,
                                   w.additional_fields,

                                   p.number_of_running_pods,
                                   (case w.phase when 'Completed' then p.latest_pod_completion_time else null end) as latest_pod_completion_time,
                                   p.current_nodepools,
                                   p.nodes,
                                   p.allocated_gpu,
                                   p.allocated_gpu_memory,
                                   p.allocated_cpu,
                                   p.allocated_cpu_memory,
                                   p.allocated_mig_devices
                              FROM
                                  workloads as w

                                      LEFT JOIN
                                  (
                                      SELECT
                                          workload_id,
                                          count(case phase when 'Running' then 1 else null end) as number_of_running_pods,
                                          max(completed_at) as latest_pod_completion_time,
                                          array_agg(DISTINCT current_nodepool) FILTER (WHERE current_nodepool IS NOT NULL)  as current_nodepools,
                                          array_agg(DISTINCT node_name) FILTER (WHERE node_name IS NOT NULL) as nodes,
                                          case WHEN -1 = ANY(array_agg(requested_gpu_limit)) THEN -1
                                               ELSE sum(requested_gpu_limit) end as requested_gpu_limit,
                                          case WHEN -1 = ANY(array_agg(requested_gpu_request)) THEN -1
                                               ELSE sum(requested_gpu_request) end as requested_gpu_request,
                                          array_agg(requested_gpu_memory_limit) FILTER (WHERE requested_gpu_memory_limit IS NOT NULL) as requested_gpu_memory_limit,
                                          array_agg(requested_gpu_memory_request) FILTER (WHERE requested_gpu_memory_request IS NOT NULL) as requested_gpu_memory_request,
                                          case WHEN -1 = ANY(array_agg(requested_cpu_limit)) THEN -1
                                               ELSE sum(requested_cpu_limit) end as requested_cpu_limit,
                                          case WHEN -1 = ANY(array_agg(requested_cpu_request)) THEN -1
                                               ELSE sum(requested_cpu_request) end as requested_cpu_request,                                                 array_agg(requested_cpu_memory_limit)  FILTER (WHERE requested_cpu_memory_limit IS NOT NULL) as requested_cpu_memory_limit,
                                          array_agg(requested_cpu_memory_request) FILTER (WHERE requested_cpu_memory_request IS NOT NULL) as requested_cpu_memory_request,
                                          array_agg(requested_mig_device) FILTER (WHERE requested_mig_device IS NOT NULL) as requested_mig_devices,

                                          sum(case when completed_at is not null then null else allocated_gpu end) as allocated_gpu,
                                          array_agg(allocated_gpu_memory) FILTER (WHERE allocated_gpu_memory IS NOT NULL) as allocated_gpu_memory,
                                          sum(case when completed_at is not null then null else allocated_cpu end) as allocated_cpu,
                                          array_agg(allocated_cpu_memory) FILTER (WHERE allocated_cpu_memory IS NOT NULL) as allocated_cpu_memory,
                                          array_agg(allocated_mig_device) FILTER (WHERE allocated_mig_device IS NOT NULL) as allocated_mig_devices
                                      FROM
                                          pods
                                      WHERE
                                        deleted_at IS NULL
                                      GROUP BY workload_id
                                  ) as p
                                  ON p.workload_id = w.id

                                      LEFT JOIN
                                  (
                                      SELECT
                                          workload_id,
                                          array_agg(DISTINCT nodepool) AS requested_nodepools
                                      FROM
                                          pods,
                                          unnest(requested_nodepools) AS t(nodepool)
                                      GROUP BY workload_id
                                  ) as prnp
                                  ON prnp.workload_id = w.id
                                  );