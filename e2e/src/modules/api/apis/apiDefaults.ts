/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Default headers for all API clients.
 * Connection: close - Forces a new TCP connection for each request to prevent
 * "socket hang up" errors caused by the server closing idle keep-alive connections
 * before the client attempts to reuse them.
 */
export const API_DEFAULT_HEADERS = { Connection: 'close' };
