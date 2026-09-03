/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const imageUtil = {
  createBase64DataUrl,
};

function createBase64DataUrl(logoData: string | undefined | null): string | undefined {
  if (!logoData) return undefined;

  if (logoData.startsWith("data:")) {
    return logoData;
  }

  return `data:image/png;base64,${logoData}`;
}
