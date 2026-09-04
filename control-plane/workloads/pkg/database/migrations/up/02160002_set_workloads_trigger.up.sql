-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE OR REPLACE FUNCTION update_phases_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD is null OR NEW.phase <> OLD.phase THEN
        NEW.phase_updated_at = NOW();
        insert into workloads_phase_history (created_at, workload_id, phase, phase_message) values (NEW.phase_updated_at, NEW.id, NEW.phase, NEW.phase_message);
    ELSE
        NEW.phase_updated_at = OLD.phase_updated_at;
    END IF;

    IF OLD is null OR NEW.k8s_phase <> OLD.k8s_phase THEN
        NEW.k8s_phase_updated_at = NOW();
    ELSE
        NEW.k8s_phase_updated_at = OLD.k8s_phase_updated_at;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER set_phases_update_time
                BEFORE INSERT OR UPDATE
                ON workloads
                FOR EACH ROW
            EXECUTE FUNCTION update_phases_update_time();
    EXCEPTION
            WHEN duplicate_object THEN NULL;
    END;
END $$;
