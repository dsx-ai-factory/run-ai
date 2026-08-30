import {
  BadRequestError,
  NotFoundError,
  InternalServerError,
  NotImplementedError,
  BadGatewayError,
  GatewayTimeoutError,
} from "@/models/http-response.model";
import { ErrorAlert } from "./error-alert.util";
import { describe, beforeEach, vi, it, expect } from "vitest";

describe("error-alert.util", () => {
  beforeEach(() => {
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  const options = {
    generalMessage: "An error occurred",
  };

  const errorAlert = new ErrorAlert(options);

  it("should generate a notification for HTTP error", () => {
    const error = new BadRequestError("Resource not found");
    const notification = errorAlert.getNotification(error);

    expect(notification.message).toEqual("Resource not found");
  });

  it("should generate a notification for non-HTTP error", () => {
    const error = new Error("Some non-HTTP error");
    const notification = errorAlert.getNotification(error);

    expect(notification.message).toEqual("An error occurred");
  });

  it("should generate a notification for non-HTTP error with custom general message", () => {
    const customOptions = {
      generalMessage: "A different error occurred",
    };
    const customErrorAlert = new ErrorAlert(customOptions);
    const error = new Error("Some non-HTTP error");
    const notification = customErrorAlert.getNotification(error);

    expect(notification.message).toEqual("A different error occurred");
  });

  it("should generate a notification for string error", () => {
    const error = "This is a string error";
    const notification = errorAlert.getNotification(error);

    expect(notification.message).toEqual("An error occurred");
  });

  it("should generate a notification for HTTP error with custom message", () => {
    const customOptions = {
      generalMessage: "A different error occurred",
      404: "Resource not found",
    };
    const customErrorAlert = new ErrorAlert(customOptions);

    let error = new NotFoundError("HTTP error");
    let notification = customErrorAlert.getNotification(error);

    expect(notification.message).toEqual("Resource not found");

    error = new BadRequestError("Bad Request");
    notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("Bad Request");
  });

  it("should generate a notification for HTTP error general message for error codes >= 500", () => {
    const customErrorAlert = new ErrorAlert(options);

    let error = new InternalServerError();
    let notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("An error occurred");

    error = new NotImplementedError();
    notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("An error occurred");

    error = new BadGatewayError();
    notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("An error occurred");

    error = new GatewayTimeoutError();
    notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("An error occurred");

    error = new InternalServerError();
    notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("An error occurred");

    error = new InternalServerError("There are issues with your connection to the cluster");
    notification = customErrorAlert.getNotification(error);
    expect(notification.message).toEqual("There are issues with your connection to the cluster");
  });
});
