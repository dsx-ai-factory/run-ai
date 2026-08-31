-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE cluster_histories DROP CONSTRAINT cluster_histories_cluster_id_fkey;
ALTER TABLE cluster_histories ADD CONSTRAINT cluster_histories_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES clusters(id);
