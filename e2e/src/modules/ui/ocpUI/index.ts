/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { settings } from '../../helpers';
import { Oauth } from './pages/Oauth';

export class OcpUI {
    oauth = new Oauth(settings.browserVideosPath, settings.browserVideosFileType);
}
