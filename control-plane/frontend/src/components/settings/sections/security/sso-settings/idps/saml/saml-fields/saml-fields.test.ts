/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { markRaw } from "vue";
import SamlFields from "./saml-fields.vue";
import type { IdpPostResponse } from "@/swagger-models/identity-manager-client";
import {
  IdpCreationRequestTypeEnum,
  SamlCreationDataMetadataXmlTypeEnum,
} from "@/swagger-models/identity-manager-client";
import { idpService } from "@/services/control-plane/idp.service/idp.service";

vi.mock("@/services/control-plane/idp.service/idp.service", () => ({
  idpService: { dryRunIdp: vi.fn() },
}));

const validExpirationDate = "2030-05-17T00:00:00.000Z";
const expiredExpirationDate = "2020-01-01T00:00:00.000Z";
const acceptedMetadata: IdpPostResponse = {
  alias: "saml",
  samlData: {
    signingCertificate: "fake-cert-content",
    certificateExpirationDate: validExpirationDate,
    validateSignature: false,
  },
};
const rejectedMetadata = {
  message: "invalid saml signing certificate",
  samlData: { signingCertificate: "expired-cert", certificateExpirationDate: expiredExpirationDate },
};

describe("saml-fields.vue", () => {
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

  // Simulates picking a file without going through the real FileReader flow.
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
    vi.mocked(idpService.dryRunIdp).mockResolvedValue(acceptedMetadata);
  });

  describe("isValidFile", () => {
    it("requires a file even for an existing IdP once the field is empty", () => {
      const { wrapper } = factory({ isNewIdp: false });

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
        isNewIdp: false,
        idpResources: { resourceName: "idp.xml", resource: fakeMetadataXml },
      });

      expect((wrapper.vm as unknown as { isValidUrl: (u: string) => boolean | string }).isValidUrl("")).toBe(
        "Enter a valid URL",
      );
    });

    it("allows leaving the URL blank for an existing URL-sourced IdP", () => {
      const { wrapper } = factory({
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

  describe("certificate details", () => {
    const lastCertificateDetails = (wrapper: ReturnType<typeof factory>["wrapper"]) =>
      wrapper.emitted("certificate-details")?.at(-1)?.[0];

    it("reports the expired-file message for an uploaded file, and hides the date once it's cleared", async () => {
      vi.mocked(idpService.dryRunIdp).mockRejectedValue(rejectedMetadata);

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect(idpService.dryRunIdp).toHaveBeenCalledWith({
        type: IdpCreationRequestTypeEnum.Saml,
        samlData: expect.objectContaining({
          metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.File,
          metadataXmlFile: btoa(fakeMetadataXml),
          fileName: "idp.xml",
        }),
        dryRun: true,
      });
      expect(lastCertificateDetails(wrapper)).toMatchObject({
        showExpirationDate: true,
        expirationDate: new Date(expiredExpirationDate),
        expirationError: "Certificate expired. Replace the metadata file above.",
      });

      (wrapper.vm as unknown as { clearFile: () => void }).clearFile();
      await wrapper.vm.$nextTick();

      expect(lastCertificateDetails(wrapper)).toMatchObject({ showExpirationDate: false, expirationError: "" });
    });

    it("reports the expired-URL message for a stored signing certificate", async () => {
      const { wrapper } = factory(
        {
          isNewIdp: false,
          hasSigningCertificate: true,
          certificateExpirationDate: new Date(expiredExpirationDate),
        },
        { samlSource: SamlCreationDataMetadataXmlTypeEnum.Url },
      );
      await flushPromises();

      expect(idpService.dryRunIdp).not.toHaveBeenCalled();
      expect(lastCertificateDetails(wrapper)).toMatchObject({
        showExpirationDate: true,
        expirationDate: new Date(expiredExpirationDate),
        expirationError: "Certificate expired. Replace the metadata URL above.",
      });
    });

    it("reports the expiration date and no error when identity-manager accepts the file", async () => {
      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect(lastCertificateDetails(wrapper)).toMatchObject({
        showExpirationDate: true,
        expirationDate: new Date(validExpirationDate),
        expirationError: "",
      });
    });

    it("parses the expiration date identity-manager returns without fractional seconds", async () => {
      const expirationDate = "2036-10-14T12:25:09Z";
      vi.mocked(idpService.dryRunIdp).mockResolvedValue({
        alias: "",
        samlData: {
          signingCertificate: "fake-cert-content",
          certificateExpirationDate: expirationDate,
          validateSignature: false,
        },
      });

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect(lastCertificateDetails(wrapper)).toMatchObject({
        showExpirationDate: true,
        expirationDate: new Date(expirationDate),
        expirationError: "",
      });
    });

    it("no longer renders the read-only identity provider fields", () => {
      const { wrapper } = factory({ isNewIdp: false });

      expect(wrapper.text()).not.toContain("Identity provider URL");
      expect(wrapper.text()).not.toContain("Certificate expiration date");
    });
  });

  describe("setSamlResources", () => {
    it("initializes metadataXmlUrl for an existing URL-sourced IdP so mapper-only updates aren't blocked", async () => {
      const { wrapper } = factory({
        isNewIdp: false,
        idpResources: {
          resourceName: SamlCreationDataMetadataXmlTypeEnum.Url,
          resource: "https://idp.example.com/metadata",
        },
      });
      await flushPromises();

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(true);
      expect(idpService.dryRunIdp).not.toHaveBeenCalled();
    });

    it("fills the URL field when the stored resource of a URL-sourced IdP is a URL", async () => {
      const { wrapper } = factory({
        isNewIdp: false,
        idpResources: {
          resourceName: SamlCreationDataMetadataXmlTypeEnum.Url,
          resource: "https://idp.example.com/metadata",
        },
      });
      await flushPromises();

      expect((wrapper.vm as unknown as { samlFields: { metadataXmlUrl?: string } }).samlFields.metadataXmlUrl).toBe(
        "https://idp.example.com/metadata",
      );
    });

    it("leaves the URL field empty when the stored resource of a URL-sourced IdP is the metadata XML", async () => {
      const { wrapper } = factory({
        isNewIdp: false,
        idpResources: { resourceName: SamlCreationDataMetadataXmlTypeEnum.Url, resource: fakeMetadataXml },
      });
      await flushPromises();

      const vm = wrapper.vm as unknown as { samlFields: { metadataXmlUrl?: string }; isValidInput: () => boolean };
      expect(vm.samlFields.metadataXmlUrl).toBeFalsy();
      expect(vm.isValidInput()).toBe(true);
    });

    it("still requires a URL for a new IdP", async () => {
      const { wrapper } = factory({ isNewIdp: true });
      (wrapper.vm as unknown as { samlSource: string }).samlSource = SamlCreationDataMetadataXmlTypeEnum.Url;
      await flushPromises();

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(false);
    });

    it("does not report changes for an untouched, existing file-sourced IdP right after mount", async () => {
      const { wrapper } = factory({
        isNewIdp: false,
        idpResources: { resourceName: "idp.xml", resource: fakeMetadataXml },
      });
      await flushPromises();

      expect((wrapper.vm as unknown as { hasChanges: () => boolean }).hasChanges()).toBe(false);
    });
  });

  describe("stale metadata validation guard", () => {
    it("ignores a dry-run result that resolves after a newer file selection has started", async () => {
      let resolveStale: (value: IdpPostResponse | PromiseLike<IdpPostResponse>) => void = () => {};
      vi.mocked(idpService.dryRunIdp).mockReturnValue(
        new Promise((resolve) => {
          resolveStale = resolve;
        }),
      );

      const { wrapper } = factory({ isNewIdp: true });
      const vm = wrapper.vm as unknown as {
        setFileContent: (n: string, c: string, t?: number) => Promise<void>;
        fileSelectionToken: number;
        expirationDate: Date | null;
      };

      const firstCall = vm.setFileContent("first.xml", fakeMetadataXml, 0);
      vm.fileSelectionToken = 1;

      resolveStale(acceptedMetadata);
      await firstCall;
      await wrapper.vm.$nextTick();

      expect(vm.expirationDate).toBeNull();
      expect(wrapper.text()).not.toContain("Certificate expired");
    });
  });

  describe("isValidInput", () => {
    it("is invalid when identity-manager rejects the uploaded file", async () => {
      vi.mocked(idpService.dryRunIdp).mockRejectedValue(rejectedMetadata);

      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(false);
    });

    it("validates an existing identity provider through the update dry-run", async () => {
      const { wrapper } = factory({ isNewIdp: false, idpAlias: "saml-idp" });
      await uploadFile(wrapper);

      expect(idpService.dryRunIdp).toHaveBeenCalledWith(
        expect.objectContaining({
          type: IdpCreationRequestTypeEnum.Saml,
          dryRun: true,
        }),
        "saml-idp",
      );
    });

    it("accepts a metadata URL without asking identity-manager to validate it", async () => {
      const { wrapper } = factory({ isNewIdp: true });
      (wrapper.vm as unknown as { samlSource: string }).samlSource = SamlCreationDataMetadataXmlTypeEnum.Url;
      await flushPromises();

      (wrapper.vm as unknown as { updateMetadataXmlUrl: (url: string) => void }).updateMetadataXmlUrl(
        "https://idp.example.com/metadata",
      );

      expect(idpService.dryRunIdp).not.toHaveBeenCalled();
      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(true);
    });

    it("is valid when identity-manager accepts the uploaded file", async () => {
      const { wrapper } = factory({ isNewIdp: true });
      await uploadFile(wrapper);

      expect((wrapper.vm as unknown as { isValidInput: () => boolean }).isValidInput()).toBe(true);
    });
  });
});
