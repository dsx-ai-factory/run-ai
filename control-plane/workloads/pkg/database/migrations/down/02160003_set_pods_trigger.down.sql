-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop trigger if exists set_pods_phases_update_time on pods;
drop function if exists update_pods_phases_update_time;