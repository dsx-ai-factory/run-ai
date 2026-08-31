/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum ERROR_ROUTE_NAMES {
  NOT_FOUND = "404",
  NO_ROLES = "no-roles",
  NO_SSO_EMAIL_MAPPER = "no-sso-email-mapper",
  BAD_GATEWAY_ERROR = "502",
  INTERNAL_SERVER_ERROR = "500",
  SERVICE_UNAVAILABLE_ERROR = "503",
  MISSING_CLAIM = "missing-claim",
  UNKNOWN_TIMEOUT = "unknown-timeout",
}
