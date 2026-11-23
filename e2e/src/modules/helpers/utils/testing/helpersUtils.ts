/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';

export function validatePropertyIfNotNull(
    actual: any,
    expected: any,
    propertyName: string,
    nullValue?: any
): void {
    if (expected !== null && expected !== undefined) {
        expect(actual)
            .property(propertyName)
            .equal(expected, `Property ${propertyName} is not equal to expected value`);
    } else if (nullValue !== undefined) {
        expect(actual)
            .property(propertyName)
            .equal(nullValue, `Property ${propertyName} is not equal to expected value`);
    }
}
