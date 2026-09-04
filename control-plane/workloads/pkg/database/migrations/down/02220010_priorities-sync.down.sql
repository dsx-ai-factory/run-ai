-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS priorities_status_unique;
DO $$
    BEGIN
        IF EXISTS
            ( SELECT 1
              FROM   information_schema.tables
              WHERE  table_schema = 'workloads'
                AND  table_name = 'priorities'
            )
        THEN
            UPDATE priorities
            SET source = 'Run:ai'
            WHERE source = 'Run.ai';
        END IF ;
    END
$$ ;
