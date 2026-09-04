/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import NfsDataMount from "./nfs-data-mount-section.vue";
import { beforeEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { NFSAssetSpec } from "@/swagger-models/assets-service-client";

describe("nfs-data-mount.vue", () => {
  let nfsDataMountCmp: VueWrapper<InstanceType<typeof NfsDataMount>>;
  const props = {
    modelValue: {
      path: "/test-path",
      readOnly: null,
      server: "server-test",
      mountPath: "/mount-path-test",
    } as NFSAssetSpec,
  };
  beforeEach(() => {
    nfsDataMountCmp = mount(NfsDataMount, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  });

  test("the component will emit the updated model-value on a value change", async () => {
    nfsDataMountCmp.find(".server-input").setValue("new-server");
    await nfsDataMountCmp.vm.$nextTick();
    const expectedEmit: NFSAssetSpec[][] = [[{ ...props.modelValue, server: "new-server" }]];
    expect(nfsDataMountCmp.emitted("update:modelValue")).toEqual(expectedEmit);
  });
});
