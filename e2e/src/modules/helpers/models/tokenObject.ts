/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export class TokenObject {
    constructor(
        public readonly idToken?: string,
        public readonly refreshToken?: string,
        public readonly externalToken?: string,
        public readonly accessToken?: string
    ) {}
}
