/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IWrapperBoxSettingMessages } from "@/models/setting.model";
import { certificateUtil } from "@/utils/certificate.util";
import { dateUtil, TimeUnit } from "@/utils/date.util";
import type { ISelectOption } from "@/models/global.model";

export const idpCommon = {
  getSettingsBoxMessages,
  getCertificateExpirationDate,
  iValidCertificateDate,
  isCertificateExpiringSoon,
  getDefaultScope,
  isSamlSigningDisabled,
};

export const SAML_SIGNING_DISABLED_TITLE = "SAML signing is disabled.";
export const SAML_SIGNING_DISABLED_MESSAGE =
  "Edit the identity provider and replace the metadata XML with an updated file.";

const DISABLED_SAML_SIGNING_ATTRIBUTES = ["WantAssertionsSigned", "AuthnRequestsSigned", "WantAuthnRequestsSigned"];

export const DEFAULT_METADATA_FILE_NAME = "metadata.xml";
export const DEFAULT_METADATA_FILE_TYPE = "application/xml";
export const CERTIFICATE_EXPIRING_SOON_THRESHOLD_DAYS = 60;
export const OPENID_SCOPE = "openid";
export const DEFAULT_SCOPE_OPTIONS: ISelectOption = { label: OPENID_SCOPE, value: OPENID_SCOPE, disable: true };
export const SCOPE_REGEX = /^([a-zA-Z0-9-._:]+)(\s[a-zA-Z0-9-._:]+)*$/;
function getSettingsBoxMessages(idpName: string, isSSOUser: boolean): IWrapperBoxSettingMessages {
  return {
    headerTitle: `Identity provider - ${idpName}`,
    editToolTip: "Edit identity provider",
    removeToolTip: isSSOUser
      ? "To remove the identity provider, first sign in as a local user to avoid losing access"
      : "Remove identity provider",
    removeDisable: "To remove the identity provider, first sign in as a local user to avoid losing access",
    confirmRemoveTitle: "Remove identity provider?",
    confirmRemoveSubTitle: "Single Sign-On will be disabled and only local users will be able to sign in.",
  };
}

function getCertificateExpirationDate(certificate: string): Promise<Date | null> {
  return certificateUtil.getExpirationDate(certificate);
}

function iValidCertificateDate(expirationDate: Date): boolean {
  return dateUtil.differenceBy(TimeUnit.minute, expirationDate as Date, new Date()) > 0;
}

function isCertificateExpiringSoon(expirationDate: Date): boolean {
  return (
    iValidCertificateDate(expirationDate) &&
    dateUtil.differenceBy(TimeUnit.day, expirationDate, new Date()) <= CERTIFICATE_EXPIRING_SOON_THRESHOLD_DAYS
  );
}

function isSamlSigningDisabled(xmlContent: string): boolean {
  try {
    const doc = new DOMParser().parseFromString(xmlContent, "text/xml");
    if (doc.querySelector("parsererror")) return false;
    const elements = doc.getElementsByTagName("*");
    for (let i = 0; i < elements.length; i++) {
      for (const name of DISABLED_SAML_SIGNING_ATTRIBUTES) {
        if (elements[i].getAttribute(name)?.toLowerCase() === "false") return true;
      }
    }
    return false;
  } catch {
    return false;
  }
}

function getDefaultScope(scopes: string | string[] | undefined, withOpenId: boolean): Array<string> {
  let scope = withOpenId ? [OPENID_SCOPE] : [];

  if (scopes) {
    if (typeof scopes === "string") {
      scope = scopes.split(" ");
    } else if (Array.isArray(scopes)) {
      scope = [...scopes];
    }
  }

  if (withOpenId && !scope.includes(OPENID_SCOPE)) {
    scope.unshift(OPENID_SCOPE);
  }

  return scope;
}
