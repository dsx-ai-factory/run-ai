export interface IJwtPayload {
  iss?: string;
  sub?: string;
  aud?: string[] | string;
  exp?: number;
  nbf?: number;
  iat?: number;
  jti?: string;
  sid?: string; // Session ID
  subject_uuid?: string; // User ID
}

export interface IRefreshTokenResponse {
  idToken: string;
  refreshToken: string;
  accessToken: string;
}
