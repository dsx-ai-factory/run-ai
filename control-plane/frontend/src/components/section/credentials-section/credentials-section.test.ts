/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";

import CredentialsSection from "./credentials-section.vue";

import type { IUIWorkloadCredentialsSectionModel } from "./credentials-section.model";
import type { TTestEmit } from "@/models/unit-test.model";
import { EWorkloadFormType } from "@/models/workload.model";

describe("credentials-section.vue", () => {
  function getMountedCmp(token: string): VueWrapper<InstanceType<typeof CredentialsSection>> {
    return mount(CredentialsSection, {
      global: {
        plugins: [Quasar],
      },
      props: {
        entityType: EWorkloadFormType.Inference,
        disableClosing: false,
        credentials: {
          token,
        },
      },
    });
  }

  test("emits 'credentials-changed' event when the input token value changes", async () => {
    const token = "new-value";
    const credentialsSectionCmp = getMountedCmp("");
    await credentialsSectionCmp.find("input").setValue(token);
    const expectedEmits: TTestEmit<IUIWorkloadCredentialsSectionModel> = [[{ token }]];
    expect(credentialsSectionCmp.emitted("credentials-changed")).toEqual(expectedEmits);
  });

  test("emits 'is-section-invalid' event when token value is valid", async () => {
    const validToken = "hf_avalidToken";
    const credentialsSectionCmp = getMountedCmp(validToken);
    const expectedValidatorEvent: TTestEmit<boolean> = [[false]];
    expect(credentialsSectionCmp.emitted("is-section-invalid")).toEqual(expectedValidatorEvent);
  });

  test("emits 'is-section-invalid' as true when token is empty", async () => {
    const invalidToken = "";
    const credentialsSectionCmp = getMountedCmp(invalidToken);
    const expectedEmit: TTestEmit<boolean> = [[true]];
    expect(credentialsSectionCmp.emitted("is-section-invalid")).toEqual(expectedEmit);
  });

  test("emits 'is-section-invalid' as true when token is invalid", async () => {
    const invalidToken = "h_invalid-token";
    const credentialsSectionCmp = getMountedCmp(invalidToken);
    const expectedEmit: TTestEmit<boolean> = [[true]];
    expect(credentialsSectionCmp.emitted("is-section-invalid")).toEqual(expectedEmit);
  });
});
