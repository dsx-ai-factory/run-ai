/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { bffServiceApi } from '../../api';
import { BrandingSettingsKeyEnum, GetLogo } from '../../api/apis/bffServiceApi';
import { flowRootState } from '../index';

export class BffUtils {
    private get accessToken(): string {
        const token = flowRootState.users.system_administrator.accessToken;
        if (!token) throw new Error('system administrator access token missing');
        return token;
    }

    async getLogo(): Promise<GetLogo> {
        bffServiceApi.settingsApi.accessToken = this.accessToken;
        const { body } = await bffServiceApi.settingsApi.getBrandingSettingByKey(
            BrandingSettingsKeyEnum.Logo
        );
        if (!body) throw new Error('branding settings "logo" API returned empty body');
        return body as GetLogo;
    }

    async updateLogo(logoDataUrl: string): Promise<void> {
        bffServiceApi.settingsApi.accessToken = this.accessToken;
        await bffServiceApi.settingsApi.updateBrandingSettingByKey(BrandingSettingsKeyEnum.Logo, {
            logo: logoDataUrl
        });
    }
}
