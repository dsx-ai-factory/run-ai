/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { h } from "vue";
import { flushPromises, mount } from "@vue/test-utils";
import { QForm, Quasar } from "quasar";
import SamlIdpDetails from "./saml-idp-details.vue";

type Props = InstanceType<typeof SamlIdpDetails>["$props"];

const createWrapper = (props: Props = {}) => mount(SamlIdpDetails, { global: { plugins: [Quasar] }, props });

describe("saml-idp-details.vue", () => {
  it("shows the identity provider URL, entity ID and expiration date", () => {
    const wrapper = createWrapper({
      singleSignOnServiceUrl: "https://idp.example.com/sso",
      idpEntityId: "idp-entity",
      expirationDate: new Date("2030-05-17T00:00:00.000Z"),
      showExpirationDate: true,
    });

    expect(wrapper.text()).toContain("Identity provider URL");
    expect(wrapper.text()).toContain("Identity provider Entity ID (issuer)");
    expect(wrapper.text()).toContain("Certificate expiration date");
    expect(wrapper.find<HTMLInputElement>("input[value='https://idp.example.com/sso']").exists()).toBe(true);
    expect(wrapper.find<HTMLInputElement>("input[value='idp-entity']").exists()).toBe(true);
  });

  it("shows the certificate expiration date first, then the URL and the entity ID", () => {
    const wrapper = createWrapper({
      singleSignOnServiceUrl: "https://idp.example.com/sso",
      idpEntityId: "idp-entity",
      expirationDate: new Date("2030-05-17T00:00:00.000Z"),
      showExpirationDate: true,
    });

    const text = wrapper.text();
    const certificate = text.indexOf("Certificate expiration date");
    const url = text.indexOf("Identity provider URL");
    const entityId = text.indexOf("Identity provider Entity ID (issuer)");

    expect(certificate).toBeGreaterThanOrEqual(0);
    expect(certificate).toBeLessThan(url);
    expect(url).toBeLessThan(entityId);
  });

  it("hides the fields that have no value", () => {
    const wrapper = createWrapper({ singleSignOnServiceUrl: "https://idp.example.com/sso" });

    expect(wrapper.text()).toContain("Identity provider URL");
    expect(wrapper.text()).not.toContain("Identity provider Entity ID (issuer)");
    expect(wrapper.text()).not.toContain("Certificate expiration date");
  });

  it("shows the expiration error message", () => {
    const wrapper = createWrapper({
      showExpirationDate: true,
      expirationDate: new Date("2020-01-01T00:00:00.000Z"),
      expirationError: "Certificate expired. Replace the metadata file above.",
    });

    expect(wrapper.text()).toContain("Certificate expired. Replace the metadata file above.");
  });

  it("shows no error when there is none", () => {
    const wrapper = createWrapper({ showExpirationDate: true, expirationDate: new Date("2030-05-17T00:00:00.000Z") });

    expect(wrapper.text()).not.toContain("Certificate expired");
  });
  describe("certificate expiration field in a form", () => {
    const mountInForm = (props: Props) =>
      mount(QForm, {
        global: { plugins: [Quasar] },
        slots: { default: () => h(SamlIdpDetails, props) },
        attachTo: document.body,
      });
    const validate = (wrapper: ReturnType<typeof mountInForm>) =>
      (wrapper.vm as unknown as { validate: () => Promise<boolean> }).validate();
    const certificateInput = (wrapper: ReturnType<typeof mountInForm>) =>
      wrapper.find<HTMLInputElement>("input[aria-label='Certificate expiration date'], input[type='text']");

    it("is disabled and does not block the form while the certificate is valid", async () => {
      const wrapper = mountInForm({ showExpirationDate: true, expirationDate: new Date("2030-05-17T00:00:00.000Z") });
      await flushPromises();

      expect(certificateInput(wrapper).attributes("disabled")).toBeDefined();
      expect(await validate(wrapper)).toBe(true);
      wrapper.unmount();
    });

    it("is active and blocks the form when the certificate is expired", async () => {
      const wrapper = mountInForm({
        showExpirationDate: true,
        expirationDate: new Date("2020-01-01T00:00:00.000Z"),
        expirationError: "Certificate expired. Replace the metadata file above.",
      });
      await flushPromises();

      expect(certificateInput(wrapper).attributes("disabled")).toBeUndefined();
      expect(await validate(wrapper)).toBe(false);
      expect(wrapper.text()).toContain("Certificate expired. Replace the metadata file above.");
      wrapper.unmount();
    });
  });
});
