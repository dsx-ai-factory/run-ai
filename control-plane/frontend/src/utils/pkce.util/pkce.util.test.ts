import { describe, it, expect } from "vitest";
import { pkceUtil } from "./pkce.util";

describe("pkceUtil.buildAuthorizationLoginUrl", () => {
  const BASE_ARGS = {
    authorizationEndpoint: "https://auth.example.com/realms/runai/protocol/openid-connect/auth",
    authClientID: "runai-admin-ui",
    connection: "runai",
    redirect_uri: "https://app.example.com",
    state: "test-state",
    codeChallenge: "test-challenge",
  };

  function build(overrides: Partial<typeof BASE_ARGS> = {}): URL {
    const args = { ...BASE_ARGS, ...overrides };
    return new URL(
      pkceUtil.buildAuthorizationLoginUrl(
        args.authorizationEndpoint,
        args.authClientID,
        args.connection,
        args.redirect_uri,
        args.state,
        args.codeChallenge,
      ),
    );
  }

  it("sets response_mode=fragment so the code is delivered in the URL hash, not the query string", () => {
    const url = build();
    expect(url.searchParams.get("response_mode")).toBe("fragment");
  });

  it("sets response_type=code", () => {
    expect(build().searchParams.get("response_type")).toBe("code");
  });

  it("appends /login/callback to the redirect_uri", () => {
    expect(build().searchParams.get("redirect_uri")).toBe("https://app.example.com/login/callback");
  });

  it("throws when authorizationEndpoint is empty", () => {
    expect(() => build({ authorizationEndpoint: "" })).toThrow("Authorization endpoint is missing");
  });
});
