---
name: configure-local-saml
description: >
  Set up a self-contained SAML 2.0 identity provider inside a Run:ai
  environment's own Keycloak instance, for manually testing SSO/SAML features
  (e.g. certificate expiry warnings, IdP metadata handling) end-to-end without
  a real external IdP. Use whenever asked to "configure SAML", "set up SSO for
  testing", or "test SAML" against a Run:ai environment (local-dev or a CI/PR
  preview env).
user-invocable: true
---

# Configure a Local SAML Server for Testing

Every Run:ai environment ships its own Keycloak. This skill uses that same
Keycloak instance as both the SAML **Service Provider** (Run:ai's `runai`
realm, already configured) and, by creating a second realm, as the SAML
**Identity Provider** — so SSO can be fully exercised without any external
IdP or DNS/cert setup.

Requires browser automation (Playwright/MCP browser tools) and:
- Run:ai admin credentials (email/password) for the target environment.
- Keycloak admin credentials for the same environment. Use the
  `keycloak-admin` skill if you have `kubectl` access to the cluster;
  otherwise ask the user for `admin`/`admin`-style temporary admin creds (CI
  preview environments commonly seed a temporary Keycloak admin).

Substitute `<BASE_URL>` below with the environment's URL, e.g.
`https://ci-a707-saml-certificate-expi.runailabs.com` or
`https://runai.local` for a local-dev environment.

## Steps

### 1. Create the IdP realm in Keycloak

1. Log in to `<BASE_URL>/auth/admin/master/console/` with the Keycloak admin
   credentials.
2. **Manage realms → Create realm**. Name it `saml` (matches the reference
   guide other engineers use; any name works as long as you're consistent).
3. Note: navigating directly to a `#/<realm>/...` URL does **not** switch the
   console's active realm state — you must click the realm's link from the
   realm list (or the left-nav realm switcher) for subsequent tabs (Clients,
   Users, ...) to operate on the right realm. Always verify "Current realm"
   in the top-left before acting.
4. In the new realm's **Realm settings → General → Endpoints**, copy the
   **SAML 2.0 Identity Provider Metadata** URL:
   `<BASE_URL>/auth/realms/saml/protocol/saml/descriptor`

### 2. Configure SAML SSO in the Run:ai UI

1. Log in to `<BASE_URL>/` as a Run:ai admin.
2. Admin → General settings → expand **Security** → **Set SSO (Single
   Sign-On)** → **IDENTITY PROVIDER** → **Custom SAML 2.0**.
3. Select **From URL** and paste the metadata descriptor URL from step 1.4.
4. Save. Note the **Redirect URL** shown (contains the auto-generated IdP
   alias, e.g. `.../auth/realms/runai/broker/saml-864549/endpoint`) and the
   **Entity ID** (e.g. `runai-<env-host>/runai`) — both are needed next.

### 3. Create the matching SAML client in the `saml` realm

1. Back in Keycloak, switch to the `saml` realm (click its realm link, don't
   navigate by URL — see step 1.3) → **Clients → Create client**.
2. Client type: **SAML**. Client ID **and** Name: the Entity ID copied in
   step 2.4 (e.g. `runai-<env-host>/runai`).
3. On the Login settings step, set both **Valid redirect URIs** and
   **Master SAML Processing URL** to the Redirect URL copied in step 2.4
   (the `.../broker/<idp-alias>/endpoint` URL).
4. Save.

### 4. Relax signature requirements

In the new client's **Keys** tab, toggle **Client signature required** to
Off (confirm the dialog). Run:ai's SAML SP does not sign its requests, so
leaving this On breaks the flow.

### 5. Add an email attribute mapper

1. Client's **Client scopes** tab → click the `<clientId>-dedicated` scope.
2. **Configure a new mapper → User Attribute**.
3. Set Name = `email`, User Attribute = `email`, SAML Attribute Name =
   `email`. Save.

This is required for Run:ai to receive the user's email via SAML assertion —
without it, SSO users can authenticate but Run:ai can't match them to access
rules.

### 6. Create a test user

In the `saml` realm → **Users → Create user**:
- Username / Email: any test address (e.g. `saml-test-user@run.ai`).
- Toggle **Email verified** on (skips a verification-email step Keycloak
  would otherwise require).
- Save, then go to the **Credentials** tab → **Set password** → toggle
  **Temporary** off (so it's not a forced-reset password) → save.

### 7. Grant the test user access in Run:ai

In the Run:ai UI → **Access → Access rules → New Access rule**:
- Subject: User, email = the test user's email (Run:ai will offer it as
  "(new user)" since they haven't logged in yet).
- Role + Scope: whatever the test scenario needs (e.g. System administrator
  on the top-level cluster scope, to mirror an existing admin rule).

### 8. Test the login

1. Log out of Run:ai:
   `<BASE_URL>/auth/realms/runai/protocol/openid-connect/logout?redirect_uri=<BASE_URL>/`
2. Navigate to `<BASE_URL>/` — the login page now shows a **CONTINUE WITH
   SSO** link/button below the local-login form.
3. Click it → redirects to the `saml` realm's own login page ("Sign in to
   saml") → log in with the test user's credentials.
4. First login may prompt Keycloak's **Update Account Information** /
   VERIFY_PROFILE step (First name / Last name required) — fill it in and
   submit.
5. Should redirect back into Run:ai, fully authenticated as the SSO user.

## Notes for certificate-expiry-style testing

- The `saml` realm's client auto-generates a signing certificate valid for
  ~3 years — freshly created clients won't trigger "expiring soon" or
  "expired" UI states. To test those, either wait out the validity window in
  a lower-level unit test instead, or (last resort) manually import an
  already-expired/near-expired certificate via the client's **Keys → Import
  key** action.
- The certificate/expiry data flows through unmodified — no separate action
  needed to make it visible in Run:ai; it's part of the same SAML metadata
  fetched in step 1.4.
