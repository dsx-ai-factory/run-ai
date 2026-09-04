/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import NimModelAndAccess from "./nim-model-and-access.vue";
import type { INimModelAndAccessModel } from "./nim-model-and-access.model";
import type { INimAccessModelModel } from "../nim-access-model/nim-access-model.model";
import type { IScopeModel } from "@/models/global.model";
import { Scope } from "@/swagger-models/assets-service-client";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

describe("nim-model-and-access.vue", () => {
  const clusterId = "cluster-1";

  const scopeModel: IScopeModel = {
    scope: Scope.Project,
    projectId: 1,
  } as IScopeModel;

  const baseAccessModel: INimAccessModelModel = {
    accessFromStorage: false,
  };

  const baseModelValue: INimModelAndAccessModel = {
    modelName: "",
    profileNameOrHash: "",
    accessModel: baseAccessModel,
  };

  const createWrapper = (modelValue: INimModelAndAccessModel = baseModelValue) =>
    mount(NimModelAndAccess, {
      global: {
        plugins: [Quasar],
        stubs: {
          NimModelSelector: true,
          NimProfileSelector: true,
          NimAccessModel: true,
        },
      },
      props: {
        modelValue: { ...modelValue, accessModel: { ...modelValue.accessModel } },
        clusterId,
        scopeModel,
      },
    });

  describe("onModelNameChanged", () => {
    it("clears profile when model name is cleared", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        modelName: "meta/llama",
        profileNameOrHash: "my-profile",
        accessModel: baseAccessModel,
      });

      wrapper.vm.onModelNameChanged("");

      const emitted = wrapper.emitted("update:model-value") as Array<[INimModelAndAccessModel]> | undefined;
      expect(emitted?.[0]?.[0]).toEqual({
        modelName: "",
        profileNameOrHash: "",
        accessModel: baseAccessModel,
      });
    });

    it("clears profile when selecting a model for the first time", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        modelName: "",
        profileNameOrHash: "",
        accessModel: baseAccessModel,
      });

      wrapper.vm.onModelNameChanged("meta/llama-3.1-8b-base");

      const emitted = wrapper.emitted("update:model-value") as Array<[INimModelAndAccessModel]> | undefined;
      expect(emitted?.[0]?.[0]).toEqual({
        modelName: "meta/llama-3.1-8b-base",
        profileNameOrHash: "",
        accessModel: baseAccessModel,
      });
    });

    it("preserves profile when model name is unchanged", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        modelName: "meta/llama",
        profileNameOrHash: "profile-hash-1",
        accessModel: baseAccessModel,
      });

      wrapper.vm.onModelNameChanged("meta/llama");

      const emitted = wrapper.emitted("update:model-value") as Array<[INimModelAndAccessModel]> | undefined;
      expect(emitted?.[0]?.[0]).toEqual({
        modelName: "meta/llama",
        profileNameOrHash: "profile-hash-1",
        accessModel: baseAccessModel,
      });
    });

    it("clears profile when switching to a different model", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        modelName: "meta/llama",
        profileNameOrHash: "profile-hash-1",
        accessModel: baseAccessModel,
      });

      wrapper.vm.onModelNameChanged("nvidia/parakeet-ctc-1.1b-asr");

      const emitted = wrapper.emitted("update:model-value") as Array<[INimModelAndAccessModel]> | undefined;
      expect(emitted?.[0]?.[0]).toEqual({
        modelName: "nvidia/parakeet-ctc-1.1b-asr",
        profileNameOrHash: "",
        accessModel: baseAccessModel,
      });
    });
  });

  describe("onProfileChanged", () => {
    it("emits update with new profile and keeps model name and access model", () => {
      const accessModel: INimAccessModelModel = {
        ...baseAccessModel,
        apiKey: "secret-key",
      };
      const wrapper = createWrapper({
        ...baseModelValue,
        modelName: "meta/llama",
        profileNameOrHash: "old-profile",
        accessModel,
      });

      wrapper.vm.onProfileChanged("new-profile");

      const emitted = wrapper.emitted("update:model-value") as Array<[INimModelAndAccessModel]> | undefined;
      expect(emitted?.[0]?.[0]).toEqual({
        modelName: "meta/llama",
        profileNameOrHash: "new-profile",
        accessModel,
      });
    });
  });

  describe("onAccessModelChanged", () => {
    it("emits update with new access model and keeps model name and profile", () => {
      const wrapper = createWrapper({
        ...baseModelValue,
        modelName: "meta/llama",
        profileNameOrHash: "profile-1",
        accessModel: baseAccessModel,
      });

      const nextAccess: INimAccessModelModel = {
        ...baseAccessModel,
        selectedCredentials: "cred",
        credentialsSource: ERunaiCredentialSource.SharedSecret,
      };
      wrapper.vm.onAccessModelChanged(nextAccess);

      const emitted = wrapper.emitted("update:model-value") as Array<[INimModelAndAccessModel]> | undefined;
      expect(emitted?.[0]?.[0]).toEqual({
        modelName: "meta/llama",
        profileNameOrHash: "profile-1",
        accessModel: nextAccess,
      });
    });
  });
});
