/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import SamlSettings from "./saml-settings.vue";
import type { Idp, SamlData } from "@/swagger-models/identity-manager-client";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import { idpCommon } from "../idp-common/idp-common";

describe("saml-settings.vue", () => {
  const mockSamlData: SamlData = {
    entityId: "runai-app.run.ai/bgu",
    singleSignOnServiceUrl: "http://test.com/sso",
    signingCertificate: "test-cert",
    validateSignature: true,
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
          NotificationAlertBar: true,
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

    expect(wrapper.emitted("update-idp")).toBeTruthy();
    expect(wrapper.emitted("update-idp")?.[0]?.[0]).toMatchObject({
      samlData: { entityId: mockSamlData.entityId },
    });
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

    expect(wrapper.emitted("update-idp")).toBeTruthy();
    expect(wrapper.emitted("update-idp")?.[0]?.[0]).toMatchObject({
      samlData: { entityId: mockSamlData.entityId },
    });
  });

  it("blocks the update when the newly uploaded metadata file has an invalid certificate", async () => {
    const { wrapper } = factory(mockIdp, false);

    await wrapper.setData({
      isIdpFieldsChanged: true,
      isValidIdpFields: false,
      idpFields: {
        metadataXmlFile: "base64-file-contents-with-expired-cert",
        metadataXmlType: "file",
      },
    });

    await wrapper.vm.onUpdate();

    expect(wrapper.emitted("update-idp")).toBeFalsy();
  });

  it("blocks unrelated changes (e.g. mappers) while the stored certificate is invalid", async () => {
    const { wrapper } = factory(mockIdp, false);

    await wrapper.setData({
      isIdpFieldsChanged: false,
      isValidIdpFields: false,
      idpMappersItems: { email: { value: "email", override: false } },
    });

    await wrapper.vm.onUpdate();

    expect(wrapper.emitted("update-idp")).toBeFalsy();
  });

  it("allows unrelated changes (e.g. mappers) while the certificate is valid", async () => {
    const { wrapper } = factory(mockIdp, false);

    await wrapper.setData({
      isIdpFieldsChanged: false,
      isValidIdpFields: true,
      idpMappersItems: { email: { value: "email", override: false } },
    });

    await wrapper.vm.onUpdate();

    expect(wrapper.emitted("update-idp")).toBeTruthy();
  });

  describe("certificate expiration warning", () => {
    const now = new Date("2026-01-01T00:00:00.000Z");
    const daysFromNow = (days: number): Date => new Date(now.getTime() + days * 24 * 60 * 60 * 1000);

    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(now);
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("shows no warning when the certificate is valid and not close to expiring", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(120));

      const { wrapper } = factory(mockIdp, false);
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        null,
      );
    });

    it("shows a warning banner when the certificate expires within 60 days", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(60));

      const { wrapper } = factory(mockIdp, false);
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        "warning",
      );
    });

    it("does not warn when the certificate expires in 61 days", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(61));

      const { wrapper } = factory(mockIdp, false);
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        null,
      );
    });

    it("shows an error banner when the certificate is already expired", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(-1));

      const { wrapper } = factory(mockIdp, false);
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        "error",
      );
    });

    it("refreshes the expiration date when the signing certificate is replaced", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(-1));

      const { wrapper } = factory(mockIdp, false);
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        "error",
      );

      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(120));

      await wrapper.setProps({
        idp: {
          ...mockIdp,
          samlData: { ...mockSamlData, signingCertificate: "new-cert" },
        },
      });
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        null,
      );
    });

    it("clears the expiration date when the signing certificate is removed", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(-1));

      const { wrapper } = factory(mockIdp, false);
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        "error",
      );

      await wrapper.setProps({
        idp: { ...mockIdp, samlData: { ...mockSamlData, signingCertificate: "" } },
      });
      await wrapper.vm.$nextTick();
      await wrapper.vm.$nextTick();

      expect((wrapper.vm as unknown as { certificateWarningVariant: string | null }).certificateWarningVariant).toBe(
        null,
      );
    });
  });
});
