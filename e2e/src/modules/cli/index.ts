/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubeCli } from './tools/KubeCli';
import { config } from '../helpers';
import { BashCli } from './tools/BashCli';
import { RunaiCliV2 } from './tools/RunaiCliV2';
import { Md5FileHasher } from './tools/Md5FileHasher';
import * as distributedInferenceHelpers from './helpers/DistributedInferenceTestHelpers';

let kubeCli = new KubeCli(config.kubeconfigPath, config.kubeconfigClusterName);
const runaiCliV2 = new RunaiCliV2();
const bashCli = new BashCli();

function setKubeCli(clusterName: string): void {
    kubeCli = new KubeCli(config.kubeconfigPath, clusterName);
}
export { kubeCli, bashCli, runaiCliV2, Md5FileHasher, setKubeCli, distributedInferenceHelpers };
