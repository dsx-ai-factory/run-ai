package router

import (
	"context"
	"net/http"
	"net/url"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func setRequestHost(hostHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.WithCtx(runaictx.LoggerHook(r.Context())).Debugf("setRequestHost middleware")
			host := r.Header.Get(hostHeader)
			if host == "" {
				host = r.Host
			}
			withHost := context.WithValue(r.Context(), utils.RequestHostKey, host) //nolint:staticcheck // SA1029
			next.ServeHTTP(w, r.WithContext(withHost))
		})
	}
}

func extractRealmNameUnauthenticated(realmName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.WithCtx(runaictx.LoggerHook(r.Context())).Debugf("ExtractRealmName middleware: selfHosted deployment")
			withRealmName := context.WithValue(r.Context(), utils.AuthRealmKey, realmName)
			next.ServeHTTP(w, r.WithContext(withRealmName))
		})
	}
}

func extractTenantIdUnauthenticated(realmDal business.RealmDal) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.WithCtx(runaictx.LoggerHook(r.Context())).Debugf("ExtractTenantIdUnauthenticated middleware")
			if r.Context().Value(utils.AuthRealmKey) == nil {
				logging.WithCtx(runaictx.LoggerHook(r.Context())).Debugf("ExtractTenantIdUnauthenticated middleware: missing realm name in ctx")
				audit.AuditFromCtx(r.Context()).Drop()
				next.ServeHTTP(w, r)
				return
			}
			realmName := r.Context().Value(utils.AuthRealmKey).(string)
			realm, err := realmDal.GetRealmByName(r.Context(), realmName)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(r.Context())).Warnf("ExtractTenantIdUnauthenticated middleware: failed to find realm for realm name %s. err: %v", realmName, err)
				audit.AuditFromCtx(r.Context()).Drop()
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if realm == nil || realm.TenantId == "" {
				logging.WithCtx(runaictx.LoggerHook(r.Context())).Warnf("ExtractTenantIdUnauthenticated middleware: missing tenantId for realm name %s. realm: %v", realmName, realm)
				audit.AuditFromCtx(r.Context()).Drop()
				w.WriteHeader(http.StatusNotFound)
				return
			}
			withTenantId := context.WithValue(r.Context(), runaictx.CfTenantId, realm.TenantId)
			next.ServeHTTP(w, r.WithContext(withTenantId))
		})
	}
}

func setAuthRealmFromTenantId(realmDal business.RealmDal) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
			if tenantId == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("middleware: no tenant id extract - failed to set auth realm in ctx")
				next.ServeHTTP(w, r)
				return
			} else {
				realm, err := realmDal.GetRealmByTenantId(ctx, *tenantId)
				if err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("middleware: failed to find realm for tenant id %s. err: %v", *tenantId, err)
					if errorHandler, ok := err.(utils.Error); ok {
						if errorHandler.Code == http.StatusNotFound {
							w.WriteHeader(http.StatusNotFound)
						} else {
							w.WriteHeader(http.StatusBadRequest)
						}
					} else {
						logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("middleware: internal error parsing res for tenant id: %s. err: %v", *tenantId, err)
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				}
				if realm == nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("middleware: missing realm for tenant id %s. err: %v", *tenantId, err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("middleware: found realm for tenant id %s. realm: %v", *tenantId, realm)
				withRealmName := context.WithValue(r.Context(), utils.AuthRealmKey, realm.Realm)
				next.ServeHTTP(w, r.WithContext(withRealmName))
				return
			}
		})
	}
}

func extractRealmNameUnauthenticatedMultiTenant(hostHeader string, realmDal business.RealmDal) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			realmName, resolved, err := resolveRealmNameByTenantId(ctx, realmDal)
			if !resolved && err == nil {
				realmName, resolved = resolveRealmNameByHostname(ctx, r, hostHeader)
			}

			if resolved {
				withRealmName := context.WithValue(r.Context(), utils.AuthRealmKey, realmName)
				next.ServeHTTP(w, r.WithContext(withRealmName))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolveRealmNameByHostname(ctx context.Context, r *http.Request, hostHeader string) (string, bool) {
	host := r.Header.Get(hostHeader)
	if host == "" {
		host = r.Host
	}
	if host == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("middleware: no realm name to extract - missing host")
		return "", false
	}
	realmName := utils.GetTenantNameFromHost(host)
	if realmName != "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("middleware: extracted realm from host %s", realmName)
		return realmName, true
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("middleware: no realm name to extract - host invalid")
	return "", false
}

func cookieToHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(utils.RefreshTokenCookieName); err == nil && cookie.Value != "" {
			ctx := context.WithValue(r.Context(), utils.RefreshTokenFromCookieKey, cookie.Value) //nolint:staticcheck // SA1029
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// basicAuthClientCredsMiddleware extracts client_id / client_secret from the
// Authorization: Basic header (client_secret_basic, RFC 6749 §2.3.1) and
// injects them into the request form, so the token handler that reads only
// form fields can pick them up. Existing form values take precedence, preserving
// client_secret_post behaviour.
func basicAuthClientCredsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username == "" {
			next.ServeHTTP(w, r)
			return
		}
		clientID, err := url.QueryUnescape(username)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(r.Context())).Errorf("basicAuthClientCredsMiddleware: failed to decode client_id: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		clientSecret, err := url.QueryUnescape(password)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(r.Context())).Errorf("basicAuthClientCredsMiddleware: failed to decode client_secret: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			logging.WithCtx(runaictx.LoggerHook(r.Context())).Errorf("basicAuthClientCredsMiddleware: failed to parse form: %v", err)
			next.ServeHTTP(w, r)
			return
		}
		if r.PostForm.Get("client_id") == "" {
			r.PostForm.Set("client_id", clientID)
			r.Form.Set("client_id", clientID)
		}
		if clientSecret != "" && r.PostForm.Get("client_secret") == "" {
			r.PostForm.Set("client_secret", clientSecret)
			r.Form.Set("client_secret", clientSecret)
		}
		next.ServeHTTP(w, r)
	})
}

func resolveRealmNameByTenantId(ctx context.Context, realmDal business.RealmDal) (realmName string, resolved bool, err error) {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("middleware: tenantId is nil")
		return "", false, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("middleware: tenantId: %s", *tenantId)

	realm, err := realmDal.GetRealmByTenantId(ctx, *tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("middleware: failed to find realm for tenant id %s. err: %v", *tenantId, err)
		return "", false, err
	}
	if realm != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("middleware: extracted realm from tenant id %s", *tenantId)
		return realm.Realm, true, nil
	}
	return "", false, nil
}
