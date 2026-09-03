/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import {
  ALL_AUTHENTICATED_USERS_LABEL,
  accessRuleSubjectTypeFilterOptions,
  subjectTypeOptions,
  subjectTypeOptionsWithSSO,
} from "./access-rule.model";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client";

describe("access-rule model - subject type options", () => {
  it("the non-SSO list offers all-authenticated-users", () => {
    const option = subjectTypeOptions.find((o) => o.value === SubjectTypeExtended.AllAuthenticatedUsers);
    expect(option?.label).toBe(ALL_AUTHENTICATED_USERS_LABEL);
  });

  it("the SSO list offers all-authenticated-users", () => {
    const option = subjectTypeOptionsWithSSO.find((o) => o.value === SubjectTypeExtended.AllAuthenticatedUsers);
    expect(option?.label).toBe(ALL_AUTHENTICATED_USERS_LABEL);
  });

  it("the table filter uses the friendly label for all-authenticated-users", () => {
    const option = accessRuleSubjectTypeFilterOptions.find((o) => o.value === SubjectTypeExtended.AllAuthenticatedUsers);
    expect(option?.label).toBe(ALL_AUTHENTICATED_USERS_LABEL);
  });
});
