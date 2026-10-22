/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { markRaw } from "vue";
import SamlFields from "./saml-fields.vue";
import type { SamlConfigSettings } from "@/models/setting.model";
import { SamlCreationDataMetadataXmlTypeEnum } from "@/swagger-models/identity-manager-client";
import { idpCommon, SAML_SIGNING_DISABLED_TITLE } from "../../idp-common/idp-common";

describe("saml-fields.vue", () => {
  const validSamlSettings: SamlConfigSettings = {
    idpEntityId: "runai-app.run.ai/bgu",
    singleSignOnServiceUrl: "http://test.com/sso",
    signingCertificate: "test-cert",
  };

  const factory = (
    props: Partial<InstanceType<typeof SamlFields>["$props"]> = {},
    dataOverrides: Record<string, unknown> = {},
  ) => {
    setActivePinia(createTestingPinia({ createSpy: vi.fn }));

    const wrapper = mount(SamlFields, {
      global: {
        plugins: [Quasar],
      },
      props: {
        isNewIdp: false,
        ...props,
      },
      data() {
        return dataOverrides;
      },
    });

    return { wrapper };
  };

  const fakeMetadataXml =
    '<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"><md:IDPSSODescriptor>' +
    '<md:KeyDescriptor use="signing"><ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:X509Data>' +
    "<ds:X509Certificate>fake-cert-content</ds:X509Certificate>" +
    "</ds:X509Data></ds:KeyInfo></md:KeyDescriptor></md:IDPSSODescriptor></md:EntityDescriptor>";

  // Simulates picking a file without going through the real FileReader flow:
  // sets metadataFile (as the real file-picker handler does) and directly
  // invokes the same certificate-parsing path setFileContent triggers.
  const uploadFile = async (wrapper: ReturnType<typeof factory>["wrapper"], fileName = "idp.xml") => {
    (wrapper.vm as unknown as { metadataFile: File | null }).metadataFile = markRaw(
      new File([fakeMetadataXml], fileName),
    );
    await wrapper.vm.$nextTick();
    await (wrapper.vm as unknown as { setFileContent: (n: string, c: string) => Promise<void> }).setFileContent(
      fileName,
      fakeMetadataXml,
    );
    await flushPromises();
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("isValidFile", () => {
    it("requires a file even for an existing IdP once the field is empty", () => {
      const { wrapper } = factory({ samlSettings: validSamlSettings, isNewIdp: false });

      expect((wrapper.vm as unknown as { isValidFile: (f: string) => boolean | string }).isValidFile("")).toBe(
        "Upload a file",
      );
    });

    it("requires a file for a new IdP", () => {
      const { wrapper } = factory({ isNewIdp: true });

      expect((wrapper.vm as unknown as { isValidFile: (f: string) => boolean | string }).isValidFile("")).toBe(
        "Upload a file",
      );
    });

    it("shows the required-file message when Save is pressed on a brand-new, untouched form", async () => {
      const { wrapper } = factory({ isNewIdp: true });

      await (wrapper.vm as unknown as { validate: () => Promise<boolean> }).validate();
      await wrapper.vm.$nextTick();

      expect(wrapper.text()).toContain("Upload a file");
    });

    it("passes once a file is present", () => {
      const { wrapper } = factory({ isNewIdp: true });

      expect(
        (wrapper.vm as unknown as { isValidFile: (f: string) => boolean | string }).isValidFile("metadata.xml"),
      ).toBe(true);
    });
  });

  describe("isValidUrl", () => {
    it("requires a URL when switching an existing file-sourced IdP to URL", () => {
      const { wrapper } = factory({
        samlSettings: validSamlSettings,
        isNewIdp: false,
        idpResources: { resourceName: "idp.xml", resource: fakeMetadataXml },
      });

      expect((wrapper.vm as unknown as { isValidUrl: (u: string) => boolean | string }).isValidUrl("")).toBe(
        "Enter a valid URL",
      );
    });

    it("allows leaving the URL blank for an existing URL-sourced IdP", () => {
      const { wrapper } = factory({
        samlSettings: validSamlSettings,
        isNewIdp: false,
        idpResources: {
          resourceName: SamlCreationDataMetadataXmlTypeEnum.Url,
          resource: "https://idp.example.com/metadata",
        },
      });

      expect((wrapper.vm as unknown as { isValidUrl: (u: string) => boolean | string }).isValidUrl("")).toBe(true);
    });

    it("requires a URL for a new IdP", () => {
      const { wrapper } = factory({ isNewIdp: true });

      expect((wrapper.vm as unknown as { isValidUrl: (u: string) => boolean | string }).isValidUrl("")).toBe(
        "Enter a valid URL",
      );
    });
  });

  describe("certificate expiration field", () => {
    const now = new Date("2026-01-01T00:00:00.000Z");
    const daysFromNow = (days: number): Date => new Date(now.getTime() + days * 24 * 60 * 60 * 1000);

    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(now);
    });

    it("shows the expired-file message for an uploaded file, and hides the field once it's cleared", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(-1));

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect(wrapper.text()).toContain("Certificate expired. Replace the metadata file above.");

      (wrapper.vm as unknown as { clearFile: () => void }).clearFile();
      await wrapper.vm.$nextTick();

      expect(wrapper.text()).not.toContain("Certificate expiration date");
      expect(wrapper.text()).not.toContain("Certificate expired");
    });

    it("shows the expired-URL message when sourced from a URL", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(-1));

      const { wrapper } = factory(
        { samlSettings: validSamlSettings, isNewIdp: false },
        { samlSource: SamlCreationDataMetadataXmlTypeEnum.Url },
      );
      await flushPromises();

      expect(wrapper.text()).toContain("Certificate expired. Replace the metadata URL above.");
    });

    it("does not show an error for a certificate that is still valid", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(120));

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect(wrapper.text()).not.toContain("Certificate expired");
    });
  });

  describe("setSamlResources", () => {
    it("initializes metadataXmlUrl for an existing URL-sourced IdP so mapper-only updates aren't blocked", async () => {
      const { wrapper } = factory({
        samlSettings: validSamlSettings,
        isNewIdp: false,
        idpResources: {
          resourceName: SamlCreationDataMetadataXmlTypeEnum.Url,
          resource: "https://idp.example.com/metadata",
        },
      });
      await flushPromises();

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(true);
    });

    it("does not report changes for an untouched, existing file-sourced IdP right after mount", async () => {
      const { wrapper } = factory({
        samlSettings: validSamlSettings,
        isNewIdp: false,
        idpResources: { resourceName: "idp.xml", resource: fakeMetadataXml },
      });
      await flushPromises();

      expect((wrapper.vm as unknown as { hasChanges: () => boolean }).hasChanges()).toBe(false);
    });
  });

  describe("stale certificate validation guard", () => {
    it("ignores a validation result that resolves after a newer file selection has started", async () => {
      let resolveStale: (d: Date | null) => void = () => {};
      const stalePromise = new Promise<Date | null>((resolve) => {
        resolveStale = resolve;
      });
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockReturnValueOnce(stalePromise);

      const { wrapper } = factory({ isNewIdp: true });
      const vm = wrapper.vm as unknown as {
        setFileContent: (n: string, c: string, t?: number) => Promise<void>;
        fileSelectionToken: number;
        expirationDate: Date | null;
      };

      // Start validating a "first" file selection (token 0), but its lookup hasn't resolved yet.
      const firstCall = vm.setFileContent("first.xml", fakeMetadataXml, 0);

      // A newer selection supersedes it before the first lookup resolves.
      vm.fileSelectionToken = 1;

      // The stale lookup finally resolves with an (irrelevant) expired date.
      resolveStale(new Date("2020-01-01T00:00:00.000Z"));
      await firstCall;
      await wrapper.vm.$nextTick();

      expect(vm.expirationDate).toBeNull();
      expect(wrapper.text()).not.toContain("Certificate expired");
    });
  });

  describe("disabled SAML signing warning", () => {
    const xmlWith = (attribute: string) =>
      `<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata">` +
      `<md:IDPSSODescriptor ${attribute}="false"/>` +
      `</md:EntityDescriptor>`;

    it.each(["WantAssertionsSigned", "AuthnRequestsSigned", "WantAuthnRequestsSigned"])(
      "shows the warning when %s is false",
      async (attribute) => {
        vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(null);

        const { wrapper } = factory({ isNewIdp: true });
        await (wrapper.vm as unknown as { setFileContent: (n: string, c: string) => Promise<void> }).setFileContent(
          "idp.xml",
          xmlWith(attribute),
        );
        await flushPromises();

        expect(wrapper.text()).toContain(SAML_SIGNING_DISABLED_TITLE);
      },
    );

    it("hides the warning when the file is cleared", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(null);

      const { wrapper } = factory({ isNewIdp: true });
      await (wrapper.vm as unknown as { setFileContent: (n: string, c: string) => Promise<void> }).setFileContent(
        "idp.xml",
        xmlWith("WantAssertionsSigned"),
      );
      await flushPromises();

      (wrapper.vm as unknown as { clearFile: () => void }).clearFile();
      await wrapper.vm.$nextTick();

      expect(wrapper.text()).not.toContain(SAML_SIGNING_DISABLED_TITLE);
    });
  });

  describe("isValidInput", () => {
    const now = new Date("2026-01-01T00:00:00.000Z");
    const daysFromNow = (days: number): Date => new Date(now.getTime() + days * 24 * 60 * 60 * 1000);

    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(now);
    });

    it("is invalid when the uploaded file's certificate is already expired", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(-1));

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(false);
    });

    it("is valid when the uploaded file's certificate is still valid", async () => {
      vi.spyOn(idpCommon, "getCertificateExpirationDate").mockResolvedValue(daysFromNow(120));

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(true);
    });
  });
});
