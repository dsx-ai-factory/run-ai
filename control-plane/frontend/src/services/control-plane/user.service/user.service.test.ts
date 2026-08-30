import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { AxiosHeaders } from "axios";
import { userService } from "./user.service";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { BulkUserCreationRequest, BulkUserCreationResponse, User } from "@/swagger-models/identity-manager-client";

vi.mock("@/services/infra/client-apis/identity-manager-api/identity-manager-api");
vi.mock("@/services/infra/https.service/http.service");

describe("userService.isUserExist", () => {
  let spy: ReturnType<typeof vi.spyOn<typeof userService, never>>;

  beforeEach(() => {
    spy = vi.spyOn(userService, "getUsers");
  });

  afterEach(() => {
    spy.mockRestore();
  });

  it("should call getUsers with exact match filter and return true if user exists", async () => {
    const username = "foo@bar.com";
    spy.mockResolvedValue([{ username: "foo@bar.com" } as User]);
    const result = await userService.isUserExist(username);
    expect(spy).toHaveBeenCalledWith([`username=@${username}`]);
    expect(result).toBe(true);
  });

  it("should return false if no user exists", async () => {
    const username = "bar@foo.com";
    spy.mockResolvedValue([]);
    const result = await userService.isUserExist(username);
    expect(result).toBe(false);
  });

  it("should respect case sensitivity when checking user existence", async () => {
    spy.mockResolvedValue([{ username: "Foo@Bar.com" } as User]);
    const result = await userService.isUserExist("foo@bar.com");
    expect(result).toBe(false);
  });

  it("should return true if multiple users are returned and one matches", async () => {
    spy.mockResolvedValue([{ username: "other@user.com" } as User, { username: "foo@bar.com" } as User]);
    const result = await userService.isUserExist("foo@bar.com");
    expect(result).toBe(true);
  });

  it("should call getUsers with correct filter for different usernames", async () => {
    const username = "baz@qux.com";
    spy.mockResolvedValue([{ username: "baz@qux.com" } as User]);
    await userService.isUserExist(username);
    expect(spy).toHaveBeenCalledWith([`username=@${username}`]);
  });
});

describe("userService.createBatch", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should successfully create users in batch", async () => {
    const bulkRequest: BulkUserCreationRequest = {
      users: [{ email: "user1@example.com" }, { email: "user2@example.com" }],
    };

    const mockResponse: BulkUserCreationResponse = {
      results: [
        {
          success: true,
          user: { id: "1", username: "user1@example.com", tempPassword: "temp123" },
        },
        {
          success: true,
          user: { id: "2", username: "user2@example.com", tempPassword: "temp456" },
        },
      ],
      totalRequested: 2,
      totalSuccessful: 2,
      totalFailed: 0,
    };

    vi.mocked(identityManagerApi.usersApi.batchCreateUsers).mockResolvedValue({
      data: mockResponse,
      status: 200,
      statusText: "OK",
      headers: {},
      config: { headers: new AxiosHeaders() },
    });

    const result = await userService.createBatch(bulkRequest);

    expect(identityManagerApi.usersApi.batchCreateUsers).toHaveBeenCalledWith(bulkRequest);
    expect(result).toEqual(mockResponse);
    expect(result.results).toHaveLength(2);
    expect(result.totalSuccessful).toBe(2);
  });

  it("should handle partial success in batch creation", async () => {
    const bulkRequest: BulkUserCreationRequest = {
      users: [{ email: "user1@example.com" }, { email: "existing@example.com" }],
    };

    const mockResponse: BulkUserCreationResponse = {
      results: [
        {
          success: true,
          user: { id: "1", username: "user1@example.com", tempPassword: "temp123" },
        },
        {
          success: false,
          error: { message: "User already exists", code: 400 },
        },
      ],
      totalRequested: 2,
      totalSuccessful: 1,
      totalFailed: 1,
    };

    vi.mocked(identityManagerApi.usersApi.batchCreateUsers).mockResolvedValue({
      data: mockResponse,
      status: 200,
      statusText: "OK",
      headers: {},
      config: { headers: new AxiosHeaders() },
    });

    const result = await userService.createBatch(bulkRequest);

    expect(identityManagerApi.usersApi.batchCreateUsers).toHaveBeenCalledWith(bulkRequest);
    expect(result).toEqual(mockResponse);
    expect(result.totalSuccessful).toBe(1);
    expect(result.totalFailed).toBe(1);
  });

  it("should handle empty user list", async () => {
    const bulkRequest: BulkUserCreationRequest = {
      users: [],
    };

    const mockResponse: BulkUserCreationResponse = {
      results: [],
      totalRequested: 0,
      totalSuccessful: 0,
      totalFailed: 0,
    };

    vi.mocked(identityManagerApi.usersApi.batchCreateUsers).mockResolvedValue({
      data: mockResponse,
      status: 200,
      statusText: "OK",
      headers: {},
      config: { headers: new AxiosHeaders() },
    });

    const result = await userService.createBatch(bulkRequest);

    expect(identityManagerApi.usersApi.batchCreateUsers).toHaveBeenCalledWith(bulkRequest);
    expect(result).toEqual(mockResponse);
    expect(result.results).toHaveLength(0);
  });

  it("should handle API errors by calling httpService.handleHttpError", async () => {
    const bulkRequest: BulkUserCreationRequest = {
      users: [{ email: "test@example.com" }],
    };

    const mockError = new Error("API Error");
    const handledError = new Error("Handled Error");

    vi.mocked(identityManagerApi.usersApi.batchCreateUsers).mockRejectedValue(mockError);
    vi.mocked(httpService.handleHttpError).mockImplementation(() => {
      throw handledError;
    });

    await expect(userService.createBatch(bulkRequest)).rejects.toThrow("Handled Error");

    expect(identityManagerApi.usersApi.batchCreateUsers).toHaveBeenCalledWith(bulkRequest);
    expect(httpService.handleHttpError).toHaveBeenCalledWith(mockError);
  });

  it("should handle single user creation successfully", async () => {
    const bulkRequest: BulkUserCreationRequest = {
      users: [{ email: "single@example.com" }],
    };

    const mockResponse: BulkUserCreationResponse = {
      results: [
        {
          success: true,
          user: { id: "1", username: "single@example.com", tempPassword: "temp789" },
        },
      ],
      totalRequested: 1,
      totalSuccessful: 1,
      totalFailed: 0,
    };

    vi.mocked(identityManagerApi.usersApi.batchCreateUsers).mockResolvedValue({
      data: mockResponse,
      status: 200,
      statusText: "OK",
      headers: {},
      config: { headers: new AxiosHeaders() },
    });

    const result = await userService.createBatch(bulkRequest);

    expect(result.results).toHaveLength(1);
    expect(result.results[0].success).toBe(true);
    expect(result.results[0].user?.username).toBe("single@example.com");
  });
});
