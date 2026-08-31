-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- create a trigger to sync any change
CREATE OR REPLACE FUNCTION k8s_objects_tracker.sync_storage_classes_from_backend_func()
        RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
           INSERT INTO k8s_objects_tracker.objects (
                type,
                scope_type,
                name,
                cluster_id,
                tenant_id,
                spec
            ) VALUES (
                'StorageClass',
                'cluster',
                NEW.storage_class_name,
                NEW.cluster_id,
                NEW.tenant_id,
                jsonb_build_object('is_default', NEW.is_default)
            )
        ON CONFLICT (tenant_id, type, cluster_id, name, project_id)
            DO UPDATE SET
                spec = EXCLUDED.spec;

    ELSIF TG_OP = 'UPDATE' THEN
        -- The constraint on backend.storage_classes is : UNIQUE (tenant_id, cluster_id, storage_class_name)
        UPDATE k8s_objects_tracker.objects SET
           spec = jsonb_build_object('is_default', NEW.is_default)
        WHERE type = 'StorageClass' AND
            tenant_id = NEW.tenant_id AND
            cluster_id = NEW.cluster_id AND
            name = NEW.storage_class_name;

    ELSIF TG_OP = 'DELETE' THEN
    DELETE FROM k8s_objects_tracker.objects
    WHERE type = 'StorageClass' AND
            OLD.tenant_id = tenant_id   AND
            OLD.cluster_id = cluster_id   AND
            OLD.storage_class_name = name;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;



-- migrate data from backend.storage_classes table
-- copy the data, then setup a trigger to sync them all.
-- remove the trigger once we delete backend.storage_classes tbl
DO $$
BEGIN
    -- create the trigger only if backend.storage_classes exists
    IF EXISTS (
        SELECT 1
        FROM pg_tables
        WHERE schemaname = 'backend'
        AND tablename = 'storage_classes'
    ) THEN

        -- copy all records
        INSERT INTO k8s_objects_tracker.objects (
                type,
                scope_type,
                name,
                cluster_id,
                tenant_id,
                spec
            )
        SELECT
            'StorageClass',
            'cluster',
            storage_class_name,
            cluster_id,
            tenant_id,
            jsonb_build_object('is_default', is_default)
        FROM
            backend.storage_classes
        ON CONFLICT (tenant_id, type, cluster_id, name, project_id)
            DO NOTHING;

        -- drop before create, since there is no 'CREATE TRIGGER IF NOT EXISTS'
        DROP TRIGGER IF EXISTS sync_storage_classes_from_backend_trigger ON backend.storage_classes;
        CREATE TRIGGER sync_storage_classes_from_backend_trigger
            BEFORE INSERT OR UPDATE OR DELETE ON backend.storage_classes
            FOR EACH ROW
            EXECUTE FUNCTION k8s_objects_tracker.sync_storage_classes_from_backend_func();

    ELSIF EXISTS (
        SELECT 1
        FROM pg_tables
        WHERE schemaname = 'public'
        AND tablename = 'storage_classes'
    ) THEN

        -- copy all records
        INSERT INTO k8s_objects_tracker.objects (
                type,
                scope_type,
                name,
                cluster_id,
                tenant_id,
                spec
            )
        SELECT
            'StorageClass',
            'cluster',
            storage_class_name,
            cluster_id,
            tenant_id,
            jsonb_build_object('is_default', is_default)
        FROM
            public.storage_classes
            ON CONFLICT (tenant_id, type, cluster_id, name, project_id)
                    DO NOTHING;

        -- drop before create, since there is no 'CREATE TRIGGER IF NOT EXISTS'
        DROP TRIGGER IF EXISTS sync_storage_classes_from_backend_trigger ON public.storage_classes;
        CREATE TRIGGER sync_storage_classes_from_backend_trigger
            BEFORE INSERT OR UPDATE OR DELETE ON public.storage_classes
            FOR EACH ROW
            EXECUTE FUNCTION k8s_objects_tracker.sync_storage_classes_from_backend_func();

    END IF;

END $$ LANGUAGE plpgsql;

