/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  BadRequestError,
  UnauthorizedError,
  ForbiddenError,
  NotFoundError,
  ConflictError,
  PreconditionFailedError,
  InternalServerError,
  NotImplementedError,
  BadGatewayError,
  ServiceUnavailableError,
  GatewayTimeoutError,
  UnknownError,
  type HttpErrorResponse,
} from "@/models/http-response.model";
import type { SubmissionError } from "@/swagger-models/authorization-client";

export const httpResponseService = { getError, isDefaultMessage, getSubmissionError };

function getError(status: number, message?: string): HttpErrorResponse {
  if (!status) return new UnknownError(status);

  let error;
  switch (status) {
    case BadRequestError.statusCode:
      error = new BadRequestError(message);
      break;
    case UnauthorizedError.statusCode:
      error = new UnauthorizedError(message);
      break;
    case ForbiddenError.statusCode:
      error = new ForbiddenError(message);
      break;
    case NotFoundError.statusCode:
      error = new NotFoundError(message);
      break;
    case ConflictError.statusCode:
      error = new ConflictError(message);
      break;
    case PreconditionFailedError.statusCode:
      error = new PreconditionFailedError(message);
      break;
    case InternalServerError.statusCode:
      error = new InternalServerError(message);
      break;
    case NotImplementedError.statusCode:
      error = new NotImplementedError(message);
      break;
    case BadGatewayError.statusCode:
      error = new BadGatewayError(message);
      break;
    case ServiceUnavailableError.statusCode:
      error = new ServiceUnavailableError(message);
      break;
    case GatewayTimeoutError.statusCode:
      error = new GatewayTimeoutError(message);
      break;
    default:
      error = new UnknownError(status, message);
  }

  return error;
}

function isDefaultMessage(statusCode: number, message: string): boolean {
  const error: HttpErrorResponse = getError(statusCode);
  return error.message === message;
}

function getSubmissionError(err: SubmissionError): HttpErrorResponse {
  const errorObj: HttpErrorResponse = getError(err.code);
  if (err.complianceIssues?.length) {
    errorObj.message = `${err.complianceIssues[0].details} (${err.complianceIssues[0].field})`;
  }
  return errorObj;
}
