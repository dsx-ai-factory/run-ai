/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import DOCS from "@/constants/docs";

export const onboardingUtil = {
  getProtocolDisplayName,
  getSsoDocumentationLink,
};

function getProtocolDisplayName(protocol: string | undefined): string {
  switch (protocol) {
    case IdpTypeEnum.Saml:
      return "SAML";
    case IdpTypeEnum.Oidc:
      return "OpenID Connect";
    case IdpTypeEnum.OpenshiftV4:
      return "OpenShift V4";
    default:
      return "";
  }
}

function getSsoDocumentationLink(protocol: string | undefined): string {
  switch (protocol) {
    case IdpTypeEnum.Saml:
      return DOCS.AUTHENTICATION.SAML;
    case IdpTypeEnum.Oidc:
      return DOCS.AUTHENTICATION.OIDC;
    case IdpTypeEnum.OpenshiftV4:
      return DOCS.AUTHENTICATION.OPENSHIFT;
    default:
      return DOCS.AUTHENTICATION.SSO;
  }
}
