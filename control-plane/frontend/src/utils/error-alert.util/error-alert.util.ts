/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { HttpErrorResponse } from "@/models/http-response.model";
import { alertUtil } from "@/utils/alert.util";
import { httpResponseService } from "@/services/infra/http-response.service/http-response.service";
import { stringUtil } from "@/utils/string.util";
import type { QNotifyCreateOptions } from "quasar";

interface IErrorCodeOptoins {
  generalMessage: string;
  [key: number]: string;
}

interface IApiErrorObject {
  message?: string;
  complianceIssues?: Array<{ details?: string; field?: string }>;
}

/**
 * Extracts error messages from various error formats (axios errors, API errors, Error objects)
 * Handles both single errors and arrays of errors from API responses
 */
export function extractErrorMessages(error: unknown, fallbackMessage = "Operation failed"): string[] {
  // Check if it's an axios error with response data
  const axiosError = error as { response?: { data?: unknown } };
  const responseData = axiosError.response?.data;

  // Handle array of API errors
  if (Array.isArray(responseData)) {
    const messages = (responseData as IApiErrorObject[])
      .map((e) => {
        if (e.complianceIssues?.length) {
          return e.complianceIssues.map((issue) => issue.details || issue.field).join("; ");
        }
        return e.message;
      })
      .filter(Boolean) as string[];

    if (messages.length > 0) {
      return messages;
    }
  }

  // Handle single API error object with complianceIssues
  const singleError = responseData as IApiErrorObject | undefined;
  if (singleError?.complianceIssues?.length) {
    return [singleError.complianceIssues.map((issue) => issue.details || issue.field).join("; ")];
  }
  if (singleError?.message) {
    return [singleError.message];
  }

  // Handle standard Error object
  if (error instanceof Error) {
    return [error.message];
  }

  return [fallbackMessage];
}

export class ErrorAlert {
  options: IErrorCodeOptoins;

  constructor(options: IErrorCodeOptoins) {
    this.options = options;
  }

  getNotification(error: unknown, removePattern?: RegExp): QNotifyCreateOptions {
    if (error instanceof HttpErrorResponse) {
      if (removePattern) {
        error.message = error.message.replace(removePattern, "");
      }
      return this.httpError(error);
    } else {
      return this.nonHttpError(error);
    }
  }

  private httpError(error: HttpErrorResponse): QNotifyCreateOptions {
    console.error(error.serialize());
    if (!error.statusCode || error.statusCode >= 500) {
      if (error.statusCode >= 500 && !httpResponseService.isDefaultMessage(error.statusCode, error.message)) {
        return alertUtil.getError(error.message);
      }

      return alertUtil.getError(this.options.generalMessage);
    } else {
      const message = this.options[error.statusCode] ? this.options[error.statusCode] : error.message;
      return alertUtil.getError(message);
    }
  }

  private nonHttpError(error: unknown): QNotifyCreateOptions {
    if (error instanceof Error) {
      console.error(error.message);
    } else {
      console.error(error);
    }
    return alertUtil.getError(this.options.generalMessage);
  }

  static createMessage(entity: string): string {
    return `Create ${entity} succeeded`;
  }

  static updateMessage(entity: string): string {
    return `Update ${entity} succeeded`;
  }

  static failedCreateMessage(entity: string): string {
    return `${stringUtil.capitalizeString(entity)} couldn't be created`;
  }

  static failedUpdateMessage(entity: string): string {
    return `Update ${entity} failed`;
  }

  static failedDeleteMessage(entity: string): string {
    return `Delete ${entity} failed`;
  }

  static failedLoadingMessage(entity: string): string {
    return `${stringUtil.capitalizeString(entity)} couldn't be loaded`;
  }

  static failedToValidateMessage(entity: string): string {
    return `Failed to validate ${entity}`;
  }

  static conflictMessage(entity: string): string {
    return `${entity} already exists`;
  }

  static failedValidateMessage(entity: string): string {
    return `${entity} validation failed`;
  }

  static failedToExportMessage(entity: string): string {
    return `Failed to export ${entity}`;
  }
}
