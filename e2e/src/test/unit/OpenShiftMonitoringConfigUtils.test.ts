/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import {
    getRunaiRemoteWrites,
    monitoringConfigWithoutRunaiRemoteWrites
} from '../../modules/helpers/utils/OpenShiftMonitoringConfigUtils';

test.describe('monitoringConfigWithoutRunaiRemoteWrites', () => {
    test('removes only Run:ai remote writes and preserves unrelated configuration', () => {
        const config = monitoringConfigWithoutRunaiRemoteWrites(`
nodeExporter:
  collectors:
    cpufreq: {}
prometheusK8s:
  retention: 15d
  remoteWrite:
    - name: customer
      url: https://customer.example.com
      unknownField: preserved
    - name: runai
      url: https://runai.example.com
`);

        expect(config).toEqual({
            nodeExporter: { collectors: { cpufreq: {} } },
            prometheusK8s: {
                retention: '15d',
                remoteWrite: [
                    {
                        name: 'customer',
                        url: 'https://customer.example.com',
                        unknownField: 'preserved'
                    }
                ]
            }
        });
    });

    test('returns the Run:ai-owned remote write', () => {
        const remoteWrites = getRunaiRemoteWrites(`
prometheusK8s:
  remoteWrite:
    - name: customer
      url: https://customer.example.com
    - name: runai
      url: https://runai.example.com
      basicAuth:
        username:
          key: remoteWriteUsername-runai
        password:
          key: remoteWritePassword-runai
`);

        expect(remoteWrites).toEqual([
            {
                name: 'runai',
                url: 'https://runai.example.com',
                basicAuth: {
                    username: { key: 'remoteWriteUsername-runai' },
                    password: { key: 'remoteWritePassword-runai' }
                }
            }
        ]);
    });

    test('removes an empty prometheusK8s parent created only for Run:ai remote writes', () => {
        const config = monitoringConfigWithoutRunaiRemoteWrites(`
prometheusK8s:
  remoteWrite:
    - name: runai
      url: https://runai.example.com
`);

        expect(config).toEqual({});
    });
});
