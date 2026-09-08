/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import SamlSettings from "./saml-settings.vue";
import type { Idp, SamlData } from "@/swagger-models/identity-manager-client";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";

describe("saml-settings.vue", () => {
  const mockSamlData: SamlData = {
    entityId: "runai-app.run.ai/custom",
    singleSignOnServiceUrl: "http://test.com/sso",
    signingCertificate: "test-cert",
  };

  const mockIdp: Idp = {
    alias: IdpTypeEnum.Saml,
    type: IdpTypeEnum.Saml,
    redirectUri: "http://test.com",
    samlData: mockSamlData,
  };

  const factory = (idp?: Idp, isNewIdp = false) => {
    setActivePinia(createTestingPinia({ createSpy: vi.fn }));

    const wrapper = mount(SamlSettings, {
      global: {
        plugins: [Quasar],
        stubs: {
          SamlFields: true,
          IdpMappers: true,
          SettingEditableField: true,
          SettingWrapperBox: {
            template: "<div><slot /></div>",
          },
          SettingReadableField: true,
          RunaiBtn: true,
        },
      },
      props: {
        title: "SAML",
        idp,
        idpMappers: {},
        isNewIdp,
      },
    });

    return { wrapper };
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("preserves the existing SP entity ID when updating an existing SAML IdP", async () => {
    const { wrapper } = factory(mockIdp, false);

    // Simulate an admin re-uploading metadata: the entity ID field itself
    // is not part of what saml-fields sends back on update.
    await wrapper.setData({
      isIdpFieldsChanged: true,
      isValidIdpFields: true,
      idpFields: {
        metadataXmlFile: "base64-file-contents",
        metadataXmlType: "file",
      },
    });

    await wrapper.vm.onUpdate();

    const emitted = wrapper.emitted("update-idp");
    expect(emitted).toBeTruthy();
    expect(emitted?.[0]?.[0]?.samlData?.entityId).toBe(mockSamlData.entityId);
  });

  it("preserves the entity ID when updating via a metadata URL instead of a file", async () => {
    const { wrapper } = factory(mockIdp, false);

    await wrapper.setData({
      isIdpFieldsChanged: false,
      isValidIdpFields: true,
      idpFields: {
        metadataXmlUrl: "https://idp.example.com/metadata.xml",
      },
    });

    await wrapper.vm.onUpdate();

    const emitted = wrapper.emitted("update-idp");
    expect(emitted).toBeTruthy();
    expect(emitted?.[0]?.[0]?.samlData?.entityId).toBe(mockSamlData.entityId);
  });
});
