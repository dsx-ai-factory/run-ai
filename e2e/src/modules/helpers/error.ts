/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

class BaseError extends Error {
    constructor(message: string) {
        super(message);
        this.name = this.constructor.name;
        Error.captureStackTrace(this, this.constructor);
    }
}

class BackendApiError extends BaseError {
    constructor(error: Error) {
        super(error.message);
    }
}
class UtilsError extends BaseError {
    constructor(error: Error) {
        super(error.message);
    }
}

class CliError extends BaseError {
    constructor(error: Error) {
        super(error.message);
    }
}
export { BackendApiError, UtilsError, CliError };
