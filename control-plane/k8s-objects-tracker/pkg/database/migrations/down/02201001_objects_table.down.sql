ALTER TABLE IF EXISTS  k8s_objects_tracker.objects DROP CONSTRAINT IF EXISTS unique_tenant_type_cluster_name_project;
DROP TABLE IF EXISTS  k8s_objects_tracker.objects;
