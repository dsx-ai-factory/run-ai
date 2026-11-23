/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { setTimer } from './runai-workloads-kit.config';
import { runKitGlobalSetup } from './kit-shared-global-setup';

export async function globalSetup(): Promise<void> {
    await runKitGlobalSetup({
        kitName: 'runai-workloads-kit',
        setTimerFn: setTimer
    });
}

export default globalSetup;
