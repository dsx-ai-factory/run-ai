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

const REJECTED_MESSAGE = "invalid saml signing certificate";

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
      samlData: { signingCertificate: "fake-cert" },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("shows the identity-manager error and blocks the form when metadata is rejected", async () => {
    vi.mocked(idpService.dryRunIdp).mockRejectedValue({ message: REJECTED_MESSAGE });
    const wrapper = mountInForm();
    await uploadMetadataFile(wrapper);

    expect(wrapper.text()).toContain(REJECTED_MESSAGE);
    expect(await validateForm(wrapper)).toBe(false);
    wrapper.unmount();
  });

  it("accepts the form when identity-manager accepts the metadata", async () => {
    const wrapper = mountInForm();
    await uploadMetadataFile(wrapper);

    expect(wrapper.text()).not.toContain(REJECTED_MESSAGE);
    expect(await validateForm(wrapper)).toBe(true);
    wrapper.unmount();
  });
});
