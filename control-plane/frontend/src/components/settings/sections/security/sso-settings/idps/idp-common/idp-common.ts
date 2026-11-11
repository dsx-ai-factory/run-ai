/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IWrapperBoxSettingMessages } from "@/models/setting.model";
import { idpService } from "@/services/control-plane/idp.service/idp.service";
import { SamlCreationDataMetadataXmlTypeEnum } from "@/swagger-models/identity-manager-client";
import { dateUtil, TimeUnit } from "@/utils/date.util";
import type { ISelectOption } from "@/models/global.model";

export const idpCommon = {
  getSettingsBoxMessages,
  validateSamlMetadata,
  parseExpirationDate,
  iValidCertificateDate,
  isCertificateExpiringSoon,
  getDefaultScope,
};

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

export interface SamlMetadataValidation {
  expirationDate: Date | null;
  error: string;
}

async function validateSamlMetadata(
  metadataXmlFile: string,
  fileName: string,
  idpAlias?: string,
): Promise<SamlMetadataValidation> {
  try {
    const { samlData, errorMessage } = await idpService.validateSamlCertificate(
      { metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.File, metadataXmlFile, fileName },
      idpAlias,
    );
    const expirationDate = parseExpirationDate(samlData?.certificateExpirationDate);
    // An expired certificate is reported by the date check; any other rejection (e.g. not yet valid) keeps its error.
    const isExpired = !!expirationDate && !iValidCertificateDate(expirationDate);
    return {
      expirationDate,
      error: isExpired ? "" : errorMessage || (expirationDate ? "" : "Could not read the certificate expiration date."),
    };
  } catch (error: unknown) {
    return {
      expirationDate: null,
      error: error instanceof Error ? error.message : "Could not validate the certificate.",
    };
  }
}

// The backend sends Go's zero time (year 1) when it cannot parse the certificate, which means "unknown".
function parseExpirationDate(value?: string | null): Date | null {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) || date.getFullYear() <= 1 ? null : date;
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
