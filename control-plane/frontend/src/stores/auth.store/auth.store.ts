import jwt_decode from "jwt-decode";
import { defineStore } from "pinia";
import { authService } from "@/services/control-plane/auth.service/auth.service";
import { tenantService } from "@/services/control-plane/tenant.service/tenant.service";
import { brandingSettingsService } from "@/services/control-plane/branding-settings.service/branding-settings.service";
import { useAppStore } from "@/stores/app.store";
import { useSettingStore } from "@/stores/setting.store";
import { useSecuritySettingsStore } from "@/stores/security-settings.store";
import type { IAuthProviderConfig, IMe } from "@/models/user.model";
import type { ITenantConfigInfo } from "@/models/tenant.model";
import type { OrgUnit, ScopeType } from "@/swagger-models/org-unit-service-client";
import { tenantUtil } from "@/utils/tenant.util";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { STORAGE_KEYS } from "@/models/storage.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { storageUtil } from "@/utils/storage.util";
import type { IJwtPayload, IRefreshTokenResponse } from "@/models/auth.model";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { usePermissionStore } from "@/stores/permissions.store";
import type { Tenant } from "@/swagger-models/tenants-manager-client";
import { pkceUtil } from "@/utils/pkce.util/pkce.util";

import {
  ACCESS_TOKEN,
  ID_TOKEN,
  REFRESH_TOKEN,
  HAD_SESSION,
  OAUTH_STATE,
  PKCE_CODE_VERIFIER,
} from "@/common/storage.constant";

import axios from "axios";
import { initClientApis } from "@/services/infra/client-apis/client-apis-init";
import { useIdpsStore } from "@/stores/idps.store";
import { cacheUtils } from "@/services/infra/cache";
const AUTH_PROVIDER_DISCOVERY_PATH = `.well-known/openid-configuration`;

// @ts-ignore
window.jwt_decode = jwt_decode;

export const useAuthStore = defineStore("Auth", {
  state: () => ({
    idToken: "" as string,
    accessToken: "" as string,
    currentUser: {} as IMe,
    tenant: {} as Tenant,
    logo: "" as string,
    lastLogin: "" as string,
    orgUnits: [] as Array<OrgUnit>,
    authProviderConfig: {} as Record<string, string>,
    idpStore: useIdpsStore(),
    appStore: useAppStore(),
    settingStore: useSettingStore(),
    securitySettingsStore: useSecuritySettingsStore(),
    logoutChannel: null as BroadcastChannel | null,
    isLoggingOut: false,
    clockOffset: 0 as number,
    lastTokenIat: 0 as number,
  }),
  getters: {
    orgUnitList(): Array<OrgUnit> {
      return this.orgUnits;
    },
    isAdmin(): boolean {
      const permissionStore = usePermissionStore();
      return permissionStore.hasPermission(ResourceType.Settings, Action.Update);
    },
    userLastLogin(): string {
      return this.lastLogin;
    },
    getUID(): number | null {
      if (!this.accessToken) return null;
      const { uid }: { uid: number } = jwt_decode(this.accessToken);
      if (Number.isInteger(uid)) return uid;
      return uid ? Number(uid) : null;
    },
    getGID(): number | null {
      if (!this.accessToken) return null;
      const { gid }: { gid: number } = jwt_decode(this.accessToken);
      if (Number.isInteger(gid)) return gid;
      return gid ? Number(gid) : null;
    },
    getSupplementaryGroups(): string | null {
      if (!this.accessToken) return null;
      const { supplementarygroups }: { supplementarygroups: string } = jwt_decode(this.accessToken);
      return supplementarygroups || null;
    },
    isSSO(): boolean {
      return this.idpStore.idps.length > 0;
    },
    isCurrentUserSSO(): boolean {
      return !!this.currentUser.identityProvider;
    },
    userEmail(): string {
      return this.currentUser.username;
    },
    userFullName(): string {
      const { firstName, lastName } = this.currentUser;
      return [firstName, lastName].filter(Boolean).join(" ");
    },
    userInitials(): string {
      const { firstName, lastName, username } = this.currentUser;
      return `${firstName?.[0] ?? ""}${lastName?.[0] ?? ""}`.toUpperCase() || username?.[0]?.toUpperCase() || "";
    },
    userIsRunAiEmp(): boolean {
      const email = this.currentUser.username;
      return email?.split("@")[1] === "run.ai";
    },
  },
  actions: {
    clearTokens(): void {
      this.idToken = "";
      this.accessToken = "";

      storageUtil.remove(ID_TOKEN);
      storageUtil.remove(ACCESS_TOKEN);
      storageUtil.remove(REFRESH_TOKEN);
      storageUtil.remove(HAD_SESSION);
    },
    initCrossTabLogoutDetection(): void {
      try {
        if (this.logoutChannel) {
          this.logoutChannel.close();
        }

        this.logoutChannel = new BroadcastChannel("runai-logout");
        this.logoutChannel.onmessage = (event) => {
          if (event.data === "logout" && this.accessToken && !this.isLoggingOut) {
            this.logout().catch(console.error);
          }
        };
      } catch (error) {
        console.warn("BroadcastChannel failed:", error);
      }
    },

    isTokenAboutToExpire(token: string): boolean {
      try {
        const tokenPayload: IJwtPayload = jwt_decode(token);
        if (tokenPayload.exp && tokenPayload.iat) {
          const tokenLifetime = tokenPayload.exp - tokenPayload.iat;
          const localTime = Date.now() / 1000;

          if (tokenPayload.iat !== this.lastTokenIat) {
            this.clockOffset = localTime - tokenPayload.iat;
            this.lastTokenIat = tokenPayload.iat;
          }

          const adjustedTime = localTime - this.clockOffset;
          const tokenAge = adjustedTime - tokenPayload.iat;
          const remainingTime = tokenLifetime - tokenAge;

          const randomThreshold = 90 + Math.random() * 90;

          return remainingTime < randomThreshold;
        }
        return false;
      } catch (error) {
        return true;
      }
    },
    async loadUserOrgUnits(): Promise<Array<OrgUnit>> {
      this.orgUnits = await orgTreeService.getOrgTreeUnits();
      return this.orgUnits;
    },
    getOrgUnitByScopeAndId(scope: ScopeType, scopeId: string): OrgUnit | undefined {
      return this.orgUnits.find((ou: OrgUnit) => ou.type === scope && ou.id === scopeId);
    },
    async loadUserInfo(): Promise<void> {
      const { email, identity_provider, tenant_id, subject_uuid }: IAuthProviderConfig = jwt_decode(this.accessToken);
      this.currentUser.username = email;
      this.currentUser.id = subject_uuid;
      this.currentUser.tenantId = tenant_id;
      this.currentUser.identityProvider = identity_provider;
      this.currentUser.firstName = "";
      this.currentUser.lastName = "";
      if (this.idToken) {
        try {
          const { given_name, family_name }: { given_name?: string; family_name?: string } = jwt_decode(this.idToken);
          this.currentUser.firstName = given_name?.trim() || "";
          this.currentUser.lastName = family_name?.trim() || "";
        } catch {
          // malformed ID token — names stay empty
        }
      }
      this.setLocalStorageLoggedInUser();
    },
    setLocalStorageLoggedInUser(): void {
      const loggedInUserEmail: string | null = localStorage.getItem(STORAGE_KEYS.LOGGED_IN_USER_EMAIL);
      if (loggedInUserEmail !== this.currentUser.username) {
        localStorage.setItem(STORAGE_KEYS.IS_FIRST_LOGIN, "true");
        localStorage.setItem(STORAGE_KEYS.LOGGED_IN_USER_EMAIL, this.currentUser.username);
      } else {
        localStorage.removeItem(STORAGE_KEYS.IS_FIRST_LOGIN);
      }
    },
    async loadTenant(): Promise<Tenant> {
      const tenants = await tenantService.get();
      this.tenant = tenants[0];
      try {
        const res = await brandingSettingsService.getLogo();
        this.logo = res?.logo ?? "";
      } catch (error: unknown) {
        console.info("Couldn't get tenant logo...", error);
      }
      return this.tenant;
    },
    async loadTenantConfigInfo(): Promise<ITenantConfigInfo> {
      const tenantName = this.appStore.config.tenantName;
      const issuerUrl = this.getIssuerUrl();
      return {
        clientConfigInfo: tenantUtil.getClientConfigInfo(issuerUrl, tenantName, this.isSSO),
        serverConfigInfo: tenantUtil.getServerConfigInfo(this.appStore.config.authHost, this.isSSO),
      };
    },
    getIssuerUrl(): string {
      const tenantHost = window.location.host;
      const tenantName = this.appStore.config.tenantName;
      return `https://${tenantHost}/auth/realms/${tenantName}`;
    },
    login(): void {
      const state = pkceUtil.generateOAuthState();
      sessionStorage.setItem(OAUTH_STATE, state);

      const { codeVerifier, codeChallenge } = pkceUtil.generatePKCE();
      sessionStorage.setItem(PKCE_CODE_VERIFIER, codeVerifier);

      const loginUrl: string = pkceUtil.buildAuthorizationLoginUrl(
        this.authProviderConfig["authorization_endpoint"],
        this.appStore.config.authClientID,
        this.appStore.config.tenantName,
        import.meta.env.DEV ? window.location.origin : this.appStore.config.redirectUrl || window.location.origin,
        state,
        codeChallenge,
      );

      window.location.replace(loginUrl);
    },
    async logout(): Promise<void> {
      if (this.isLoggingOut) return;
      this.isLoggingOut = true;

      storageUtil.remove(STORAGE_KEYS.ADMIN_MESSAGE_ID);
      const securitySettingsStore = useSecuritySettingsStore();
      const searchParams = new URLSearchParams();
      if (this.appStore?.config?.authClientID) {
        searchParams.set("client_id", this.appStore.config.authClientID);
      }

      let redirectUri = window.location.origin;

      const cachedLogoutSettings = securitySettingsStore.settings?.logoutRedirectUri;
      if (cachedLogoutSettings?.value?.uri) {
        redirectUri = cachedLogoutSettings.value.uri;
      }

      searchParams.set("post_logout_redirect_uri", redirectUri);
      if (this.idToken) {
        searchParams.set("id_token_hint", this.idToken);
      }

      this.clearTokens();
      storageUtil.remove(HAD_SESSION);
      cacheUtils.clearAll();
      try {
        if (this.logoutChannel) {
          this.logoutChannel.postMessage("logout");
        }
      } catch (error) {
        console.warn("Failed to broadcast logout:", error);
      }

      const endSessionEndpoint = this.authProviderConfig["end_session_endpoint"];
      if (endSessionEndpoint) {
        const signOutUrl = `${endSessionEndpoint}?${searchParams.toString()}`;
        window.location.replace(signOutUrl);
      } else {
        console.error("Logout: end_session_endpoint is missing - redirecting to origin (local logout completed)");
        window.location.replace(redirectUri);
      }
    },
    async loadAuthProviderConfig(authURL: string): Promise<void> {
      const url = `${authURL}/${AUTH_PROVIDER_DISCOVERY_PATH}`;
      try {
        const response = await axios.get(url);
        this.authProviderConfig = response.data;
      } catch (error: unknown) {
        console.error("Failed to load auth provider data: ", error);
        console.log("re-trying to load auth provider data...");
        const response = await axios.get(url);
        this.authProviderConfig = response.data;
      }
    },
    async postLogin(): Promise<void> {
      // The authorization server delivers code/state in the URL fragment
      // (#code=…&state=…) so they are never sent to servers or Referer headers.
      const hashParams = new URLSearchParams(window.location.hash.slice(1));
      const code: string | null = hashParams.get("code");
      const redirectUri = window.location.origin + window.location.pathname;
      if (!code) return;

      const returnedState = hashParams.get("state");
      const expectedState = sessionStorage.getItem(OAUTH_STATE);
      sessionStorage.removeItem(OAUTH_STATE);
      if (!returnedState || returnedState !== expectedState) {
        throw new Error("OAuth state mismatch — login session is invalid");
      }

      const codeVerifier = sessionStorage.getItem(PKCE_CODE_VERIFIER);
      if (!codeVerifier) {
        throw new Error("PKCE code verifier is missing — login session is invalid");
      }

      const tokens: IRefreshTokenResponse = await authService.exchangeCode(code, redirectUri, codeVerifier);
      sessionStorage.removeItem(PKCE_CODE_VERIFIER);
      this.setTokens(tokens);
    },
    setTokens(tokens: IRefreshTokenResponse, shouldInitClients = true): void {
      this.idToken = tokens.idToken || "";
      this.accessToken = tokens.accessToken || "";
      storageUtil.save(HAD_SESSION, "true");

      shouldInitClients && initClientApis(this.accessToken);
    },
    async updateAccessToken(): Promise<void> {
      const tokens: IRefreshTokenResponse = await authService.refreshToken();
      this.setTokens(tokens);
    },
    updateUserLastLoggedIn(): void {
      this.lastLogin = dateUtil.dateFormat(new Date(), DATE_FORMATS.STANDARD_DATE_TIME);
    },
  },
});
