DELETE FROM workload_type_resource_interfaces
WHERE workload_type_id IN (
    SELECT id FROM workload_types
    WHERE (name, "group") IN (
        ('NIMCache', 'apps.nvidia.com'),
        ('NIMService', 'apps.nvidia.com'),
        ('SPOTRequest', 'egx.nvidia.io')
    )
);

DELETE FROM external_workloads_type_config
WHERE workload_type_id IN (
    SELECT id FROM workload_types
    WHERE (name, "group") IN (
        ('NIMCache', 'apps.nvidia.com'),
        ('NIMService', 'apps.nvidia.com'),
        ('SPOTRequest', 'egx.nvidia.io')
    )
);

DELETE FROM workload_types
WHERE (name, "group") IN (
    ('NIMCache', 'apps.nvidia.com'),
    ('NIMService', 'apps.nvidia.com'),
    ('SPOTRequest', 'egx.nvidia.io')
);
