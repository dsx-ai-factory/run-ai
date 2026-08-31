-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create index if not exists container_pod_id_idx on containers(pod_id);
create index if not exists workloads_children_id_idx on workloads_children(workload_id);