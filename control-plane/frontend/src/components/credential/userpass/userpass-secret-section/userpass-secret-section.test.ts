/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import UserpassSecretSection from "./userpass-secret-section.vue";
import { beforeEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { PasswordCreationSpec } from "@/swagger-models/assets-service-client";

describe("userpass-secret-section.vue", () => {
  let UserpassSecretCmp: VueWrapper<InstanceType<typeof UserpassSecretSection>>;
  const props = {
    secretDetails: {
      existingSecretName: null,
      password: "some-pass",
      user: "some-user",
    } as PasswordCreationSpec,
    existingSecretNames: ["userpass1", "userpass2"] as string[],
    loading: false as boolean,
    projectId: null,
  };
  beforeEach(() => {
    UserpassSecretCmp = mount(UserpassSecretSection, {
      global: {
        plugins: [Quasar],
      },
      props,
      data() {
        return {
          isExistingSecret: false as boolean,
          isPassHidden: true as boolean,
        };
      },
    });
  });

  test("the component will emit the updated secretDetails on a value change", async () => {
    UserpassSecretCmp.find("[aid=userpass-username-input]").setValue("new-user");
    await UserpassSecretCmp.vm.$nextTick();
    const expectedEmit: PasswordCreationSpec[][] = [[{ ...props.secretDetails, user: "new-user" }]];
    expect(UserpassSecretCmp.emitted("update-secret-details")).toEqual(expectedEmit);
  });
});
