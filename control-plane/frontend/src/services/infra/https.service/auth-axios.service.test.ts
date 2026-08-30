import { describe, it, expect, vi, beforeEach } from "vitest";
import { FAKE_ACCESS_TOKEN_C } from "@/tests/fixtures/auth.fixtures";
import type { AxiosRequestConfig } from "axios";
import { authenticatedAxios } from "./auth-axios.service";

// Type for accessing internal axios interceptor handlers
type InterceptorHandler = {
  fulfilled?: (config: AxiosRequestConfig) => Promise<AxiosRequestConfig>;
  rejected?: (error: unknown) => Promise<unknown>;
};

// Helper to get the request interceptor handler
const getRequestInterceptor = () =>
  (authenticatedAxios.interceptors.request as unknown as { handlers: InterceptorHandler[] }).handlers[0]?.fulfilled;

// Helper to get the response interceptor error handler
const getResponseInterceptorError = () =>
  (authenticatedAxios.interceptors.response as unknown as { handlers: InterceptorHandler[] }).handlers[0]?.rejected;

// Mock auth store with realistic behavior
const mockAuthStore = {
  accessToken: FAKE_ACCESS_TOKEN_C,
  isTokenAboutToExpire: vi.fn(),
  updateAccessToken: vi.fn(),
  logout: vi.fn(),
  clearTokens: vi.fn(),
};

vi.mock("@/stores/auth.store", () => ({
  useAuthStore: vi.fn(() => mockAuthStore),
}));

// Mock Web Locks API
const mockNavigatorLocks = {
  request: vi.fn(),
};

Object.defineProperty(navigator, "locks", {
  value: mockNavigatorLocks,
  configurable: true,
});

describe("authenticatedAxios", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Reset auth store defaults
    mockAuthStore.accessToken = "valid-token";
    mockAuthStore.isTokenAboutToExpire.mockReturnValue(false);
    mockAuthStore.updateAccessToken.mockResolvedValue(undefined);
    mockAuthStore.logout.mockResolvedValue(undefined);

    // Default Web Locks behavior - immediately execute callback
    mockNavigatorLocks.request.mockImplementation(async (lockName, callback) => {
      return await callback();
    });
  });

  it("should create a working axios instance", () => {
    expect(authenticatedAxios).toBeDefined();
    expect(authenticatedAxios.get).toBeInstanceOf(Function);
    expect(authenticatedAxios.post).toBeInstanceOf(Function);
    expect(authenticatedAxios.interceptors).toBeDefined();
  });

  describe("Request Interceptor", () => {
    it("should skip auth endpoints without token validation", async () => {
      const authEndpoints = ["/api/v1/token", "/auth/login", "/login"];

      for (const url of authEndpoints) {
        const config = { url, headers: {} };
        const interceptor = getRequestInterceptor();

        const result = await interceptor?.(config);

        expect(result?.url).toBe(url);
        expect(mockAuthStore.isTokenAboutToExpire).not.toHaveBeenCalled();
        expect(mockNavigatorLocks.request).not.toHaveBeenCalled();
      }
    });

    it("should skip validation when no access token", async () => {
      mockAuthStore.accessToken = "";

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      const result = await interceptor?.(config);

      expect(result).toEqual(config);
      expect(mockAuthStore.isTokenAboutToExpire).not.toHaveBeenCalled();
    });

    it("should validate token for regular endpoints", async () => {
      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      await interceptor?.(config);

      expect(mockNavigatorLocks.request).toHaveBeenCalledWith("runai-token-refresh", expect.any(Function));
      expect(mockAuthStore.isTokenAboutToExpire).toHaveBeenCalledWith("valid-token");
    });

    it("should add Authorization header when token is valid", async () => {
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(false);

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      const result = await interceptor?.(config);

      expect(result?.headers?.["Authorization"]).toBe("Bearer valid-token");
    });

    it("should refresh token when expired", async () => {
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(true);
      mockAuthStore.accessToken = "new-fresh-token";

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      const result = await interceptor?.(config);

      expect(mockAuthStore.updateAccessToken).toHaveBeenCalled();
      expect(result?.headers?.["Authorization"]).toBe("Bearer new-fresh-token");
    });
  });

  describe("Web Locks Coordination", () => {
    it("should use Web Locks API for coordination", async () => {
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(true);

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      await interceptor?.(config);

      expect(mockNavigatorLocks.request).toHaveBeenCalledWith("runai-token-refresh", expect.any(Function));
    });

    it("should coordinate multiple simultaneous requests", async () => {
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(true);
      let lockCallbackCount = 0;

      // Simulate Web Locks coordination - only first call refreshes
      mockNavigatorLocks.request.mockImplementation(async (lockName, callback) => {
        lockCallbackCount++;
        if (lockCallbackCount === 1) {
          // First lock holder - token is expired, refresh it
          return await callback();
        } else {
          // Other lock holders - token already refreshed, skip
          mockAuthStore.isTokenAboutToExpire.mockReturnValue(false);
          return await callback();
        }
      });

      const config1 = { url: "/api/data1", headers: {} };
      const config2 = { url: "/api/data2", headers: {} };
      const interceptor = getRequestInterceptor();

      await Promise.all([interceptor?.(config1), interceptor?.(config2)]);

      expect(mockNavigatorLocks.request).toHaveBeenCalledTimes(2);
      expect(mockAuthStore.updateAccessToken).toHaveBeenCalledTimes(1); // Only called once
    });
  });

  describe("Error Handling", () => {
    it("should logout on authentication errors", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(true);
      const authError = {
        response: { status: 401 },
        isAxiosError: true,
      };
      mockAuthStore.updateAccessToken.mockRejectedValue(authError);

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      await interceptor?.(config);

      expect(mockAuthStore.logout).toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("should not logout on network errors", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(true);
      const networkError = new Error("Network timeout");
      mockAuthStore.updateAccessToken.mockRejectedValue(networkError);

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      await expect(interceptor?.(config)).rejects.toThrow("Authentication failed");
      expect(mockAuthStore.logout).not.toHaveBeenCalled();
      errorSpy.mockRestore();
    });

    it("should handle Web Locks coordination errors", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      mockNavigatorLocks.request.mockRejectedValue(new Error("Lock failed"));

      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      await expect(interceptor?.(config)).rejects.toThrow("Authentication failed");
      errorSpy.mockRestore();
    });
  });

  describe("Token Validation", () => {
    it("should check token validity for each lock holder", async () => {
      let callCount = 0;
      mockNavigatorLocks.request.mockImplementation(async (lockName, callback) => {
        callCount++;
        // Simulate different token states for each call
        if (callCount === 1) {
          mockAuthStore.isTokenAboutToExpire.mockReturnValue(true); // Needs refresh
        } else {
          mockAuthStore.isTokenAboutToExpire.mockReturnValue(false); // Already refreshed
        }
        return await callback();
      });

      const config1 = { url: "/api/data1", headers: {} };
      const config2 = { url: "/api/data2", headers: {} };
      const interceptor = getRequestInterceptor();

      await Promise.all([interceptor?.(config1), interceptor?.(config2)]);

      expect(mockAuthStore.isTokenAboutToExpire).toHaveBeenCalledTimes(2);
      expect(mockAuthStore.updateAccessToken).toHaveBeenCalledTimes(1); // Only first call refreshes
    });

    it("should always evaluate token freshness (no cache)", async () => {
      const config = { url: "/api/data", headers: {} };
      const interceptor = getRequestInterceptor();

      // Make multiple calls
      await interceptor?.(config);
      await interceptor?.(config);
      await interceptor?.(config);

      // Should check token every time (no cache)
      expect(mockAuthStore.isTokenAboutToExpire).toHaveBeenCalledTimes(3);
    });
  });

  describe("Real Request Flow", () => {
    it("should handle actual HTTP request with token refresh", async () => {
      // Mock a server that returns data
      const mockServer = vi.fn().mockResolvedValue({ data: "success" });

      // Override axios for this test
      const testAxios = authenticatedAxios;
      testAxios.defaults.adapter = async (config) => {
        mockServer(config);
        return {
          data: "success",
          status: 200,
          statusText: "OK",
          headers: {},
          config,
        };
      };

      // Token about to expire
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(true);
      mockAuthStore.accessToken = "fresh-token-after-refresh";

      const response = await testAxios.get("/api/data");

      expect(response.data).toBe("success");
      expect(mockAuthStore.updateAccessToken).toHaveBeenCalled();
      expect(mockServer).toHaveBeenCalledWith(
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: "Bearer fresh-token-after-refresh",
          }),
        }),
      );
    });
  });

  describe("Response Interceptor - Logout 401 Handling", () => {
    it("should not attempt token refresh when logout returns 401", async () => {
      const logoutError = {
        isAxiosError: true,
        response: { status: 401 },
        config: { url: "/api/v1/logout", headers: {} },
      };

      const responseInterceptorError = getResponseInterceptorError();
      await expect(responseInterceptorError?.(logoutError)).rejects.toBeDefined();

      expect(mockAuthStore.updateAccessToken).not.toHaveBeenCalled();
      expect(mockAuthStore.logout).not.toHaveBeenCalled();
    });

    it("should send Authorization header on logout requests when token exists", async () => {
      mockAuthStore.accessToken = "valid-token";
      mockAuthStore.isTokenAboutToExpire.mockReturnValue(false);

      const config = { url: "/api/v1/logout", headers: {} };
      const interceptor = getRequestInterceptor();

      const result = await interceptor?.(config);

      expect(result?.headers?.["Authorization"]).toBe("Bearer valid-token");
    });
  });
});
