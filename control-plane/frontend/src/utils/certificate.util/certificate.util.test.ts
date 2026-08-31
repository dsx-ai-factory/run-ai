/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// @ts-nocheck
import { describe, it, expect } from "vitest";
import { certificateUtil } from "@/utils/certificate.util/certificate.util";
import { FAKE_BASE64_CERT } from "@/tests/fixtures/certificate.fixtures";

describe("certificate.util", () => {
  it("should get expiration date", async () => {
    const base64Cert = FAKE_BASE64_CERT;

    const result: Promise<Date> = await certificateUtil.getExpirationDate(base64Cert);
    expect(result).toStrictEqual(new Date("2031-10-14T11:01:39.000Z")); // or adjust based on your default behavior
  });
  it("should get no expiration date", async () => {
    const base64Cert = "nothing";

    const result: Promise<Date> = await certificateUtil.getExpirationDate(base64Cert);
    expect(result).toStrictEqual(null); // or adjust based on your default behavior
  });
});
