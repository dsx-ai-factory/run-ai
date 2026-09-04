-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop trigger if exists set_phases_update_time on workloads;
drop function if exists update_phases_update_time;