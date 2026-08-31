/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest"; // Replace 'yourModule' with the actual module path
import type { Mappers, OidcCreationData, SamlCreationData } from "@/swagger-models/identity-manager-client";
import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import { settingsUtil } from "@/utils/settings.util/settings.util";
import type { ISettingEmailServer } from "@/models/setting.model";
import { EmailAuthType } from "@/swagger-models/notifications-proxy-client";
import type { EmailChannelConfig } from "@/swagger-models/notifications-proxy-client";

describe("settingsUtil", () => {
  describe("getIdpMappersItems", () => {
    it("should return an array of IIdpMapperItem with correct values", () => {
      const mockMappers: Mappers = {
        uid: "456",
        gid: "123",
        groups: "admins",
        supplementaryGroups: "additionalGroups",
        email: "email",
        firstName: "given_name",
        lastName: "family_name",
      };

      const result = settingsUtil.getIdpMappersItems(mockMappers);

      expect(result).toEqual({
        uid: { key: "uid", label: "Linux user ID", value: "456" },
        gid: { key: "gid", label: "Linux group ID", value: "123" },
        groups: { key: "groups", label: "User role groups", value: "admins" },
        supplementaryGroups: { key: "supplementaryGroups", label: "SUPPLEMENTARYGROUPS", value: "additionalGroups" },
        email: { key: "email", label: "Email", value: "email" },
        firstName: { key: "firstName", label: "First name", value: "given_name" },
        lastName: { key: "lastName", label: "Last name", value: "family_name" },
      });
    });
  });

  describe("isIdpMappersEqual", () => {
    it("should return true when mappers and items have the same values", () => {
      const mockMappers: Mappers = {
        uid: "456",
        gid: "123",
        groups: "admins",
        supplementaryGroups: "additionalGroups",
        email: "email",
        firstName: "given_name",
        lastName: "family_name",
      };

      const mockItems = {
        uid: { key: "uid", label: "Linux user ID", value: "456" },
        gid: { key: "gid", label: "Linux group ID", value: "123" },
        groups: { key: "groups", label: "User role groups", value: "admins" },
        supplementaryGroups: { key: "supplementaryGroups", label: "SUPPLEMENTARYGROUPS", value: "additionalGroups" },
        email: { key: "email", label: "Email", value: "email" },
        firstName: { key: "firstName", label: "First name", value: "given_name" },
        lastName: { key: "lastName", label: "Last name", value: "family_name" },
      };

      const result = settingsUtil.isIdpMappersEqual(mockMappers, mockItems);

      expect(result).toBe(true);
    });

    it("should return false when mappers and items have different values", () => {
      const mockMappers: Mappers = {
        uid: "456",
        gid: "123",
        groups: "admins",
        supplementaryGroups: "additionalGroups",
        email: "email",
        firstName: "given_name",
        lastName: "family_name",
      };

      const mockItems = {
        uid: { key: "uid", label: "Linux user ID", value: "456" },
        gid: { key: "gid", label: "Linux group ID", value: "456" },
        groups: { key: "groups", label: "User role groups", value: "admins" },
        supplementaryGroups: { key: "supplementaryGroups", label: "SUPPLEMENTARYGROUPS", value: "additionalGroups" },
        email: { key: "email", label: "Email", value: "email" },
        firstName: { key: "firstName", label: "First name", value: "given_name" },
        lastName: { key: "lastName", label: "Last name", value: "family_name" },
      };

      const result = settingsUtil.isIdpMappersEqual(mockMappers, mockItems);

      expect(result).toBe(false);
    });
  });

  describe("getMappersFromIdpMapperItems", () => {
    it("should return Mappers object with correct values from IdpMapperItems", () => {
      const mockItems = {
        uid: { key: "uid", label: "Linux user ID", value: "456" },
        gid: { key: "gid", label: "Linux group ID", value: "123" },
        groups: { key: "groups", label: "User role groups", value: "admins" },
        supplementaryGroups: { key: "supplementaryGroups", label: "SUPPLEMENTARYGROUPS", value: "additionalGroups" },
        email: { key: "email", label: "Email", value: "email" },
        firstName: { key: "firstName", label: "First name", value: "given_name" },
        lastName: { key: "lastName", label: "Last name", value: "family_name" },
      };

      const result: Mappers = settingsUtil.getMappersFromIdpMapperItems(mockItems);

      expect(result).toEqual({
        uid: "456",
        gid: "123",
        groups: "admins",
        supplementaryGroups: "additionalGroups",
        email: "email",
        firstName: "given_name",
        lastName: "family_name",
      });
    });
  });
  describe("createIdpCreationRequest", () => {
    const oidcData: OidcCreationData = {
      discoverDocumentUrl: "https://test.run.ai/auth/realms/workAI/.well-known/openid-configuration",
      clientId: "my-client-id",
      clientSecret: "test",
    };
    const samlData: SamlCreationData = {
      metadataXmlUrl: "https://test.run.ai/auth/realms/workAI/protocol/saml/descriptor",
    };
    const mappers: Mappers = {
      uid: "456",
      gid: "123",
      groups: "admins",
      supplementaryGroups: "additionalGroups",
      email: "email",
    };

    it("should return an object for Oidc type", () => {
      const result = settingsUtil.createIdpCreationRequest(IdpTypeEnum.Oidc, oidcData, mappers);
      expect(result).toEqual({
        type: IdpTypeEnum.Oidc,
        oidcData: oidcData,
        mappers: mappers,
      });
    });

    it("should return an object for Saml type", () => {
      const result = settingsUtil.createIdpCreationRequest(IdpTypeEnum.Saml, samlData, mappers);
      expect(result).toEqual({
        type: IdpTypeEnum.Saml,
        samlData: samlData,
        mappers: mappers,
      });
    });

    it("should return null for unsupported type", () => {
      // Mock any dependencies if needed

      const unsupportedType = "UnsupportedType";
      const result = settingsUtil.createIdpCreationRequest(unsupportedType, oidcData, mappers);
      expect(result).toBeNull();
    });
  });

  describe("emailSettingsToSmtpConfig", () => {
    const emailSettings: ISettingEmailServer = {
      smtpHost: "smtp.example.com",
      smtpPort: 587,
      username: "test@example.com",
      password: "password123",
      fromEmail: "noreply@example.com",
      fromName: "Example Sender",
      authType: EmailAuthType.AuthPlain,
    };
    it("should convert email settings to SMTP config", () => {
      const result = settingsUtil.emailSettingsToSmtpConfig(emailSettings);

      expect(result).toEqual({
        username: "test@example.com",
        password: "password123",
        smtpHost: "smtp.example.com",
        from: "noreply@example.com",
        smtpPort: 587,
        authType: EmailAuthType.AuthPlain,
        smtpTlsEnabled: true,
        recipients: [],
        fromDisplayName: "Example Sender",
        directNotifications: true,
        subject: "{{ .title }}",
      });
    });
  });

  describe("smtpConfigToEmailSettings", () => {
    const emailChannelConfig: EmailChannelConfig = {
      username: "test@example.com",
      password: "password123",
      smtpHost: "smtp.example.com",
      from: "noreply@example.com",
      smtpPort: 587,
      authType: EmailAuthType.AuthPlain,
      smtpTlsEnabled: true,
      recipients: [],
      fromDisplayName: "Example Sender",
      directNotifications: true,
      subject: "{{ .title }}",
    };

    it("should convert SMTP config to email settings", () => {
      const result = settingsUtil.smtpConfigToEmailSettings(emailChannelConfig);

      expect(result).toEqual({
        smtpHost: "smtp.example.com",
        smtpPort: 587,
        username: "test@example.com",
        password: "password123",
        fromEmail: "noreply@example.com",
        fromName: "Example Sender",
        authType: EmailAuthType.AuthPlain,
      });
    });

    it("should use default values for missing properties", () => {
      const incompleteChannelConfig: EmailChannelConfig = {
        smtpHost: "",
        smtpPort: 0,
        username: "",
        password: "",
        from: "",
        authType: EmailAuthType.AuthPlain,
        smtpTlsEnabled: true,
        recipients: [],
        fromDisplayName: "",
        directNotifications: true,
        subject: "{{ .title }}",
      };

      const result = settingsUtil.smtpConfigToEmailSettings(incompleteChannelConfig);

      expect(result).toEqual({
        smtpHost: "",
        smtpPort: 587, // assuming DEFAULT_SMTP_PORT is 25
        username: "",
        password: "",
        fromEmail: "",
        fromName: "",
        authType: EmailAuthType.AuthPlain,
      });
    });

    it("should getDefaultEntityId", () => {
      const result = settingsUtil.getDefaultEntityId("hostname");
      expect(result).toEqual(`runai-hostname/runai`);
    });
  });

  describe("getToggleSectionCollapsedTitle", () => {
    it("should count boolean true and false as enabled/disabled", () => {
      const settings = {
        a: { key: "a", value: true, label: "A" },
        b: { key: "b", value: false, label: "B" },
      };
      expect(settingsUtil.getToggleSectionCollapsedTitle(settings)).toBe("Enabled (1) / Disabled (1)");
    });

    it("should count string 'true' and 'false' as enabled/disabled", () => {
      const settings = {
        a: { key: "a", value: "true", label: "A" },
        b: { key: "b", value: "false", label: "B" },
      };
      expect(settingsUtil.getToggleSectionCollapsedTitle(settings)).toBe("Enabled (1) / Disabled (1)");
    });

    it("should exclude null and non-boolean-like values from the count", () => {
      const settings = {
        a: { key: "a", value: true, label: "A" },
        b: null,
        c: { key: "c", value: "other", label: "C" },
        d: { key: "d", value: 42, label: "D" },
      };
      expect(settingsUtil.getToggleSectionCollapsedTitle(settings)).toBe("Enabled (1) / Disabled (0)");
    });

    it("should return Enabled (0) / Disabled (0) for empty or non-matching settings", () => {
      expect(settingsUtil.getToggleSectionCollapsedTitle({})).toBe("Enabled (0) / Disabled (0)");
      expect(
        settingsUtil.getToggleSectionCollapsedTitle({
          a: { key: "a", value: "nope", label: "A" },
        }),
      ).toBe("Enabled (0) / Disabled (0)");
    });
  });
});
