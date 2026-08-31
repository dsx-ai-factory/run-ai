-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_pvc ADD COLUMN IF NOT EXISTS data_sharing BOOLEAN DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS assets_pvc_datavolume (
    asset_id uuid primary key,
    datavolume_id uuid
);

CREATE OR REPLACE FUNCTION delete_pvc_datavolume_rows()
RETURNS TRIGGER AS $$
    BEGIN
        IF OLD.kind = 'pvc' THEN
            DELETE FROM assets_pvc_datavolume WHERE asset_id = OLD.id;
        END IF;
        RETURN OLD;
    END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS after_delete_pvc ON assets_metadata;
CREATE TRIGGER after_delete_pvc
    AFTER DELETE ON assets_metadata
    FOR EACH ROW
    EXECUTE FUNCTION delete_pvc_datavolume_rows();