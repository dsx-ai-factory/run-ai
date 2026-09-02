-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER  IF EXISTS sync_storage_classes_from_backend_trigger ON backend.storage_classes;
DROP TRIGGER  IF EXISTS sync_storage_classes_from_backend_trigger ON public.storage_classes;
DROP FUNCTION IF EXISTS sync_storage_classes_from_backend_func;