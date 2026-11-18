/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { RunaiV1 } from './runaiV1';
import { OcpUI } from './ocpUI';
import { NavigateUtils } from './utils/navigationUtils';

const runaiV1 = new RunaiV1();
const navigateUtils = new NavigateUtils();
const ocpUI = new OcpUI();

export { runaiV1, ocpUI, navigateUtils };
