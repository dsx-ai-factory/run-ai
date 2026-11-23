/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DatabaseConnection } from './DatabaseConnection.js';

export { DatabaseConnection } from './DatabaseConnection.js';
export type { DatabaseConfig, QueryResult } from './DatabaseConnection.js';

export const db = new DatabaseConnection();
