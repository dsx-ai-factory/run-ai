package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/rs/zerolog"
)

type JWKSVerifierConfig struct {
	IssuerURL  string
	HTTPClient *http.Client
	Logger     zerolog.Logger
}

type JWKSVerifier struct {
	issuer  string
	keyfunc jwt.Keyfunc
	cfg     JWKSVerifierConfig
}

type openidConfiguration struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

func NewJWKSVerifier(ctx context.Context, cfg JWKSVerifierConfig) (*JWKSVerifier, error) {
	if cfg.IssuerURL == "" {
		return nil, errors.New("jwks verifier: issuer URL is required")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	discoveryURL := strings.TrimRight(cfg.IssuerURL, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("jwks verifier: build discovery request: %w", err)
	}
	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jwks verifier: fetch %s: %w", discoveryURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks verifier: %s returned %d", discoveryURL, resp.StatusCode)
	}
	var oidc openidConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&oidc); err != nil {
		return nil, fmt.Errorf("jwks verifier: decode discovery doc: %w", err)
	}
	if oidc.JWKSURI == "" || oidc.Issuer == "" {
		return nil, fmt.Errorf("jwks verifier: discovery doc missing issuer or jwks_uri")
	}

	k, err := keyfunc.NewDefaultCtx(ctx, []string{oidc.JWKSURI})
	if err != nil {
		return nil, fmt.Errorf("jwks verifier: init keyfunc for %s: %w", oidc.JWKSURI, err)
	}

	cfg.Logger.Info().
		Str("issuer", oidc.Issuer).
		Str("jwks_uri", oidc.JWKSURI).
		Msg("JWKS verifier initialized")

	return &JWKSVerifier{
		issuer:  oidc.Issuer,
		keyfunc: k.Keyfunc,
		cfg:     cfg,
	}, nil
}

func (v *JWKSVerifier) Issuer() string { return v.cfg.IssuerURL }

func (v *JWKSVerifier) Verify(ctx context.Context, token string, req *http.Request) (*mcpauth.TokenInfo, error) {
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, v.keyfunc,
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		v.cfg.Logger.Warn().Err(err).Msg("token verification failed")
		return nil, fmt.Errorf("%w: %v", mcpauth.ErrInvalidToken, err)
	}
	if !parsed.Valid {
		v.cfg.Logger.Warn().Msg("token marked invalid after parsing")
		return nil, fmt.Errorf("%w: token invalid", mcpauth.ErrInvalidToken)
	}

	exp, _ := claims.GetExpirationTime()
	if exp == nil {
		v.cfg.Logger.Warn().Msg("token missing exp claim")
		return nil, fmt.Errorf("%w: missing exp", mcpauth.ErrInvalidToken)
	}

	sub, _ := claims.GetSubject()
	info := &mcpauth.TokenInfo{
		Expiration: exp.Time,
		UserID:     sub,
		Scopes:     extractScopes(claims),
		Extra:      map[string]any{"raw_token": token},
	}
	return info, nil
}

func extractScopes(claims jwt.MapClaims) []string {
	scope, _ := claims["scope"].(string)
	if scope == "" {
		return nil
	}
	return strings.Fields(scope)
}

func TokenInfoToRequestContext(ctx context.Context) context.Context {
	info := mcpauth.TokenInfoFromContext(ctx)
	if info == nil || info.Extra == nil {
		return ctx
	}
	raw, _ := info.Extra["raw_token"].(string)
	if raw == "" {
		return ctx
	}
	return WithRequestToken(ctx, raw)
}

func RawTokenPropagator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := TokenInfoToRequestContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var _ mcpauth.TokenVerifier = (*JWKSVerifier)(nil).Verify
