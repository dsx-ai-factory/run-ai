/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clearTimer } from './runai-workloads-kit.config';
import { runKitGlobalTeardown } from './kit-shared-global-teardown';

export default async function (): Promise<void> {
    await runKitGlobalTeardown({
        kitName: 'runai-workloads-kit',
        clearTimerFn: clearTimer
    });
}
