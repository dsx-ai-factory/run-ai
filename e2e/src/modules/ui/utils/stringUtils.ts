/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

const camelCaseToHyphenated = (str: string): string => {
    return str.replace(/[A-Z]/g, function (match: string, offset: number) {
        return offset > 0 ? '-' + match.toLowerCase() : match.toLowerCase();
    });
};

const extractUrl = (data: string): string => {
    const urlRegex = /(https?:\/\/[^\s]+)/g;
    return data.match(urlRegex).at(0);
};
export const stringUtil = {
    camelCaseToHyphenated,
    extractUrl
};
