/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export class SoftAssert {
    private errors: string[];

    constructor() {
        this.errors = [];
    }

    assert(condition: boolean, message: string): void {
        if (!condition) {
            this.errors.push(message);
        }
    }

    assertNonStrictEqual<T>(actual: T, expected: T, message: string): void {
        if ((actual === null && expected === undefined) || (expected === null && actual === undefined)) {
            return;
        }
        if (actual !== expected) {
            this.errors.push(`${message}: Expected: ${expected}, Actual: ${actual}`);
        }
    }

    assertDeepNonStrictEqual<T>(actual: T, expected: T, message: string): void {
        for (const key in expected) {
            // eslint-disable-next-line eqeqeq
            if (actual[key] != expected[key]) {
                this.errors.push(
                    `${message}: ${key} is not equal. Expected: ${JSON.stringify(expected[key])}, Actual: ${JSON.stringify(actual[key])}`
                );
            }
        }
    }

    getResult(): string[] {
        return this.errors;
    }

    throwErrorIfFailed(): void {
        const assertMsg = 'SoftAssert failed:\n';
        if (this.errors.length > 0) {
            throw new Error(`${assertMsg}${this.errors.join('\n')}`);
        }
    }
}
