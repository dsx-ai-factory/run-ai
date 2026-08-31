-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Kept out of the main migrations directory on purpose: this migration only
-- exists to make the migrator fail, so the failure-recording path can be tested.
SELECT * FROM table_that_does_not_exist;
