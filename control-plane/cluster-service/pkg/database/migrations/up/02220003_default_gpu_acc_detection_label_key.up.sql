UPDATE nodepools
SET gpu_network_acceleration_label_key = 'nvidia.com/gpu.clique'
WHERE name = 'default'
  AND gpu_network_acceleration_label_key = ''
  AND deleted_at IS NULL;
