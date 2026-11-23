/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { config } from '../index';

export interface AccessRulesFilterCase {
    name: string;
    filterBy: string[];
    field: string;
    expected: string;
    matcher: (value: string, expected: string) => void;
}

export const accessRulesFilterCases: AccessRulesFilterCase[] = [
    {
        name: 'test type filter',
        filterBy: ['subjectType!=app,subjectType!=service-account,subjectType!=group'],
        field: 'subjectType',
        expected: 'user',
        matcher: (value, expected): void => {
            expect(value, 'subjectType should be user when other subject types are excluded').to.equal(
                expected
            );
        }
    },
    {
        name: 'test subject filter',
        filterBy: ['subjectId=@l1-researcher'],
        field: 'subjectId',
        expected: 'l1-researcher',
        matcher: (value, expected): void => {
            expect(value, 'subjectId should contain the filtered subject').to.contain(expected);
        }
    },
    {
        name: 'test role filter',
        filterBy: ['roleName=@compute resource'],
        field: 'roleName',
        expected: 'Compute resource administrator',
        matcher: (value, expected): void => {
            expect(value, 'roleName should match the filtered role').to.equal(expected);
        }
    },
    {
        name: 'test scope filter',
        filterBy: ['scopeName=@runai'],
        field: 'scopeName',
        expected: 'runai',
        matcher: (value, expected): void => {
            expect(value, 'scopeName should match the filtered scope').to.equal(expected);
        }
    },
    {
        name: 'test status filter',
        filterBy: ['phase!=Creating,phase!=Deleting,phase!=NotReady'],
        field: 'status.phase',
        expected: 'Ready',
        matcher: (value, expected): void => {
            expect(value, 'status phase should be Ready when transitional phases are excluded').to.equal(
                expected
            );
        }
    },
    {
        name: 'test authorized by filter',
        filterBy: [`createdBy=@${config.controlPlaneAdminUsername}`],
        field: 'createdBy',
        expected: config.controlPlaneAdminUsername,
        matcher: (value, expected): void => {
            expect(value, 'createdBy should match the filtered creator').to.equal(expected);
        }
    }
];

export function getAccessRuleFieldValue(accessRule: object, field: string): string {
    return field.split('.').reduce((o, k) => o[k], accessRule as any);
}
