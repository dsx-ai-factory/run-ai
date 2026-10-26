/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"

	"github.com/MicahParks/keyfunc"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc/userinfo"
	"github.com/modern-go/reflect2"
	"github.com/rs/zerolog/log"

	"github.com/golang-jwt/jwt/v5"
)

const (
	authorizationHeader    = "Authorization"
	bearerPrefix           = "Bearer "
	wsAuthorizationHeader  = "Sec-Websocket-Protocol"
	wsBearerPrefix         = "auth.bearer."
	wsBearerSuffix         = ", Bearer"
	tokenExtracted         = "successfully extracted token: %s"
	clientsSuffix          = "@clients"
	missingClaimsError     = "token is missing the 'email' and 'clientId' claims"
	subjectNotFoundMessage = "Subject was not found, tried to read by sub field"
	uidNumberClaimMissing  = "Token is missing 'uid/uidNumber' claim."
	gidNumberClaimMissing  = "Token is missing 'gid/gidNumber' claim."
	debugGetUserDetails    = "getUserDetailsFromToken %v"
)

var (
	errMissingAuthHeader = errors.New("missing authorization header")
	errTokenIsEmpty      = errors.New("token is empty")
	errInvalidAuthHeader = errors.New("invalid authorization header")
)

type RunAiClaims struct {
	ExpiresAt         int64            `json:"exp,omitempty"`
	IssuedAt          int64            `json:"iat,omitempty"`
	NotBefore         int64            `json:"nbf,omitempty"`
	Audience          jwt.ClaimStrings `json:"aud,omitempty"`
	Issuer            string           `json:"iss,omitempty"`
	Subject           string           `json:"sub,omitempty"`
	Nickname          string           `json:"nickname,omitempty"`
	Name              string           `json:"name,omitempty"`
	Picture           string           `json:"picture,omitempty"`
	UpdatedAt         string           `json:"updated_at,omitempty"`
	Email             string           `json:"email,omitempty"`
	EmailVerified     bool             `json:"email_verified,omitempty"`
	Gid               interface{}      `json:"gid,omitempty"`
	Uid               interface{}      `json:"uid,omitempty"`
	GidNumber         interface{}      `json:"gidNumber,omitempty"`
	UidNumber         interface{}      `json:"uidNumber,omitempty"`
	Groups            []string         `json:"groups,omitempty"`
	ClientId          string           `json:"clientId,omitempty"`
	PreferredUsername string           `json:"preferred_username,omitempty"`
}

func (rac *RunAiClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	timeValue := time.Unix(rac.ExpiresAt, 0)
	return jwt.NewNumericDate(timeValue), nil
}

func (rac *RunAiClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	timeValue := time.Unix(rac.IssuedAt, 0)
	return jwt.NewNumericDate(timeValue), nil
}

func (rac *RunAiClaims) GetNotBefore() (*jwt.NumericDate, error) {
	timeValue := time.Unix(rac.NotBefore, 0)
	return jwt.NewNumericDate(timeValue), nil
}

func (rac *RunAiClaims) GetIssuer() (string, error) {
	return rac.Issuer, nil
}

func (rac *RunAiClaims) GetSubject() (string, error) {
	return rac.Subject, nil
}

func (rac *RunAiClaims) GetAudience() (jwt.ClaimStrings, error) {
	return rac.Audience, nil
}

func (rac *RunAiClaims) Valid() error {
	// require by the jwks package
	return nil
}

// Get the following details from the token:
//   - username: the name of the user who sent the request, based on name/email/nickname claims
//   - uid, gid: pointers to (optional) user ID and group ID of the user who sent the request
//
// Unlike uid/gid for which we issue warning in case they are missing from the token, username
// is a mandatory claim which result in error in case it is missing
func getUserDetailsFromClaims(claims *RunAiClaims) (*Info, error) {
	info := &Info{}

	info.PreferredUsername = claims.PreferredUsername

	username := ""

	var err error

	if claims.Email != "" {
		username = claims.Email
	} else if claims.ClientId != "" {
		username = claims.ClientId + clientsSuffix
	} else {
		err = errors.New(missingClaimsError)
	}

	info.Username = username

	info.Subject = claims.Subject
	info.OriginalSubject = claims.Subject

	if len(info.Subject) == 0 {
		log.Debug().Msg(subjectNotFoundMessage)
	}

	if !reflect2.IsNil(claims.UidNumber) {
		info.Uid = claimToInt64(claims.UidNumber, "uidNumber")
	} else if !reflect2.IsNil(claims.Uid) {
		info.Uid = claimToInt64(claims.Uid, "uid")
	} else {
		log.Debug().Msg(uidNumberClaimMissing)
	}

	if !reflect2.IsNil(claims.GidNumber) {
		info.Gid = claimToInt64(claims.GidNumber, "gidNumber")
	} else if !reflect2.IsNil(claims.Gid) {
		info.Gid = claimToInt64(claims.Gid, "gid")
	} else {
		log.Debug().Msg(gidNumberClaimMissing)
	}

	if !reflect2.IsNil(claims.Groups) && len(claims.Groups) > 0 {
		info.Groups = claims.Groups
	}

	log.Debug().Msgf(debugGetUserDetails, info)

	return info, err
}

func getGroupsFromUserInfoProvider(token string, userInfoProvider userinfo.Provider) ([]string, error) {
	userInfo, err := userInfoProvider.Load(token)
	if err != nil {
		return nil, fmt.Errorf("failed loading userinfo: %v", err)
	}

	return userInfo.Groups, nil
}

func GetAuthInfo(jwksConfig *keyfunc.JWKS, userInfoProvider userinfo.Provider, token string) (*Info, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	claims, err := validateTokenAndExtractClaims(token, jwksConfig)
	if err != nil {
		return nil, err
	}

	result, err := getUserDetailsFromClaims(claims)
	if err != nil {
		log.Warn().Msgf("Error in getting user details from token: %v", err)
		return nil, err
	}

	if len(result.Groups) == 0 && !isApplication(result) {
		result.Groups, err = getGroupsFromUserInfoProvider(token, userInfoProvider)
		if err != nil {
			log.Warn().Msgf("Error in getting user details from userinfo provider: %v", err)
			return nil, err
		}
	}

	result.RawToken = token

	log.Debug().Msgf("Extracted user info from token: %v\n", result)

	return result, nil
}

func isApplication(claims *Info) bool {
	return strings.HasPrefix(claims.PreferredUsername, "service-account")
}

// validateTokenAndExtractClaims validates a given token using the keys fetched from the given jwksURL.
// In our case, the jwksURL should point to Keycloak's JWKS endpoint that contains
// the public keys that should be used to verify Keycloak's tokens.
// The JWKS return value represents the JWKS refresh worker and should be terminated by
// calling its EndBackground() method when no longer needed
func validateTokenAndExtractClaims(encodedToken string, jwks *keyfunc.JWKS) (*RunAiClaims, error) {
	claims := &RunAiClaims{}
	_, err := jwt.ParseWithClaims(encodedToken, claims, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, fmt.Errorf("token was signed with an invalid algorithm")
		}

		unvalidatedKeyID, exists := token.Header["kid"]
		if !exists {
			return nil, fmt.Errorf("could not find key ID (kid) in token header")
		}

		keyID, ok := unvalidatedKeyID.(string)
		if !ok {
			return nil, fmt.Errorf("token header contains non-string key ID (kid)")
		}

		key, exists := jwks.ReadOnlyKeys()[keyID]
		if !exists {
			return nil, fmt.Errorf("key with kid [%s] was not found", keyID)
		}

		return key, nil
	})
	// Suppress WSL error on the next line
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// utility function to convert claim containing either numeric or string to int64
// any error will be logged to debug, and return nil pointer.
// to be versatile, we support both uid=123 and uid="123"
// because we cannot be sure how the token will look like, it changes b/w environments
func claimToInt64(numberOrString interface{}, claimName string) *int64 {
	var result int64

	switch realValue := numberOrString.(type) {
	case int, int32, int64, float32, float64:
		result = int64(reflect.ValueOf(realValue).Float())
	case string:
		var err error

		result, err = strconv.ParseInt(realValue, 10, 0)
		if err != nil {
			log.Debug().Msgf("claim %s value %v: %s", claimName, realValue, err.Error())
			return nil
		}
	default:
		log.Debug().Msgf("claim %s value %v: unexpected type", claimName, realValue)
		return nil
	}

	return &result
}

func ExtractAuthTokenFromRequest(r *http.Request) (string, error) {
	token, err := getToken(r)
	if err != nil {
		return "", err
	}

	if token == nil {
		return "", errTokenIsEmpty
	}

	log.Debug().Msg(tokenExtracted)

	return *token, nil
}

func getToken(r *http.Request) (*string, error) {
	token, err := getAuthorizationHeader(r)
	if err != nil && !errors.Is(err, api.NotFoundError{}) {
		return nil, err
	}

	if token != nil {
		return token, nil
	}

	token, err = geWebSocketAuthorizationHeader(r)
	if err != nil {
		if errors.Is(err, api.NotFoundError{}) {
			return nil, errMissingAuthHeader
		}

		return nil, err
	}

	return token, nil
}

func getAuthorizationHeader(r *http.Request) (*string, error) {
	authHeaders := r.Header[authorizationHeader]
	if len(authHeaders) == 0 {
		return nil, api.NotFoundError{}
	}

	hdr := authHeaders[0]
	hdrParts := strings.Split(hdr, bearerPrefix)

	if len(hdrParts) > 1 {
		if len(hdrParts[1]) == 0 {
			return nil, errTokenIsEmpty
		}

		return &hdrParts[1], nil
	}

	return nil, errInvalidAuthHeader
}

func geWebSocketAuthorizationHeader(r *http.Request) (*string, error) {
	authHeaders := r.Header[wsAuthorizationHeader]
	if len(authHeaders) == 0 {
		return nil, api.NotFoundError{}
	}

	hdr := authHeaders[0]
	hdr = strings.TrimPrefix(hdr, wsBearerPrefix)
	hdr = strings.TrimSuffix(hdr, wsBearerSuffix)

	if len(hdr) == 0 {
		return nil, errTokenIsEmpty
	}

	return &hdr, nil
}
