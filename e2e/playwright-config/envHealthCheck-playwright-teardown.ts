/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { hooksUtils } from '../src/modules/helpers';

function envHealthCheckGlobalTeardown(): void {
    hooksUtils.forceTheTerminationOfOpenBrowsers();
}

export default envHealthCheckGlobalTeardown; 