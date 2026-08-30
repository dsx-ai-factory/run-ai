alter table workloads add column if not exists department_id varchar;
alter table workloads add column if not exists department_name varchar;

CREATE INDEX IF NOT EXISTS idx_workloads_department_id ON workloads (department_id) where department_id <> '';
CREATE INDEX IF NOT EXISTS idx_workloads_department_name ON workloads (department_name) where department_name <> '';

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
--                                   those added
                                   w.department_name,
                                   w.department_id,
--
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