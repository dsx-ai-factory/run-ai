/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import PvcDataMount from "./pvc-data-mount.vue";
import { describe, expect, test } from "vitest";
import { Quasar } from "quasar";
import type { StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";

describe("pvc-data-mount.vue", () => {
  test("the component will emit the correct data", async () => {
    const props = {
      pvcDataModel: {
        path: "/testing-path",
        claimName: "test claim name",
        existingPvc: true,
        claimInfo: null,
        readOnly: null,
      },
      isProjectScoped: true,
      pvcList: ["test1", "test2", "test3"] as string[],
      storageClasses: [
        { storageClassName: "sClass1", clusterId: "cluster-1", spec: { is_default: true } },
        { storageClassName: "sClass2", clusterId: "cluster-1", spec: { is_default: true } },
      ] as StorageClassV2[],
      loadingStorageClass: false as boolean,
      loading: false as boolean,
    };
    const wrapper = mount(PvcDataMount, {
      global: {
        plugins: [Quasar],
      },
      props,
    });

    wrapper.vm.$emit("update:pvc-data-model", props.pvcDataModel);
    await wrapper.vm.$nextTick();
    const emitsList = wrapper.emitted();
    expect(emitsList["update:pvc-data-model"][0]).toEqual([props.pvcDataModel]);
  });
});
