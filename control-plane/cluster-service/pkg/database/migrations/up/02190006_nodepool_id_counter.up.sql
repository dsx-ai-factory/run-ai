-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

SELECT setval(pg_get_serial_sequence('nodepools', 'id'), 100000, false);