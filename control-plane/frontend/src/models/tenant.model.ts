export enum EContractTypes {
  Normal = "normal",
  Freemium = "freemium",
}

export interface ITenantClientConfigInfo {
  airgapped: boolean;
  authFlow: string;
  realm: string;
  clientId: string;
  idpIssuerUrl: string;
  redirectUri?: string;
}

export interface ITenantServerConfigInfo {
  clientId: string;
  issuerUrl: string;
  groupsClaim?: string;
}

export interface ITenantConfigInfo {
  clientConfigInfo: ITenantClientConfigInfo;
  serverConfigInfo: ITenantServerConfigInfo;
}

export interface IAuthProviderData {
  ssoEnabled: boolean;
  issuerUrl: string;
}
