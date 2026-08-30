import type { ResponseType } from "axios";

export interface IHTTPOptions {
  responseType: ResponseType;
  headers?: RequestHeaders;
  useIdToken?: boolean;
}

export interface RequestHeaders {
  [key: string]: string | undefined;
  "content-type": string;
  Authorization?: string;
  AuthInfo?: string;
  "Realm-Name"?: string;
  validation?: string;
}

export abstract class HttpErrorResponse extends Error {
  statusCode: number;
  type: string | undefined;
  constructor(type: string, statusCode: number, message: string) {
    super(message);
    this.type = type;
    this.statusCode = statusCode;
  }

  serialize(): string {
    return `[status-code: ${this.statusCode}] [type: ${this.type}] message: ${this.message}`;
  }
}

// 4XX
export class BadRequestError extends HttpErrorResponse {
  static statusCode = 400;

  constructor(message = "Bad request") {
    super("BadRequest", BadRequestError.statusCode, message);
  }
}

export class UnauthorizedError extends HttpErrorResponse {
  static statusCode = 401;
  type = "Unauthorized";

  constructor(message = "Unauthorized") {
    super("Unauthorized", UnauthorizedError.statusCode, message);
  }
}

export class ForbiddenError extends HttpErrorResponse {
  static statusCode = 403;

  constructor(message = "Forbidden: You don't have permission to perform this action.") {
    super("Forbidden", ForbiddenError.statusCode, message);
  }
}

export class NotFoundError extends HttpErrorResponse {
  static statusCode = 404;

  constructor(message = "Not Found") {
    super("NotFound", NotFoundError.statusCode, message);
  }
}

export class ConflictError extends HttpErrorResponse {
  static statusCode = 409;

  constructor(message = "Conflict") {
    super("Conflict", ConflictError.statusCode, message);
  }
}

export class PreconditionFailedError extends HttpErrorResponse {
  static statusCode = 412;

  constructor(message = "Precondition Failed") {
    super("PreconditionFailed", PreconditionFailedError.statusCode, message);
  }
}

// 5XX
export class InternalServerError extends HttpErrorResponse {
  static statusCode = 500;

  constructor(message = "Internal Server error") {
    super("InternalServerError", InternalServerError.statusCode, message);
  }
}

export class NotImplementedError extends HttpErrorResponse {
  static statusCode = 501;

  constructor(message = "Not Implemented error") {
    super("NotImplementedError", NotImplementedError.statusCode, message);
  }
}

export class BadGatewayError extends HttpErrorResponse {
  static statusCode = 502;

  constructor(message = "Bad Gateway error") {
    super("BadGatewayError", BadGatewayError.statusCode, message);
  }
}

export class ServiceUnavailableError extends HttpErrorResponse {
  static statusCode = 503;

  constructor(message = "Service Unavailable error") {
    super("ServiceUnavailableError", ServiceUnavailableError.statusCode, message);
  }
}

export class GatewayTimeoutError extends HttpErrorResponse {
  static statusCode = 504;

  constructor(message = "Gateway Timeout error") {
    super("GatewayTimeoutError", GatewayTimeoutError.statusCode, message);
  }
}

// general
export class UnknownError extends HttpErrorResponse {
  constructor(statusCode: number, message = "Unknown error") {
    super("UnknownError", statusCode, message);
  }
}
