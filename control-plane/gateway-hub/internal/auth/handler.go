/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	runaictxhttp "github.com/run-ai/runai/runai-common-packages/runaictx/http"
)

const (
	subjectTypeUser           = "User"
	subjectTypeServiceAccount = "ServiceAccount"
	clientTypePersonalApp     = "personal-application"

	maxBearerTokenLen = 16 * 1024
)

type Handler struct {
	keycloak        *KeycloakClient
	managementRealm string
}

func NewHandler(keycloakURL, managementRealm string) *Handler {
	return &Handler{
		keycloak:        NewKeycloakClient(keycloakURL),
		managementRealm: managementRealm,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.WithCtx(runaictx.LoggerHook(ctx))

	tokenStr, err := extractBearerToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "authorization header is missing or invalid")
		return
	}

	// Security ordering: the claims are parsed WITHOUT signature verification here
	// so we can read the issuer and select the correct realm public key. Only the
	// issuer is trusted at this point (and only to look up a key on the fixed,
	// configured Keycloak host - see RealmFromIssuer). Every other claim is treated
	// as untrusted until verifySignature succeeds below; no identity header is set
	// unless verification passes.
	claims, err := parseUnverified(tokenStr)
	if err != nil {
		log.Errorf("failed to parse JWT claims: %v", err)
		respondError(w, http.StatusUnauthorized, "token could not be parsed")
		return
	}

	realm, err := RealmFromIssuer(claims.Issuer)
	if err != nil {
		log.Errorf("invalid issuer in JWT: %v", err)
		respondError(w, http.StatusUnauthorized, "token has invalid issuer")
		return
	}

	pubKey, err := h.keycloak.GetPublicKey(ctx, realm)
	if err != nil {
		log.Errorf("failed to get Keycloak public key for realm %s: %v", realm, err)
		respondError(w, http.StatusUnauthorized, "failed to get public key")
		return
	}

	if err := verifySignature(tokenStr, pubKey); err != nil {
		log.Errorf("JWT signature invalid: %v", err)
		respondError(w, http.StatusUnauthorized, "signature is invalid")
		return
	}

	if err := claims.valid(); err != nil {
		log.Errorf("JWT time claims invalid: %v", err)
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if claims.Subject == "" || int(claims.TenantId) == 0 {
		respondError(w, http.StatusUnauthorized, "missing claim")
		return
	}

	isManagement := h.managementRealm != "" && realm == h.managementRealm
	setResponseHeaders(w, ctx, claims, isManagement)
	w.WriteHeader(http.StatusOK)
}

func setResponseHeaders(w http.ResponseWriter, ctx context.Context, claims *RunaiClaims, isManagement bool) {
	clientId := ""
	if isPersonalApp(claims) {
		clientId = claims.ClientId
	}

	hctx := context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", int(claims.TenantId)))
	hctx = context.WithValue(hctx, runaictx.CfSubjectID, claims.Subject)
	hctx = context.WithValue(hctx, runaictx.CfSubjectType, subjectType(claims))
	hctx = context.WithValue(hctx, runaictx.CfGId, claims.GID)
	hctx = context.WithValue(hctx, runaictx.CfUId, claims.UID)
	hctx = context.WithValue(hctx, runaictx.CfSId, claims.SID)
	hctx = context.WithValue(hctx, runaictx.CfSupplementalGroups, intSliceToStr(claims.SupplementaryGroups))
	hctx = context.WithValue(hctx, runaictx.CfIsManagementTenant, fmt.Sprintf("%t", isManagement))
	hctx = context.WithValue(hctx, runaictx.CfClientId, clientId)
	if len(claims.Groups) > 0 {
		hctx = context.WithValue(hctx, runaictx.CfGroups, []string{strings.Join(claims.Groups, ",")})
	}

	for name, values := range *runaictxhttp.GetHeaders(hctx) {
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	v := r.Header.Get("Authorization")
	for _, prefix := range []string{"Bearer ", "bearer "} {
		if strings.HasPrefix(v, prefix) {
			token := strings.TrimPrefix(v, prefix)
			if token == "" || len(token) > maxBearerTokenLen {
				return "", fmt.Errorf("authorization header is missing or invalid")
			}
			return token, nil
		}
	}
	return "", fmt.Errorf("authorization header is missing or invalid")
}

func splitJWT(tokenStr string) ([]string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT")
	}
	return parts, nil
}

// parseUnverified decodes the JWT payload WITHOUT verifying the signature.
//
// Security contract: the returned claims are untrusted. They must only be used to
// select the verification key (via the issuer), never to make an authorization or
// identity decision, until verifySignature has succeeded for the same token. See
// ServeHTTP for the required ordering and Sonar S5659.
func parseUnverified(tokenStr string) (*RunaiClaims, error) {
	parts, err := splitJWT(tokenStr)
	if err != nil {
		return nil, err
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}
	claims := &RunaiClaims{}
	if err := json.Unmarshal(payload, claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}
	return claims, nil
}

// verifySignature verifies the JWT signature and pins the algorithm to the RSA family.
//
// The signing method is resolved from the token header and type-asserted to
// *jwt.SigningMethodRSA, which rejects "none" (unsigned tokens) and any non-RSA
// algorithm such as HMAC - preventing alg:none bypass and RS256->HS256 key-confusion
// attacks (Sonar S5659). Only RS256/RS384/RS512 signed with the realm's RSA public
// key are accepted.
func verifySignature(tokenStr string, pub *rsa.PublicKey) error {
	parts, err := splitJWT(tokenStr)
	if err != nil {
		return err
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode JWT header: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return fmt.Errorf("failed to parse JWT header: %w", err)
	}

	method, ok := jwt.GetSigningMethod(header.Alg).(*jwt.SigningMethodRSA)
	if !ok {
		return fmt.Errorf("unsupported signing algorithm: %s (expected RS256, RS384, or RS512)", header.Alg)
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode JWT signature: %w", err)
	}

	signingInput := parts[0] + "." + parts[1]
	return method.Verify(signingInput, sig, pub)
}

func subjectType(c *RunaiClaims) string {
	if c.ClientId != "" {
		return subjectTypeServiceAccount
	}
	return subjectTypeUser
}

func isPersonalApp(c *RunaiClaims) bool {
	return subjectType(c) == subjectTypeServiceAccount && c.ClientType == clientTypePersonalApp
}

func intSliceToStr(ints []int) string {
	parts := make([]string, len(ints))
	for i, v := range ints {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ",")
}

func respondError(w http.ResponseWriter, code int, msg string) {
	body, _ := json.Marshal(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{Code: code, Message: msg})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}
