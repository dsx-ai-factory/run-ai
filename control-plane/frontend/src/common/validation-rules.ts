/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { errorMessages } from "./error-message.constant";
import {
  isNotEmpty,
  isValidChartUrl,
  isValidDirectoryPath,
  isValidDockerServerUrl,
  isValidNgcApiKey,
  isValidPrivateUrl,
  isValidUrl,
} from "./form.validators";

export const validationRules = {
  validateValueEntered,
  validateValueSelected,
  validateUrlNotEmpty,
  validateUrlFormat,
  validateChartUrlFormat,
  validatePathNotEmpty,
  validatePathFormat,
  validateNameNotEmpty,
  validateNameSelected,
  validateServerNotEmpty,
  validateKeyNotEmpty,
  validateSecretNotEmpty,
  validatePasswordNotEmpty,
  validateUsernameNotEmpty,
  validateDockerServerFormat,
  validatePrivateUrlFormat,
  validateFormIncomplete,
  validateGenericToken,
};

function validateValueEntered(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.VALUE_NOT_EMPTY;
}

function validateValueSelected(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.SELECT_A_VALUE;
}

function validateUrlNotEmpty(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.URL_NOT_EMPTY;
}

function validateUrlFormat(value: string) {
  return isValidUrl(value) || errorMessages.INVALID_URL;
}

function validateChartUrlFormat(value: string) {
  return isValidChartUrl(value) || errorMessages.INVALID_URL;
}

function validatePrivateUrlFormat(value: string) {
  return isValidPrivateUrl(value) || errorMessages.INVALID_URL;
}

function validatePathNotEmpty(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.PATH_NOT_EMPTY;
}

function validatePathFormat(value: string): string | boolean {
  return isValidDirectoryPath(value) || errorMessages.INVALID_DIRECTORY_PATH;
}

function validateNameNotEmpty(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.NAME_NOT_EMPTY;
}

function validateNameSelected(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.SELECT_A_NAME;
}

function validateServerNotEmpty(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.SERVER_NAME_OR_IP_EMPTY;
}

function validateKeyNotEmpty(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.KEY_NOT_EMPTY;
}

function validateSecretNotEmpty(value: string): string | boolean {
  return isNotEmpty(value) || errorMessages.SECRET_NOT_EMPTY;
}

function validatePasswordNotEmpty(val: string): boolean | string {
  return isNotEmpty(val) || errorMessages.PASSWORD_NOT_EMPTY;
}
function validateUsernameNotEmpty(val: string): boolean | string {
  return isNotEmpty(val) || errorMessages.USERNAME_NOT_EMPTY;
}

function validateDockerServerFormat(value: string): string | boolean {
  return isValidDockerServerUrl(value) || errorMessages.INVALID_URL;
}

function validateFormIncomplete(val: boolean): boolean | string {
  return !val ? true : errorMessages.FORM_INCOMPLETE;
}

function validateGenericToken(val: string): boolean | string {
  return isValidNgcApiKey(val) || errorMessages.ENTER_A_VALID_TOKEN;
}
