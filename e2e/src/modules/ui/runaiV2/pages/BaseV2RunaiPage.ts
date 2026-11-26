/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseRunaiPage } from '../../runaiV1/pages/BaseRunaiPage';
import { FrameLocator } from 'playwright-core';

export class BaseV2RunaiPage extends BaseRunaiPage {
    iframeLocator: string;

    constructor(videoFilePath: string, videoFileType: string) {
        super(videoFilePath, videoFileType);
        this.iframeLocator = 'iframe#runai-new-app';
    }

    getIframeLocator(): FrameLocator {
        return this.page.frameLocator(this.iframeLocator);
    }
}
