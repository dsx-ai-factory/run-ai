-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create user postgres  SUPERUSER;

CREATE DATABASE runai WITH
    OWNER postgres
    ENCODING = 'UTF8';


\connect runai

CREATE TABLE IF NOT EXISTS public."UserEvents" (
    id text NOT NULL,
    tenant_id text NOT NULL,
    user_id text NOT NULL,
    listener_id text NOT NULL,
    properties jsonb,
    description text NOT NULL,
    is_read boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NULL,
    CONSTRAINT user_events_id_unique UNIQUE (id),
    PRIMARY KEY (id, tenant_id, user_id, listener_id)
);

ALTER TABLE public."UserEvents" OWNER TO runai;
GRANT ALL ON TABLE public."UserEvents" TO runai;