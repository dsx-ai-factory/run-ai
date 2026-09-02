-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE resources SET priority = 'LowPlus' where priority='Low';
UPDATE resources SET priority = 'Low' where priority='VeryLow';
UPDATE resources SET priority = 'Normal' where priority='MediumLow';
UPDATE resources SET priority = 'Expedite' where priority='Medium';
UPDATE resources SET priority = 'HighPlus' where priority='High';
UPDATE resources SET priority = 'High' where priority='MediumHigh';
UPDATE resources SET priority = 'VeryHigh' where priority='Critical';
