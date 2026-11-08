/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { QForm, Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { h, markRaw } from "vue";
import SamlMetadata from "./saml-metadata.vue";
import SamlFields from "@/components/settings/sections/security/sso-settings/idps/saml/saml-fields/saml-fields.vue";
import { idpService } from "@/services/control-plane/idp.service/idp.service";

vi.mock("@/services/control-plane/idp.service/idp.service", () => ({
  idpService: { dryRunIdp: vi.fn() },
}));

const metadataXml =
  '<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"><md:IDPSSODescriptor><md:KeyDescriptor use="signing">' +
  '<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:X509Data><ds:X509Certificate>fake-cert</ds:X509Certificate>' +
  "</ds:X509Data></ds:KeyInfo></md:KeyDescriptor></md:IDPSSODescriptor></md:EntityDescriptor>";

const EXPIRED_MESSAGE = "Certificate expired. Replace the metadata file above.";
const validExpirationDate = "2030-05-17T00:00:00.000Z";
const expiredExpirationDate = "2020-01-01T00:00:00.000Z";

describe("saml-metadata.vue (onboarding)", () => {
  // Onboarding renders every step inside one q-form and validates it when moving to the next step.
  const mountInForm = () => {
    setActivePinia(createTestingPinia({ createSpy: vi.fn }));
    return mount(QForm, {
      global: { plugins: [Quasar] },
      slots: { default: () => h(SamlMetadata, { mappers: {} }) },
      attachTo: document.body,
    });
  };

  const uploadMetadataFile = async (wrapper: ReturnType<typeof mountInForm>) => {
    const fields = wrapper.findComponent(SamlFields);
    (fields.vm as unknown as { metadataFile: File | null }).metadataFile = markRaw(new File([metadataXml], "idp.xml"));
    await fields.vm.$nextTick();
    await (fields.vm as unknown as { setFileContent: (n: string, c: string) => Promise<void> }).setFileContent(
      "idp.xml",
      metadataXml,
    );
    await flushPromises();
  };

  const validateForm = (wrapper: ReturnType<typeof mountInForm>) =>
    (wrapper.vm as unknown as { validate: () => Promise<boolean> }).validate();

  beforeEach(() => {
    vi.mocked(idpService.dryRunIdp).mockResolvedValue({
      alias: "saml",
      samlData: { signingCertificate: "fake-cert", certificateExpirationDate: validExpirationDate },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("shows the certificate error and blocks the form when the uploaded certificate is expired", async () => {
    vi.mocked(idpService.dryRunIdp).mockRejectedValue({
      message: "invalid saml signing certificate",
      samlData: { certificateExpirationDate: expiredExpirationDate },
    });
    const wrapper = mountInForm();
    await uploadMetadataFile(wrapper);

    expect(wrapper.text()).toContain(EXPIRED_MESSAGE);
    expect(wrapper.text()).toContain("Certificate expiration date");
    expect(await validateForm(wrapper)).toBe(false);
    wrapper.unmount();
  });

  it("shows the expiration date and accepts the form when identity-manager accepts the metadata", async () => {
    const wrapper = mountInForm();
    await uploadMetadataFile(wrapper);

    expect(wrapper.text()).toContain("Certificate expiration date");
    expect(wrapper.text()).not.toContain("Certificate expired");
    expect(await validateForm(wrapper)).toBe(true);
    wrapper.unmount();
  });
});
