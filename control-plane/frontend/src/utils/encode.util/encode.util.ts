/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const encodeUtil = {
  toSha256,
};

async function toSha256(message: string): Promise<string> {
  // Encode the message as a UTF-8 byte array
  const encoder = new TextEncoder();
  const data = encoder.encode(message);
  // Hash the message with SHA-256
  const hashBuffer = await crypto.subtle.digest("SHA-256", data);
  // Convert the hash to a hexadecimal string
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
  return hashHex;
}
