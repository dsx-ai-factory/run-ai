/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { settings } from '../../helpers';
import { Login } from './pages/Login';
import { Cluster } from './pages/Cluster';

export class RunaiV1 {
    login = new Login(settings.browserVideosPath, settings.browserVideosFileType);

    cluster = new Cluster(settings.browserVideosPath, settings.browserVideosFileType);
}
