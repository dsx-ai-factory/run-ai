import type { Mock } from "vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useAuthStore } from "./auth.store";
import { authService } from "@/services/control-plane/auth.service/auth.service";
import { storageUtil } from "@/utils/storage.util";
import jwt_decode from "jwt-decode";

vi.mock("jwt-decode", () => ({
  default: vi.fn(),
}));

vi.mock("@/services/control-plane/auth.service/auth.service", () => ({
  authService: {
    refreshToken: vi.fn(),
    exchangeCode: vi.fn(),
    changePassword: vi.fn(),
    isValidSsoEmailMapperToken: vi.fn(),
  },
}));

vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn(),
    save: vi.fn(),
    remove: vi.fn(),
  },
}));

vi.mock("@/services/infra/client-apis/client-apis-init", () => ({
  initClientApis: vi.fn(),
}));

vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({ config: { authClientID: "test-client" } }),
}));

vi.mock("@/stores/setting.store", () => ({
  useSettingStore: () => ({}),
}));

const mockSecuritySettingsStoreState = { settings: null as unknown };

vi.mock("@/stores/security-settings.store", () => ({
  useSecuritySettingsStore: vi.fn(() => mockSecuritySettingsStoreState),
}));

vi.mock("@/stores/idps.store", () => ({
  useIdpsStore: () => ({ idps: [] }),
}));

vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({
    hasPermission: vi.fn().mockReturnValue(false),
  }),
}));

vi.mock("@/plugins/amplitude", () => ({
  amplitudeService: {
    init: vi.fn(),
  },
}));

Object.defineProperty(window, "location", {
  value: {
    origin: "https://runai.example.com",
    replace: vi.fn(),
  },
  writable: true,
});

vi.mock("@/common/storage.constant", () => ({
  ACCESS_TOKEN: "runai_access_token",
  ID_TOKEN: "runai_id_token",
  REFRESH_TOKEN: "runai_refresh_token",
  HAD_SESSION: "had-session",
  OAUTH_STATE: "oauth-state",
  PKCE_CODE_VERIFIER: "pkce-code-verifier",
  STORAGE_KEYS: {
    ADMIN_MESSAGE_ID: "admin-message-id",
  },
}));

describe("Auth Store - Token Refresh Optimization", () => {
  let authStore: ReturnType<typeof useAuthStore>;
  let mockRefreshToken: Mock;
  let mockJwtDecode: Mock;

  const mockCurrentToken = "current.jwt.token";

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();

    vi.clearAllMocks();
    vi.restoreAllMocks(); // Restore Date.now and other mocks

    mockRefreshToken = authService.refreshToken as Mock;
    mockJwtDecode = jwt_decode as Mock;

    authStore.accessToken = mockCurrentToken;
    authStore.idToken = "current.id.token";

    authStore.clockOffset = 0;
    authStore.lastTokenIat = 0;
  });

  describe("isTokenAboutToExpire", () => {
    it("should return true for token about to expire (remaining time < 90 seconds)", () => {
      const now = Date.now();
      const iat = Math.floor(now / 1000);
      const exp = iat + 60;

      authStore.clockOffset = 0;
      authStore.lastTokenIat = 0;

      mockJwtDecode.mockReturnValue({ iat, exp });
      vi.spyOn(Math, "random").mockReturnValue(0); // This gives us threshold of 90

      const result = authStore.isTokenAboutToExpire("test.jwt.token");

      expect(result).toBe(true);
      expect(mockJwtDecode).toHaveBeenCalledWith("test.jwt.token");
    });

    it("should return false for token not about to expire (remaining time > 180 seconds)", () => {
      const now = Date.now();
      const iat = Math.floor(now / 1000);
      const exp = iat + 600;

      authStore.clockOffset = 0;
      authStore.lastTokenIat = 0;

      mockJwtDecode.mockReturnValue({ iat, exp });
      vi.spyOn(Math, "random").mockReturnValue(1); // This gives us threshold of 180

      const result = authStore.isTokenAboutToExpire("test.jwt.token");

      expect(result).toBe(false);
      expect(mockJwtDecode).toHaveBeenCalledWith("test.jwt.token");
    });

    it("should return false for token with missing exp field", () => {
      const now = Date.now();
      const iat = Math.floor(now / 1000) - 3600;

      mockJwtDecode.mockReturnValue({ iat }); // Missing exp

      const result = authStore.isTokenAboutToExpire("test.jwt.token");

      expect(result).toBe(false);
    });

    it("should return false for token with missing iat field", () => {
      const now = Date.now();
      const exp = Math.floor(now / 1000) + 3600;

      mockJwtDecode.mockReturnValue({ exp }); // Missing iat

      const result = authStore.isTokenAboutToExpire("test.jwt.token");

      expect(result).toBe(false);
    });

    it("should return true for invalid token that causes jwt_decode to throw", () => {
      mockJwtDecode.mockImplementation(() => {
        throw new Error("Invalid token");
      });

      const result = authStore.isTokenAboutToExpire("invalid.token");

      expect(result).toBe(true);
    });

    it("should use random threshold between 90-180 seconds", () => {
      const now = Date.now();
      const iat = Math.floor(now / 1000);
      const exp = iat + 200;

      authStore.clockOffset = 0;
      authStore.lastTokenIat = 0;

      mockJwtDecode.mockReturnValue({ iat, exp });

      vi.spyOn(Math, "random").mockReturnValue(0);
      let result = authStore.isTokenAboutToExpire("test.jwt.token");
      expect(result).toBe(false);

      authStore.clockOffset = 0;
      authStore.lastTokenIat = 0;

      vi.spyOn(Math, "random").mockReturnValue(1);
      result = authStore.isTokenAboutToExpire("test.jwt.token");
      expect(result).toBe(false);
    });

    it("should handle clock offset making old tokens appear fresh", () => {
      const baseTime = 1700000000;
      const iat = baseTime - 1800;
      const exp = baseTime - 600;

      vi.spyOn(Date, "now").mockReturnValue(baseTime * 1000);

      mockJwtDecode.mockReturnValue({ iat, exp });
      vi.spyOn(Math, "random").mockReturnValue(0.5); // Random threshold = 135

      const result = authStore.isTokenAboutToExpire("old.token");

      expect(result).toBe(false);
    });

    it("should detect and handle clock offset (client ahead of server)", () => {
      const baseTime = 1700000000;
      const iat = baseTime - 300;
      const exp = iat + 900;
      vi.spyOn(Date, "now").mockReturnValue((baseTime + 3600) * 1000);

      authStore.clockOffset = 0;
      authStore.lastTokenIat = 0;

      mockJwtDecode.mockReturnValue({ iat, exp });
      vi.spyOn(Math, "random").mockReturnValue(0.5); // threshold = 135

      const result = authStore.isTokenAboutToExpire("new.token.123");

      expect(result).toBe(false);
    });

    it("should handle subsequent validations with established offset", () => {
      const baseTime = 1700000000;
      const iat = baseTime - 300;
      const exp = iat + 900;
      vi.spyOn(Date, "now").mockReturnValue((baseTime + 3600) * 1000);

      authStore.clockOffset = 3600;
      authStore.lastTokenIat = iat;

      mockJwtDecode.mockReturnValue({ iat, exp });
      vi.spyOn(Math, "random").mockReturnValue(0);

      const result = authStore.isTokenAboutToExpire("same.token.123");

      expect(authStore.clockOffset).toBe(3600);
      expect(result).toBe(false);
    });
  });

  describe("updateAccessToken", () => {
    it("should call refresh API without arguments (refresh token comes from httpOnly cookie)", async () => {
      const refreshedTokens = {
        accessToken: "new.access.token",
        idToken: "new.id.token",
      };

      mockRefreshToken.mockResolvedValue(refreshedTokens);

      const setTokensSpy = vi.spyOn(authStore, "setTokens");
      await authStore.updateAccessToken();

      expect(mockRefreshToken).toHaveBeenCalledWith();
      expect(setTokensSpy).toHaveBeenCalledWith(refreshedTokens);
    });
  });

  describe("integration scenarios", () => {
    it("should always call refresh API", async () => {
      const refreshedTokens = {
        accessToken: "new.access.token",
        idToken: "new.id.token",
        refreshToken: "new.refresh.token",
      };

      mockRefreshToken.mockResolvedValue(refreshedTokens);
      const setTokensSpy = vi.spyOn(authStore, "setTokens");

      await authStore.updateAccessToken();

      expect(mockRefreshToken).toHaveBeenCalled();
      expect(setTokensSpy).toHaveBeenCalledWith(refreshedTokens);
    });
  });
});

describe("Auth Store - loadUserInfo ID token decoding", () => {
  let authStore: ReturnType<typeof useAuthStore>;
  let mockJwtDecode: Mock;

  const accessPayload = {
    email: "user@example.com",
    identity_provider: "local",
    tenant_id: 1,
    subject_uuid: "user-uuid-123",
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();
    vi.clearAllMocks();
    mockJwtDecode = jwt_decode as Mock;
    authStore.accessToken = "mock.access.token";
    vi.spyOn(authStore, "setLocalStorageLoggedInUser").mockImplementation(() => {});
  });

  it("should extract firstName and lastName from ID token claims", async () => {
    authStore.idToken = "valid.id.token";
    mockJwtDecode.mockReturnValueOnce(accessPayload).mockReturnValueOnce({ given_name: "John", family_name: "Doe" });

    await authStore.loadUserInfo();

    expect(authStore.currentUser.firstName).toBe("John");
    expect(authStore.currentUser.lastName).toBe("Doe");
  });

  it("should trim whitespace from name claims", async () => {
    authStore.idToken = "valid.id.token";
    mockJwtDecode
      .mockReturnValueOnce(accessPayload)
      .mockReturnValueOnce({ given_name: "  John  ", family_name: "  Doe  " });

    await authStore.loadUserInfo();

    expect(authStore.currentUser.firstName).toBe("John");
    expect(authStore.currentUser.lastName).toBe("Doe");
  });

  it("should reset names to empty strings when idToken is absent", async () => {
    authStore.idToken = "";
    authStore.currentUser.firstName = "Stale";
    authStore.currentUser.lastName = "Name";
    mockJwtDecode.mockReturnValueOnce(accessPayload);

    await authStore.loadUserInfo();

    expect(authStore.currentUser.firstName).toBe("");
    expect(authStore.currentUser.lastName).toBe("");
  });

  it("should not throw and should clear names when ID token is malformed", async () => {
    authStore.idToken = "malformed.token";
    mockJwtDecode.mockReturnValueOnce(accessPayload).mockImplementationOnce(() => {
      throw new Error("InvalidTokenError");
    });

    await expect(authStore.loadUserInfo()).resolves.not.toThrow();
    expect(authStore.currentUser.firstName).toBe("");
    expect(authStore.currentUser.lastName).toBe("");
  });

  it("should fall back to empty strings when claims are missing from ID token", async () => {
    authStore.idToken = "valid.id.token";
    mockJwtDecode.mockReturnValueOnce(accessPayload).mockReturnValueOnce({});

    await authStore.loadUserInfo();

    expect(authStore.currentUser.firstName).toBe("");
    expect(authStore.currentUser.lastName).toBe("");
  });
});

describe("Auth Store - userFullName and userInitials getters", () => {
  let authStore: ReturnType<typeof useAuthStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();
    vi.clearAllMocks();
  });

  describe("userFullName", () => {
    it("should return 'First Last' when both names are set", () => {
      authStore.currentUser.firstName = "John";
      authStore.currentUser.lastName = "Doe";
      expect(authStore.userFullName).toBe("John Doe");
    });

    it("should return only firstName when lastName is empty", () => {
      authStore.currentUser.firstName = "John";
      authStore.currentUser.lastName = "";
      expect(authStore.userFullName).toBe("John");
    });

    it("should return empty string when both names are empty", () => {
      authStore.currentUser.firstName = "";
      authStore.currentUser.lastName = "";
      expect(authStore.userFullName).toBe("");
    });
  });

  describe("userInitials", () => {
    it("should return uppercase initials from firstName and lastName", () => {
      authStore.currentUser.firstName = "John";
      authStore.currentUser.lastName = "Doe";
      authStore.currentUser.username = "john@example.com";
      expect(authStore.userInitials).toBe("JD");
    });

    it("should return only first initial when lastName is empty", () => {
      authStore.currentUser.firstName = "John";
      authStore.currentUser.lastName = "";
      authStore.currentUser.username = "john@example.com";
      expect(authStore.userInitials).toBe("J");
    });

    it("should fall back to first letter of username when names are empty", () => {
      authStore.currentUser.firstName = "";
      authStore.currentUser.lastName = "";
      authStore.currentUser.username = "john@example.com";
      expect(authStore.userInitials).toBe("J");
    });

    it("should return empty string when names and username are all empty", () => {
      authStore.currentUser.firstName = "";
      authStore.currentUser.lastName = "";
      authStore.currentUser.username = "";
      expect(authStore.userInitials).toBe("");
    });
  });
});

describe("Auth Store - Logout Functionality", () => {
  let authStore: ReturnType<typeof useAuthStore>;
  let mockStorageRemove: Mock;
  let mockWindowLocationReplace: Mock;

  beforeEach(async () => {
    setActivePinia(createPinia());
    authStore = useAuthStore();

    vi.clearAllMocks();

    const { storageUtil } = await import("@/utils/storage.util");
    mockStorageRemove = storageUtil.remove as Mock;

    mockWindowLocationReplace = window.location.replace as Mock;

    authStore.idToken = "test.id.token";
    authStore.accessToken = "test.access.token";

    authStore.authProviderConfig = {
      end_session_endpoint: "https://auth.example.com/logout",
    };

    mockSecuritySettingsStoreState.settings = null;
  });

  describe("logout with cached redirect URI", () => {
    it("should use cached logout redirect URI when available", async () => {
      const customRedirectUri = "https://custom.redirect.com";
      mockSecuritySettingsStoreState.settings = {
        logoutRedirectUri: {
          value: { uri: customRedirectUri },
        },
      };

      await authStore.logout();

      expect(mockWindowLocationReplace).toHaveBeenCalledWith(
        expect.stringContaining(`post_logout_redirect_uri=${encodeURIComponent(customRedirectUri)}`),
      );
      expect(mockWindowLocationReplace).toHaveBeenCalledWith(expect.stringContaining("client_id=test-client"));
      expect(mockWindowLocationReplace).toHaveBeenCalledWith(expect.stringContaining("id_token_hint=test.id.token"));
    });

    it("should use window.location.origin as fallback when no cached redirect URI", async () => {
      mockSecuritySettingsStoreState.settings = null;

      await authStore.logout();

      expect(mockWindowLocationReplace).toHaveBeenCalledWith(
        expect.stringContaining(`post_logout_redirect_uri=${encodeURIComponent(window.location.origin)}`),
      );
    });

    it("should use window.location.origin when cached settings exist but no logoutRedirectUri", async () => {
      mockSecuritySettingsStoreState.settings = {
        otherSettings: "some value",
      };

      await authStore.logout();

      expect(mockWindowLocationReplace).toHaveBeenCalledWith(
        expect.stringContaining(`post_logout_redirect_uri=${encodeURIComponent(window.location.origin)}`),
      );
    });

    it("should use window.location.origin when cached logoutRedirectUri is empty", async () => {
      mockSecuritySettingsStoreState.settings = {
        logoutRedirectUri: {
          value: { uri: "" },
        },
      };

      await authStore.logout();

      expect(mockWindowLocationReplace).toHaveBeenCalledWith(
        expect.stringContaining(`post_logout_redirect_uri=${encodeURIComponent(window.location.origin)}`),
      );
    });
  });

  describe("logout cleanup and flow", () => {
    it("should clear tokens and storage correctly", async () => {
      mockSecuritySettingsStoreState.settings = null;

      await authStore.logout();

      expect(authStore.idToken).toBe("");
      expect(authStore.accessToken).toBe("");

      expect(mockStorageRemove).toHaveBeenCalledWith("admin-message-id");
    });

    it("should construct proper logout URL with all parameters", async () => {
      const customRedirectUri = "https://custom.redirect.com";
      mockSecuritySettingsStoreState.settings = {
        logoutRedirectUri: { value: { uri: customRedirectUri } },
      };

      await authStore.logout();

      const expectedUrl = new URL("https://auth.example.com/logout");
      expectedUrl.searchParams.set("client_id", "test-client");
      expectedUrl.searchParams.set("post_logout_redirect_uri", customRedirectUri);
      expectedUrl.searchParams.set("id_token_hint", "test.id.token");

      expect(mockWindowLocationReplace).toHaveBeenCalledWith(expectedUrl.toString());
    });

    it("should handle logout without id_token_hint when idToken is empty", async () => {
      authStore.idToken = "";
      mockSecuritySettingsStoreState.settings = null;

      await authStore.logout();

      const calledUrl = mockWindowLocationReplace.mock.calls[0][0];
      expect(calledUrl).not.toContain("id_token_hint");
      expect(calledUrl).toContain("client_id=test-client");
      expect(calledUrl).toContain(`post_logout_redirect_uri=${encodeURIComponent(window.location.origin)}`);
    });

    it("should handle broadcast channel logout message", async () => {
      const mockLogoutChannel = {
        postMessage: vi.fn(),
      } as Partial<BroadcastChannel> as BroadcastChannel;
      authStore.logoutChannel = mockLogoutChannel;
      mockSecuritySettingsStoreState.settings = null;

      await authStore.logout();

      expect(mockLogoutChannel.postMessage).toHaveBeenCalledWith("logout");
    });

    it("should handle broadcast channel errors gracefully", async () => {
      authStore.logoutChannel = {
        postMessage: vi.fn().mockImplementation(() => {
          throw new Error("Broadcast failed");
        }),
      } as Partial<BroadcastChannel> as BroadcastChannel;
      mockSecuritySettingsStoreState.settings = null;

      const consoleSpy = vi.spyOn(console, "warn").mockImplementation(() => {});

      await authStore.logout();

      expect(consoleSpy).toHaveBeenCalledWith("Failed to broadcast logout:", expect.any(Error));
      expect(mockWindowLocationReplace).toHaveBeenCalled(); // Logout should still continue

      consoleSpy.mockRestore();
    });
  });

  describe("logout performance and reliability", () => {
    it("should complete logout without errors", async () => {
      const customRedirectUri = "https://custom.redirect.com";
      mockSecuritySettingsStoreState.settings = {
        logoutRedirectUri: { value: { uri: customRedirectUri } },
      };

      await authStore.logout();

      expect(authStore.idToken).toBe("");
      expect(authStore.accessToken).toBe("");
      expect(mockWindowLocationReplace).toHaveBeenCalled();
    });

    it("should be synchronous and fast", async () => {
      mockSecuritySettingsStoreState.settings = {
        logoutRedirectUri: { value: { uri: "https://custom.redirect.com" } },
      };

      const startTime = performance.now();

      await authStore.logout();

      const duration = performance.now() - startTime;
      expect(duration).toBeLessThan(10);
    });
  });
});

describe("Auth Store - Token Storage (httpOnly cookie flow)", () => {
  let authStore: ReturnType<typeof useAuthStore>;
  // Backend returns refreshToken via httpOnly cookie now, so the body field is
  // unused by the store. Keep an empty string to satisfy IRefreshTokenResponse.
  const mockTokens = {
    accessToken: "mock-access-token",
    idToken: "mock-id-token",
    refreshToken: "",
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();
    vi.clearAllMocks();
  });

  it("should not write tokens to localStorage on setTokens (refresh token lives in cookie)", () => {
    authStore.setTokens(mockTokens);

    expect(storageUtil.save).not.toHaveBeenCalledWith("runai_id_token", expect.anything());
    expect(storageUtil.save).not.toHaveBeenCalledWith("runai_access_token", expect.anything());
    expect(storageUtil.save).not.toHaveBeenCalledWith("runai_refresh_token", expect.anything());
  });

  it("should mark had-session on setTokens so the app knows to attempt a refresh on reload", () => {
    authStore.setTokens(mockTokens);

    expect(storageUtil.save).toHaveBeenCalledWith("had-session", "true");
  });

  it("should clear legacy localStorage token entries and had-session on clearTokens", () => {
    authStore.clearTokens();

    expect(storageUtil.remove).toHaveBeenCalledWith("runai_id_token");
    expect(storageUtil.remove).toHaveBeenCalledWith("runai_access_token");
    expect(storageUtil.remove).toHaveBeenCalledWith("runai_refresh_token");
    expect(storageUtil.remove).toHaveBeenCalledWith("had-session");
  });
});

describe("Auth Store - postLogin (fragment mode)", () => {
  let authStore: ReturnType<typeof useAuthStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();
    vi.clearAllMocks();

    Object.assign(window.location, {
      hash: "#code=test-auth-code&state=test-state",
      pathname: "/login/callback",
    });

    // history.replaceState can throw a SecurityError in jsdom when called with a
    // cross-origin URL (our mock uses https://runai.example.com while jsdom sits
    // at http://localhost). Suppress it — we don't assert on replaceState here.
    vi.spyOn(history, "replaceState").mockImplementation(() => {});

    sessionStorage.setItem("oauth-state", "test-state");
    sessionStorage.setItem("pkce-code-verifier", "test-verifier");
  });

  afterEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
  });

  it("reads code and state from the URL hash and exchanges them for tokens", async () => {
    (authService.exchangeCode as Mock).mockResolvedValue({ accessToken: "a", idToken: "i", refreshToken: "" });

    await authStore.postLogin();

    expect(authService.exchangeCode).toHaveBeenCalledWith(
      "test-auth-code",
      "https://runai.example.com/login/callback",
      "test-verifier",
    );
  });

  it("throws on state mismatch", async () => {
    sessionStorage.setItem("oauth-state", "wrong-state");

    await expect(authStore.postLogin()).rejects.toThrow("OAuth state mismatch");
    expect(authService.exchangeCode).not.toHaveBeenCalled();
  });

  it("throws when PKCE verifier is missing", async () => {
    sessionStorage.removeItem("pkce-code-verifier");

    await expect(authStore.postLogin()).rejects.toThrow("PKCE code verifier is missing");
    expect(authService.exchangeCode).not.toHaveBeenCalled();
  });

  it("returns early without calling exchangeCode when no code is in the hash", async () => {
    Object.assign(window.location, { hash: "" });

    await authStore.postLogin();

    expect(authService.exchangeCode).not.toHaveBeenCalled();
  });
});
