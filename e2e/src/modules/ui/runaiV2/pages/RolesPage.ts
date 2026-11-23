/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

export class RolesPage extends BaseV2Page {
    constructor(page: Page) {
        super(page);

        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Access}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Roles}"]`;
    }
}
