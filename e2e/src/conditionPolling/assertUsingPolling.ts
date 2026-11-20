/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';

function sleep(sleepDurationInSeconds): void {
    const now = new Date().getTime();
    while (new Date().getTime() < now + sleepDurationInSeconds * 1000) {
        // sleep for sleepDuration seconds
    }
}

class UnexpectedValueError extends Error {}

/**
 * Validates that running <instance>.<fn> returns <expectedValue> using condition polling.
 * It will retry <retries> times (default=30) and sleep in intervals of <sleepInterval> seconds
 * (default=2s) between retries.
 * Throws an Error in case it doesn't.
 * @param instance - an instance of the class containing the function(fn).
 * @param fn - function to run
 * @param args - ordered list with all the fn args
 * @param expectedValue - expected result of applying fn with args
 * @param retries - amount of retries
 * @param sleepInterval - interval to sleep between retries (seconds)
 */
export function assertExpectedValue(
    instance: object,
    fn,
    expectedValue: string,
    args: any[],
    retries = 10,
    sleepInterval = 2
): void {
    let functionOutput: any;
    console.log(
        `Running function: ${fn.name}.\n  arguments: ${args}\n  retries: ${retries}\n  sleep interval(seconds): ${sleepInterval}`
    );
    for (let i = 0; i < retries; i++) {
        try {
            functionOutput = fn.call(instance, ...args);
            expect(functionOutput).to.be.eq(expectedValue);
            return;
        } catch (err) {
            console.log(
                `Attempt number ${
                    i + 1
                } - Expected ${functionOutput} to be equal to ${expectedValue}, got this error: ${err}`
            );
            sleep(sleepInterval);
        }
    }
    throw new UnexpectedValueError(
        `Expected ${functionOutput} to be equal to ${expectedValue}/n # end of error # /n`
    );
}
export function assertExpectedValueInclude(
    instance: object,
    fn,
    expectedValue: string,
    args: any[],
    retries = 10,
    sleepInterval = 2
): void {
    let functionOutput: any;
    console.log(
        `Running function: ${fn.name}.\n  arguments: ${args}\n  retries: ${retries}\n  sleep interval(seconds): ${sleepInterval}`
    );
    for (let i = 0; i < retries; i++) {
        try {
            functionOutput = fn.call(instance, ...args);
            expect(functionOutput).to.include(expectedValue);
            return;
        } catch (err) {
            console.log(
                `Attempt number ${
                    i + 1
                } - Expected ${functionOutput} to include ${expectedValue}, got this error: ${err}`
            );
            sleep(sleepInterval);
        }
    }
    throw new UnexpectedValueError(
        `Expected ${functionOutput} to include ${expectedValue}/n # end of error # /n`
    );
}
