-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE resources SET priority = 'VeryLow' where priority='Low';
UPDATE resources SET priority = 'Low' where priority='LowPlus';
UPDATE resources SET priority = 'MediumLow' where priority='Normal';
UPDATE resources SET priority = 'Medium' where priority='Expedite';
UPDATE resources SET priority = 'MediumHigh' where priority='High';
UPDATE resources SET priority = 'High' where priority='HighPlus';
UPDATE resources SET priority = 'Critical' where priority='VeryHigh';