/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
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
});
