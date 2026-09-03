/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { allAccessRulesColumnsMap } from "./access-rule.table-model";
import { ALL_AUTHENTICATED_USERS_LABEL } from "@/models/access-rule.model";
import { SubjectTypeExtended, type AccessRule } from "@/swagger-models/authorization-client";

const makeRule = (overrides: Partial<AccessRule> = {}): AccessRule =>
  ({
    subjectId: "alice@run.ai",
    subjectType: SubjectTypeExtended.User,
    roleId: 1,
    scopeId: "1001",
    scopeType: "tenant",
    ...overrides,
  }) as AccessRule;

describe("access-rule table model - subject column", () => {
  const subjectField = allAccessRulesColumnsMap.subject.field as (row: AccessRule) => string;

  it("shows the subjectId for a regular subject type", () => {
    expect(subjectField(makeRule({ subjectType: SubjectTypeExtended.User, subjectId: "alice@run.ai" }))).toBe(
      "alice@run.ai",
    );
  });

  it("shows a dash for the all-authenticated-users type (subjectId is a sentinel)", () => {
    expect(
      subjectField(
        makeRule({ subjectType: SubjectTypeExtended.AllAuthenticatedUsers, subjectId: "all-authenticated-users" }),
      ),
    ).toBe("-");
  });
});

describe("access-rule table model - type column", () => {
  const typeFormat = allAccessRulesColumnsMap.type.format as (subjectType: SubjectTypeExtended) => string;

  it("maps the all-authenticated-users type to its friendly label", () => {
    expect(typeFormat(SubjectTypeExtended.AllAuthenticatedUsers)).toBe(ALL_AUTHENTICATED_USERS_LABEL);
  });

  it("maps a known type to its label", () => {
    expect(typeFormat(SubjectTypeExtended.User)).toBe("User");
  });

  it("falls back to a dash for an unknown type", () => {
    expect(typeFormat("something-else" as SubjectTypeExtended)).toBe("-");
  });
});
