ALTER TABLE workloads ADD COLUMN IF NOT EXISTS resource_refs JSONB;

DROP VIEW IF EXISTS applications_view;
CREATE VIEW applications_view AS (SELECT DISTINCT ON (a.id)
    a.*,
    cs.spec,
    cs.desired_phase,
    cs.update_count,
    sum(p.requested_gpu_limit) as requested_gpu_limit,
    sum(p.requested_gpu_request) as requested_gpu_request,
    array_agg(p.requested_gpu_memory_limit) FILTER (WHERE p.requested_gpu_memory_limit IS NOT NULL) as requested_gpu_memory_limit,
    array_agg(p.requested_gpu_memory_request) FILTER (WHERE p.requested_gpu_memory_request IS NOT NULL) as requested_gpu_memory_request,
    sum(p.requested_cpu_limit) as requested_cpu_limit,
    sum(p.requested_cpu_request) as requested_cpu_request,
    array_agg(p.requested_cpu_memory_limit) FILTER (WHERE p.requested_cpu_memory_limit IS NOT NULL) as requested_cpu_memory_limit,
    array_agg(p.requested_cpu_memory_request) FILTER (WHERE p.requested_cpu_memory_request IS NOT NULL) as requested_cpu_memory_request,
    sum(p.allocated_gpu) as allocated_gpu,
    array_agg(p.allocated_gpu_memory) FILTER (WHERE p.allocated_gpu_memory IS NOT NULL) as allocated_gpu_memory,
    sum(p.allocated_cpu) as allocated_cpu,
    array_agg(p.allocated_cpu_memory) FILTER (WHERE p.allocated_cpu_memory IS NOT NULL) as allocated_cpu_memory,
    json_agg(json_build_object('workloadId', w.id, 'endpoints', w.endpoints)) FILTER (WHERE w.endpoints IS NOT NULL) as endpoints,
    (
        SELECT json_agg(json_build_object('workloadId', w2.id, 'resourceRefs', w2.resource_refs))
        FROM workloads w2
        WHERE w2.application_id = a.id AND w2.deleted_at IS NULL AND w2.resource_refs IS NOT NULL
    ) as resource_connections
 FROM applications as a
     LEFT JOIN application_crud_sync as cs
 ON a.id = cs.application_id
     LEFT JOIN workloads as w
     ON a.id = w.application_id and w.deleted_at is null
     LEFT JOIN pods as p
     ON p.workload_id = w.id and p.deleted_at IS NULL
 GROUP BY a.id, cs.spec, cs.desired_phase, cs.update_count
     );
