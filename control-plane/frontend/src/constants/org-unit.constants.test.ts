/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { ENABLE_ORG_UNIT_FORMS_V3 } from "./org-unit.constants";

describe("org unit constants", () => {
  it("keeps the V3 org unit forms disabled by default", () => expect(ENABLE_ORG_UNIT_FORMS_V3).toBe(false));
});
