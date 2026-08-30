// @ts-nocheck
import { vi, describe, it, expect, beforeEach } from "vitest";
import { notificationService } from "./notification.service";
import { notificationServiceApi } from "@/services/infra/client-apis/notification-service-api/notification-service-api";
import type { Subscription } from "@/swagger-models/notifications-proxy-client";

// Mock the dependencies
vi.mock("@/services/infra/client-apis/notification-service-api/notification-service-api");
vi.mock("@/services/infra/https.service/http.service", () => ({
  httpService: {
    handleHttpError: vi.fn((err) => err),
  },
}));

describe("notificationService", () => {
  beforeEach(() => {
    // Reset any mock implementations before each test
    vi.resetAllMocks();
  });

  it("should get email successfully", async () => {
    const mockEmailChannel = {
      config: {
        username: "test@example.com",
        password: "password",
        smtpHost: "smtp.test.com",
        smtpPort: 587,
      },
    };

    notificationServiceApi.emailApi.getEmail.mockResolvedValue({
      data: mockEmailChannel,
    });

    const result = await notificationService.getEmail();
    expect(result).toEqual(mockEmailChannel);
    expect(notificationServiceApi.emailApi.getEmail).toHaveBeenCalled();
  });

  it("should update email successfully", async () => {
    const mockEmailChannel = {
      config: {
        username: "test@example.com",
        password: "password",
        smtpHost: "smtp.test.com",
        smtpPort: 587,
      },
    };

    notificationServiceApi.emailApi.updateEmail.mockResolvedValue({});

    await notificationService.updateEmail(mockEmailChannel);
    expect(notificationServiceApi.emailApi.updateEmail).toHaveBeenCalledWith(mockEmailChannel);
  });

  it("should create email successfully", async () => {
    const mockEmailChannel = {
      config: {
        username: "test@example.com",
        password: "password",
        smtpHost: "smtp.test.com",
        smtpPort: 587,
      },
    };

    notificationServiceApi.emailApi.createEmail.mockResolvedValue({
      data: mockEmailChannel,
    });

    const result = await notificationService.createEmail(mockEmailChannel);
    expect(result).toEqual(mockEmailChannel);
    expect(notificationServiceApi.emailApi.createEmail).toHaveBeenCalledWith(mockEmailChannel);
  });

  it("should delete email successfully", async () => {
    notificationServiceApi.emailApi.deleteEmail.mockResolvedValue({});

    await notificationService.deleteEmail();
    expect(notificationServiceApi.emailApi.deleteEmail).toHaveBeenCalled();
  });

  it("should validate notification channel successfully", async () => {
    const mockValidationResponse = { status: true };
    const mockValidateRequest = { smtpServer: "smtp.test.com" };

    notificationServiceApi.emailApi.validateEmail.mockResolvedValue({
      data: mockValidationResponse,
    });

    const result = await notificationService.validateEmail(mockValidateRequest);
    expect(result).toBe(true);
    expect(notificationServiceApi.emailApi.validateEmail).toHaveBeenCalledWith(mockValidateRequest);
  });

  it("should get user notification types successfully", async () => {
    const mockNotificationTypes = [{ id: "1", name: "Email" }];

    vi.mocked(notificationServiceApi.notificationTypesApi.getNotificationTypes).mockResolvedValue({
      data: mockNotificationTypes,
    });

    const result = await notificationService.getUserNotificationTypes();
    expect(result).toEqual(mockNotificationTypes);
    expect(notificationServiceApi.notificationTypesApi.getNotificationTypes).toHaveBeenCalled();
  });

  it("should get user subscriptions successfully", async () => {
    const mockSubscriptions: Subscription[] = [{ id: "1", type: "Email", isActive: true }];

    notificationServiceApi.subscriptionApi.getUserSubscriptions.mockResolvedValue({
      data: mockSubscriptions,
    });

    const result = await notificationService.getUserSubscriptions();
    expect(result).toEqual(mockSubscriptions);
    expect(notificationServiceApi.subscriptionApi.getUserSubscriptions).toHaveBeenCalled();
  });

  it("should create user subscription successfully", async () => {
    const mockSubscription: Subscription = { id: "2", type: "SMS", isActive: true };

    notificationServiceApi.subscriptionApi.createSubscription.mockResolvedValue({
      data: mockSubscription,
    });

    const result = await notificationService.createUserSubscription(mockSubscription);
    expect(result).toEqual(mockSubscription);
    expect(notificationServiceApi.subscriptionApi.createSubscription).toHaveBeenCalledWith(mockSubscription);
  });

  it("should update user subscription successfully", async () => {
    const mockSubscription: Subscription = { id: "1", type: "Email", isActive: false };

    notificationServiceApi.subscriptionApi.updateSubscription.mockResolvedValue({
      data: mockSubscription,
    });

    const result = await notificationService.updateUserSubscription("1", mockSubscription);
    expect(result).toEqual(mockSubscription);
    expect(notificationServiceApi.subscriptionApi.updateSubscription).toHaveBeenCalledWith("1", mockSubscription);
  });

  // New Test: isSmtpEnabledForTenant
  it("should check if SMTP is enabled for the tenant", async () => {
    const mockResponse = { data: { enabled: true } };

    notificationServiceApi.emailApi.getEmailNotificationState.mockResolvedValue(mockResponse);

    const result = await notificationService.isSmtpEnabledForTenant();
    expect(result).toBe(true);
    expect(notificationServiceApi.emailApi.getEmailNotificationState).toHaveBeenCalled();
  });

  it("should get detailed slack notification state successfully", async () => {
    const mockResponse = {
      data: { enabled: true, settings: { workspaceName: "workspace-name", appCredentialsConfigured: true } },
    };

    notificationServiceApi.slackApi.getDetailedSlackNotificationState.mockResolvedValue(mockResponse);

    const result = await notificationService.getDetailedSlackNotificationState();
    expect(result.enabled).toBe(true);
    expect(result.settings.workspaceName).toBe("workspace-name");
    expect(result.settings.appCredentialsConfigured).toBe(true);
    expect(notificationServiceApi.slackApi.getDetailedSlackNotificationState).toHaveBeenCalled();
  });

  it("should check if slack is enabled for the tenant", async () => {
    const mockResponse = {
      data: { enabled: true },
    };

    notificationServiceApi.slackApi.getSlackNotificationState.mockResolvedValue(mockResponse);

    const result = await notificationService.getSlackNotificationState();
    expect(result.enabled).toBe(true);
    expect(notificationServiceApi.slackApi.getSlackNotificationState).toHaveBeenCalled();
  });

  it("should get notification channels successfully", async () => {
    const mockChannels = [
      { name: "slack-123", type: "slack" },
      { name: "email-456", type: "email" },
    ];

    notificationServiceApi.notificationChannelsApi.getNotificationChannels.mockResolvedValue({
      data: mockChannels,
    });

    const result = await notificationService.getNotificationChannels();
    expect(result).toEqual(mockChannels);
    expect(notificationServiceApi.notificationChannelsApi.getNotificationChannels).toHaveBeenCalled();
  });

  it("should delete notification channel config successfully", async () => {
    const channelName = "slack-123";
    notificationServiceApi.notificationChannelsApi.deleteNotificationChannel.mockResolvedValue({});

    await notificationService.deleteNotificationChannel(channelName);
    expect(notificationServiceApi.notificationChannelsApi.deleteNotificationChannel).toHaveBeenCalledWith(channelName);
  });

  it("should connect Slack workspace successfully", async () => {
    const mockRedirectUrl = "slack-redirect-url";
    notificationServiceApi.slackApi.connectSlackWorkspace.mockResolvedValue({
      data: { redirectUrl: mockRedirectUrl },
    });

    const result = await notificationService.connectSlackWorkspace();
    expect(result).toBe(mockRedirectUrl);
    expect(notificationServiceApi.slackApi.connectSlackWorkspace).toHaveBeenCalled();
  });

  it("should throw error if connecting Slack workspace fails", async () => {
    notificationServiceApi.slackApi.connectSlackWorkspace.mockRejectedValue(
      new Error("Failed to connect Slack workspace"),
    );

    await expect(notificationService.connectSlackWorkspace()).rejects.toThrow("Failed to connect Slack workspace");
  });

  it("should disconnect Slack workspace successfully", async () => {
    notificationServiceApi.slackApi.disconnectSlackWorkspace.mockResolvedValue({});

    await notificationService.disconnectSlackWorkspace();
    expect(notificationServiceApi.slackApi.disconnectSlackWorkspace).toHaveBeenCalled();
  });

  it("should throw error if disconnecting Slack workspace fails", async () => {
    notificationServiceApi.slackApi.disconnectSlackWorkspace.mockRejectedValue(
      new Error("Failed to disconnect Slack workspace"),
    );

    await expect(notificationService.disconnectSlackWorkspace()).rejects.toThrow("Failed to disconnect Slack workspace");
  });

  it("should create Slack app successfully", async () => {
    const mockAccessToken = "slack-access-token";
    notificationServiceApi.slackApi.createSlackApp.mockResolvedValue({});

    await notificationService.createSlackApp(mockAccessToken);
    expect(notificationServiceApi.slackApi.createSlackApp).toHaveBeenCalledWith({ accessToken: mockAccessToken });
  });

  it("should throw error if creating Slack app fails", async () => {
    notificationServiceApi.slackApi.createSlackApp.mockRejectedValue(new Error("Failed to create Slack app"));

    await expect(notificationService.createSlackApp("slack-access-token")).rejects.toThrow("Failed to create Slack app");
  });

  it("should delete Slack app successfully", async () => {
    notificationServiceApi.slackApi.deleteSlackApp.mockResolvedValue({});

    await notificationService.deleteSlackApp();
    expect(notificationServiceApi.slackApi.deleteSlackApp).toHaveBeenCalled();
  });

  it("should throw error if deleting Slack app fails", async () => {
    notificationServiceApi.slackApi.deleteSlackApp.mockRejectedValue(new Error("Failed to delete Slack app"));

    await expect(notificationService.deleteSlackApp()).rejects.toThrow("Failed to delete Slack app");
  });
});
