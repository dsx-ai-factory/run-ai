DROP VIEW IF EXISTS applications_view;
DROP INDEX IF EXISTS applications_search_vector_idx;
ALTER TABLE applications DROP COLUMN IF EXISTS search_vector;

CREATE VIEW applications_view AS (SELECT DISTINCT ON (a.id)
                                      a.*,
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
                                      array_agg(p.allocated_cpu_memory) FILTER (WHERE p.allocated_cpu_memory IS NOT NULL) as allocated_cpu_memory
                                 FROM applications as a
                                          LEFT JOIN workloads as w
                                                    ON a.id = w.application_id and w.deleted_at is null
                                          LEFT JOIN pods as p
                                                    ON p.workload_id = w.id and p.deleted_at IS NULL
                                 GROUP BY a.id
                                     );
