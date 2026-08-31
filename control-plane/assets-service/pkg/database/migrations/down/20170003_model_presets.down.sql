-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- order matters, first the containing assets, then the contained ones, then the metadata

DELETE FROM assets_supported_workload_types WHERE asset_id in
    (select id FROM assets_metadata WHERE kind = 'environment' and scope = 'system' and name in
        ('falcon-7b-instruct',
         'falcon-40b-instruct',
         'llama-2-7b-chat-hf',
         'llama-2-13b-chat-hf'));

DELETE FROM assets_model WHERE asset_id in
   (select id FROM assets_metadata WHERE kind = 'model' and scope = 'system' and name in
       ('falcon-7b-instruct',
        'falcon-40b-instruct',
        'llama-2-7b-chat-hf',
        'llama-2-13b-chat-hf'));

DELETE FROM assets_compute WHERE asset_id in
   (select id FROM assets_metadata WHERE kind = 'compute' and scope = 'system' and name in
       ('gpu-memory-16g',
        'gpu-memory-20g',
        'gpu-memory-40g',
        'gpu-memory-80g'));

DELETE FROM assets_environment WHERE asset_id in
  (select id FROM assets_metadata WHERE kind = 'environment' and scope = 'system' and name in
      ('falcon-7b-instruct',
       'falcon-40b-instruct',
       'llama-2-7b-chat-hf',
       'llama-2-13b-chat-hf'));

DELETE FROM assets_metadata WHERE scope = 'system' and (
    (kind = 'compute' and name in
        ('gpu-memory-16g',
         'gpu-memory-20g',
         'gpu-memory-40g',
         'gpu-memory-80g')) or
    (kind = 'environment' and name in
        ('falcon-7b-instruct',
         'falcon-40b-instruct',
         'llama-2-7b-chat-hf',
         'llama-2-13b-chat-hf')) or
    (kind = 'model' and name in
        ('falcon-7b-instruct',
         'falcon-40b-instruct',
         'llama-2-7b-chat-hf',
         'llama-2-13b-chat-hf'))
    );
