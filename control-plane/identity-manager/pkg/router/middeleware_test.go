/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	businessMocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"

	"github.com/stretchr/testify/assert"
)

const (
	MOCK_REQ_URL   = "mock.run.ai"
	REQUEST_METHOD = "GET"
)

func TestExtractRealmNameUnauthorized_SelfHostedHappyFlow(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	middleware := extractRealmNameUnauthenticated("runai")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if val, ok := req.Context().Value(utils.AuthRealmKey).(string); !ok {
			assert.Fail(t, "%s is missing from header", utils.AuthRealmKey)
		} else {
			assert.Equal(t, "runai", val)
		}
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractRealmNameUnauthorized_SaaSHappyFlow(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	req.Host = "test.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	middleware := extractRealmNameUnauthenticatedMultiTenant("Host", dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if val, ok := req.Context().Value(utils.AuthRealmKey).(string); !ok {
			assert.Fail(t, "%s is missing from header", utils.AuthRealmKey)
		} else {
			assert.Equal(t, "test", val)
		}
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractTenantIdUnauthorized_HappyFlow(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	contextWithRealmName := context.WithValue(req.Context(), utils.AuthRealmKey, "test")
	req = req.WithContext(contextWithRealmName)
	req.Host = "test.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByName", req.Context(), "test").Return(&models.Realm{TenantId: "123"}, nil).Once()
	middleware := extractTenantIdUnauthenticated(dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if val, ok := req.Context().Value(runaictx.CfTenantId).(string); !ok {
			assert.Fail(t, "%s is not in context", runaictx.CfTenantId)
		} else {
			assert.Equal(t, "123", val)
		}
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractTenantIdUnauthorized_NoRealmOnCtx(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	req.Host = "test.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	middleware := extractTenantIdUnauthenticated(dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := req.Context().Value(runaictx.CfTenantId).(string); ok {
			assert.Fail(t, "%s should not be in context", runaictx.CfTenantId)
		}
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractTenantIdUnauthorized_ErrorGetRealm(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	contextWithRealmName := context.WithValue(req.Context(), utils.AuthRealmKey, "test")
	req = req.WithContext(contextWithRealmName)
	req.Host = "test.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByName", req.Context(), "test").Return(nil, errors.New("error")).Once()
	middleware := extractTenantIdUnauthenticated(dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Handler should not be called when realm lookup fails")
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestExtractTenantIdUnauthorized_NilRealm(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	contextWithRealmName := context.WithValue(req.Context(), utils.AuthRealmKey, "test")
	req = req.WithContext(contextWithRealmName)
	req.Host = "test.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByName", req.Context(), "test").Return(nil, nil).Once()
	middleware := extractTenantIdUnauthenticated(dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Handler should not be called when realm is nil")
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestExtractTenantIdUnauthorized_NoTenantIdOnRealm(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	contextWithRealmName := context.WithValue(req.Context(), utils.AuthRealmKey, "test")
	req = req.WithContext(contextWithRealmName)
	req.Host = "test.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByName", req.Context(), "test").Return(&models.Realm{}, nil).Once()
	middleware := extractTenantIdUnauthenticated(dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "Handler should not be called when realm has no tenant ID")
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestExtractRealmName_HappyFlow(t *testing.T) {
	tenantId := "12345"
	realmName := "test"
	dalMock := businessMocks.NewRealmDal(t)
	middleware := setAuthRealmFromTenantId(dalMock)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(&models.Realm{TenantId: tenantId, Realm: realmName}, nil)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if val, ok := req.Context().Value(utils.AuthRealmKey).(string); !ok {
			assert.Fail(t, "%s is missing from header", utils.AuthRealmKey)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			assert.Equal(t, realmName, val)
			w.WriteHeader(http.StatusOK)
		}
	})

	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.WithValue(req.Context(), runaictx.CfTenantId, tenantId))
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestExtractRealmName_MissingTenantIdCtx(t *testing.T) {
	dalMock := businessMocks.NewRealmDal(t)
	middleware := setAuthRealmFromTenantId(dalMock)
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := middleware(okHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestExtractRealmName_MissingTenantInDb(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	tenantId := "12345"
	ctx := context.WithValue(req.Context(), runaictx.CfTenantId, tenantId)
	req = req.WithContext(ctx)
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(nil, utils.Error{Code: http.StatusNotFound})

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := setAuthRealmFromTenantId(dalMock)
	rr := httptest.NewRecorder()
	handler := middleware(okHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestExtractRealmName_ErrorInDb(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	tenantId := "12345"
	ctx := context.WithValue(req.Context(), runaictx.CfTenantId, tenantId)
	req = req.WithContext(ctx)
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(nil, utils.Error{Code: http.StatusInternalServerError})

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := setAuthRealmFromTenantId(dalMock)
	rr := httptest.NewRecorder()
	handler := middleware(okHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestExtractRealmName_UnKnownErrorInDb(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	tenantId := "12345"
	ctx := context.WithValue(req.Context(), runaictx.CfTenantId, tenantId)
	req = req.WithContext(ctx)
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(nil, errors.New("not runai error"))

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := setAuthRealmFromTenantId(dalMock)
	rr := httptest.NewRecorder()
	handler := middleware(okHandler)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestExtractRealmNameUnauthenticatedMultiTenant_TenantIdInCtx_RealmFound(t *testing.T) {
	tenantId := "12345"
	realmName := "my-realm"
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.WithValue(req.Context(), runaictx.CfTenantId, tenantId))
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(&models.Realm{TenantId: tenantId, Realm: realmName}, nil)
	middleware := extractRealmNameUnauthenticatedMultiTenant("Host", dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		val, ok := req.Context().Value(utils.AuthRealmKey).(string)
		assert.True(t, ok, "AuthRealmKey should be set in context")
		assert.Equal(t, realmName, val)
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractRealmNameUnauthenticatedMultiTenant_TenantIdInCtx_RealmLookupError_SkipsHostFallback(t *testing.T) {
	tenantId := "12345"
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "fallback.run.ai"
	req = req.WithContext(context.WithValue(req.Context(), runaictx.CfTenantId, tenantId))
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(nil, errors.New("db error"))
	middleware := extractRealmNameUnauthenticatedMultiTenant("Host", dalMock)
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		_, ok := req.Context().Value(utils.AuthRealmKey).(string)
		assert.False(t, ok, "AuthRealmKey should not be set when realm lookup fails")
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
	assert.True(t, handlerCalled, "next handler should be called")
}

func TestExtractRealmNameUnauthenticatedMultiTenant_TenantIdInCtx_NilRealm_FallsBackToHost(t *testing.T) {
	tenantId := "12345"
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "hostfallback.run.ai"
	req = req.WithContext(context.WithValue(req.Context(), runaictx.CfTenantId, tenantId))
	dalMock := businessMocks.NewRealmDal(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, tenantId).Return(nil, nil)
	middleware := extractRealmNameUnauthenticatedMultiTenant("Host", dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		val, ok := req.Context().Value(utils.AuthRealmKey).(string)
		assert.True(t, ok, "AuthRealmKey should be set from host fallback")
		assert.Equal(t, "hostfallback", val)
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractRealmNameUnauthenticatedMultiTenant_NoTenantId_NoHost(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = ""
	dalMock := businessMocks.NewRealmDal(t)
	middleware := extractRealmNameUnauthenticatedMultiTenant("Host", dalMock)
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		_, ok := req.Context().Value(utils.AuthRealmKey).(string)
		assert.False(t, ok, "AuthRealmKey should not be set when no tenantId and no host")
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
	assert.True(t, handlerCalled, "next handler should be called")
}

func TestExtractRealmNameUnauthorized_CustomHostHeader(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	req.Header.Set("X-Forwarded-Host", "custom.run.ai")
	req.Host = "fallback.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	middleware := extractRealmNameUnauthenticatedMultiTenant("X-Forwarded-Host", dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if val, ok := req.Context().Value(utils.AuthRealmKey).(string); !ok {
			assert.Fail(t, "%s is missing from header", utils.AuthRealmKey)
		} else {
			assert.Equal(t, "custom", val)
		}
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestExtractRealmNameUnauthorized_CustomHostHeaderFallback(t *testing.T) {
	req, err := http.NewRequest(REQUEST_METHOD, MOCK_REQ_URL, nil)
	// Don't set the custom header, should fallback to r.Host
	req.Host = "fallback.run.ai"
	if err != nil {
		t.Fatal(err)
	}
	dalMock := businessMocks.NewRealmDal(t)
	middleware := extractRealmNameUnauthenticatedMultiTenant("X-Forwarded-Host", dalMock)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if val, ok := req.Context().Value(utils.AuthRealmKey).(string); !ok {
			assert.Fail(t, "%s is missing from header", utils.AuthRealmKey)
		} else {
			assert.Equal(t, "fallback", val)
		}
	})
	rr := httptest.NewRecorder()
	handler := middleware(testHandler)
	handler.ServeHTTP(rr, req)
}

func TestCookieToHeaderMiddleware_WithValidCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(utils.RefreshTokenFromCookieKey).(string)
		assert.True(t, ok)
		assert.Equal(t, "valid-refresh-token", token)
		w.WriteHeader(http.StatusOK)
	})

	middleware := cookieToHeaderMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, utils.RefreshTokenCookiePath, nil)
	req.AddCookie(&http.Cookie{Name: utils.RefreshTokenCookieName, Value: "valid-refresh-token"})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCookieToHeaderMiddleware_NoCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(utils.RefreshTokenFromCookieKey).(string)
		assert.False(t, ok)
		assert.Empty(t, token)
		w.WriteHeader(http.StatusOK)
	})

	middleware := cookieToHeaderMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, utils.RefreshTokenCookiePath, nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCookieToHeaderMiddleware_EmptyCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(utils.RefreshTokenFromCookieKey).(string)
		assert.False(t, ok)
		assert.Empty(t, token)
		w.WriteHeader(http.StatusOK)
	})

	middleware := cookieToHeaderMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, utils.RefreshTokenCookiePath, nil)
	req.AddCookie(&http.Cookie{Name: utils.RefreshTokenCookieName, Value: ""})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCookieToHeaderMiddleware_AlwaysProcessesCookie(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(utils.RefreshTokenFromCookieKey).(string)
		assert.True(t, ok)
		assert.Equal(t, "test-token", token)
		w.WriteHeader(http.StatusOK)
	})

	middleware := cookieToHeaderMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/any/path", nil)
	req.AddCookie(&http.Cookie{Name: utils.RefreshTokenCookieName, Value: "test-token"})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBasicAuthClientCredsMiddleware_InjectsClientIdAndSecret(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "myclient", r.FormValue("client_id"))
		assert.Equal(t, "mysecret", r.FormValue("client_secret"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := basicAuthClientCredsMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/token", nil)
	req.SetBasicAuth("myclient", "mysecret")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBasicAuthClientCredsMiddleware_NoAuthHeader_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.FormValue("client_id"))
		assert.Empty(t, r.FormValue("client_secret"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := basicAuthClientCredsMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/token", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBasicAuthClientCredsMiddleware_FormValuesTakePrecedence(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "form-client", r.FormValue("client_id"))
		assert.Equal(t, "form-secret", r.FormValue("client_secret"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := basicAuthClientCredsMiddleware(handler)

	body := strings.NewReader("client_id=form-client&client_secret=form-secret")
	req := httptest.NewRequest(http.MethodPost, "/api/v2/token", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("basic-client", "basic-secret")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBasicAuthClientCredsMiddleware_PercentEncodedCredentials_AreDecoded(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "client:id", r.FormValue("client_id"))
		assert.Equal(t, "secret with space", r.FormValue("client_secret"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := basicAuthClientCredsMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/token", nil)
	req.SetBasicAuth("client%3Aid", "secret+with+space")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBasicAuthClientCredsMiddleware_NoPassword_OnlyInjectsClientId(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "publicclient", r.FormValue("client_id"))
		assert.Empty(t, r.FormValue("client_secret"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := basicAuthClientCredsMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/token", nil)
	req.SetBasicAuth("publicclient", "")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCookieToHeaderMiddleware_ContextPropagation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		existingValue := r.Context().Value("existing-key")
		assert.Equal(t, "existing-value", existingValue)

		token, ok := r.Context().Value(utils.RefreshTokenFromCookieKey).(string)
		assert.True(t, ok)
		assert.Equal(t, "token-value", token)
		w.WriteHeader(http.StatusOK)
	})

	middleware := cookieToHeaderMiddleware(handler)

	ctx := context.WithValue(context.Background(), "existing-key", "existing-value")
	req := httptest.NewRequest(http.MethodPost, utils.RefreshTokenCookiePath, nil)
	req = req.WithContext(ctx)
	req.AddCookie(&http.Cookie{Name: utils.RefreshTokenCookieName, Value: "token-value"})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
