import { sha256 } from "js-sha256";

function base64URLEncode(bytes: Uint8Array): string {
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

function generateOAuthState(): string {
  return crypto.randomUUID();
}

function generatePKCE(): { codeVerifier: string; codeChallenge: string } {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  const codeVerifier = base64URLEncode(array);
  const codeChallenge = base64URLEncode(new Uint8Array(sha256.arrayBuffer(codeVerifier)));
  return { codeVerifier, codeChallenge };
}

function buildAuthorizationLoginUrl(
  authorizationEndpoint: string,
  authClientID: string,
  connection: string,
  redirect_uri: string,
  state: string,
  codeChallenge: string,
): string {
  if (!authorizationEndpoint) {
    throw new Error("Authorization endpoint is missing — cannot redirect to login");
  }

  const searchParams = new URLSearchParams();
  searchParams.set("response_type", "code");
  searchParams.set("connection", connection);
  searchParams.set("client_id", authClientID);
  searchParams.set("redirect_uri", `${redirect_uri}/login/callback`);
  searchParams.set("scope", "openid email profile");
  searchParams.set("state", state);
  searchParams.set("code_challenge", codeChallenge);
  searchParams.set("code_challenge_method", "S256");
  // Deliver the authorization code in the URL fragment (#code=…) rather than
  // the query string (?code=…).  The hash is never sent to servers or included
  // in Referer headers, so the code cannot leak via same-origin subrequests.
  searchParams.set("response_mode", "fragment");

  return `${authorizationEndpoint}?${searchParams.toString()}`;
}

export const pkceUtil = {
  base64URLEncode,
  generateOAuthState,
  generatePKCE,
  buildAuthorizationLoginUrl,
};
