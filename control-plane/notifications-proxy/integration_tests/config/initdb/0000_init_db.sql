-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create user postgres  SUPERUSER;

CREATE DATABASE runai WITH
    OWNER postgres
    ENCODING = 'UTF8';