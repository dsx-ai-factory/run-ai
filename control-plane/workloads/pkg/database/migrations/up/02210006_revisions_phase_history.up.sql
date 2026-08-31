-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS revisions_phase_history (
                                                       created_at timestamp with time zone,
                                                       tenant_id bigint,
                                                       cluster_id uuid,
                                                       revision_id uuid,
                                                       FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    phase varchar NOT NULL,
    phase_message varchar,
    old_phase varchar DEFAULT ''
    );

CREATE INDEX IF NOT EXISTS idx_phase_create_time_revision ON revisions_phase_history (created_at desc);
CREATE INDEX IF NOT EXISTS idx_phase_revision ON revisions_phase_history (revision_id);
CREATE INDEX IF NOT EXISTS idx_phase_history_tenant_revision ON revisions_phase_history (tenant_id, revision_id);

CREATE OR REPLACE FUNCTION update_revision_phases_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD IS NULL OR OLD.phase IS NULL OR (NEW.phase <> OLD.phase AND NEW.phase <> 'Unknown') THEN
        NEW.phase_updated_at = NOW();
        INSERT INTO revisions_phase_history (created_at, tenant_id, cluster_id, revision_id, phase, phase_message, old_phase) VALUES (NEW.phase_updated_at, NEW.tenant_id, NEW.cluster_id, NEW.id, NEW.phase, NEW.phase_message, OLD.phase);
    ELSE
        IF NEW.phase = 'Unknown' THEN
            NEW.phase = OLD.phase;
        END IF;
        NEW.phase_updated_at = OLD.phase_updated_at;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER set_revision_phases_update_time
            BEFORE INSERT OR UPDATE
                ON revisions
                FOR EACH ROW
                EXECUTE FUNCTION update_revision_phases_update_time();
        EXCEPTION
            WHEN duplicate_object THEN NULL;
    END;
END $$;