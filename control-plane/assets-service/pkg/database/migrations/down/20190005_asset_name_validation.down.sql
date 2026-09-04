-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


--- Revert the function to how it was prior adding the support for system assets

CREATE OR REPLACE FUNCTION asset_name_validation()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.name IN (select name FROM assets_metadata WHERE NEW.kind = assets_metadata.kind AND
        NEW.tenant_id = assets_metadata.tenant_id AND NEW.scope != assets_metadata.scope)
    THEN
        RAISE EXCEPTION 'non-% % asset named % already exists', NEW.scope, NEW.kind, NEW.name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
