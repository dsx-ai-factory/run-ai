package orchestrator2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeRegistryJWT builds an unsigned (header.payload.sig) registry token with the
// given subject and granted actions for repo. The prober only decodes the
// payload, so the signature is irrelevant.
func makeRegistryJWT(t *testing.T, subject, repo string, actions []string) string {
	t.Helper()
	claims := map[string]any{"sub": subject}
	if repo != "" {
		claims["access"] = []map[string]any{{"type": "repository", "name": repo, "actions": actions}}
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"none"}`)) + "." + enc(payload) + "." + enc([]byte("sig"))
}

// newFakeTokenRegistry starts a TLS test registry that requires bearer-token auth.
// Its token endpoint authenticates validUser/validPass and grants the requested
// scope (push only when grantPush); any other credentials get an anonymous,
// no-scope token (mimicking Harbor's fallback). It returns the prober, a request
// template pointed at the server, and the path to the server's CA file.
func newFakeTokenRegistry(t *testing.T, repo, validUser, validPass string, grantPush bool) (*httptest.Server, string) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			w.Header().Set("WWW-Authenticate", `Bearer realm="https://`+r.Host+`/token",service="reg"`)
			w.WriteHeader(http.StatusUnauthorized)
		case "/token":
			user, pass, _ := r.BasicAuth()
			var jwt string
			if user == validUser && pass == validPass {
				scope := r.URL.Query().Get("scope")
				var actions []string
				if strings.Contains(scope, "pull") {
					actions = append(actions, "pull")
				}
				if strings.Contains(scope, "push") && grantPush {
					actions = append(actions, "push")
				}
				jwt = makeRegistryJWT(t, validUser, repo, actions)
			} else {
				jwt = makeRegistryJWT(t, "", "", nil) // anonymous fallback
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"token": jwt})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caPath, pemBytes, 0o600); err != nil {
		t.Fatalf("write CA: %v", err)
	}
	return srv, caPath
}

func hostOf(srv *httptest.Server) string {
	return strings.TrimPrefix(srv.URL, "https://")
}

func caFileFor(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	caPath := filepath.Join(t.TempDir(), "ca.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caPath, pemBytes, 0o600); err != nil {
		t.Fatalf("write CA: %v", err)
	}
	return caPath
}

func TestRegistryProber_InvalidCredentials(t *testing.T) {
	srv, ca := newFakeTokenRegistry(t, "runai/probe", "valid", "valid", true)
	prober := NewRegistryProber()

	res, err := prober.Check(context.Background(), RegistryCheckRequest{
		Host: hostOf(srv), Repo: "runai/probe", Username: "wrong", Password: "wrong", CAPath: ca, CheckPush: true,
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if res.Reachable != true || res.Authenticated != false {
		t.Fatalf("expected reachable + not authenticated, got %+v", res)
	}
}

func TestRegistryProber_ValidButCannotPush(t *testing.T) {
	srv, ca := newFakeTokenRegistry(t, "runai/probe", "valid", "valid", false) // pull only
	prober := NewRegistryProber()

	res, err := prober.Check(context.Background(), RegistryCheckRequest{
		Host: hostOf(srv), Repo: "runai/probe", Username: "valid", Password: "valid", CAPath: ca, CheckPush: true,
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !res.Authenticated || !res.CanPull {
		t.Fatalf("expected authenticated + can pull, got %+v", res)
	}
	if !res.PushChecked || res.CanPush {
		t.Fatalf("expected push checked + cannot push, got %+v", res)
	}
}

func TestRegistryProber_ValidAndCanPush(t *testing.T) {
	srv, ca := newFakeTokenRegistry(t, "runai/probe", "valid", "valid", true)
	prober := NewRegistryProber()

	res, err := prober.Check(context.Background(), RegistryCheckRequest{
		Host: hostOf(srv), Repo: "runai/probe", Username: "valid", Password: "valid", CAPath: ca, CheckPush: true,
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !res.Authenticated || !res.CanPull || !res.CanPush {
		t.Fatalf("expected authenticated + pull + push, got %+v", res)
	}
}

// TestRegistryProber_BasicAuth covers the Basic-auth (non-bearer) registry flow.
func TestRegistryProber_BasicAuth(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if user == "valid" && pass == "valid" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	t.Cleanup(srv.Close)
	ca := caFileFor(t, srv)
	prober := NewRegistryProber()

	t.Run("valid", func(t *testing.T) {
		res, err := prober.Check(context.Background(), RegistryCheckRequest{Host: hostOf(srv), Username: "valid", Password: "valid", CAPath: ca})
		if err != nil {
			t.Fatalf("check: %v", err)
		}
		if !res.Authenticated || !res.CanPull {
			t.Fatalf("expected authenticated + can pull, got %+v", res)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		res, err := prober.Check(context.Background(), RegistryCheckRequest{Host: hostOf(srv), Username: "bad", Password: "bad", CAPath: ca})
		if err != nil {
			t.Fatalf("check: %v", err)
		}
		if !res.Reachable || res.Authenticated {
			t.Fatalf("expected reachable + not authenticated, got %+v", res)
		}
	})
}

// TestRegistryProber_HarborProject covers project existence via the Harbor API.
func TestRegistryProber_HarborProject(t *testing.T) {
	newHarbor := func(projectExists bool) (*httptest.Server, string) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/v2/":
				w.Header().Set("WWW-Authenticate", `Bearer realm="https://`+r.Host+`/token",service="reg"`)
				w.WriteHeader(http.StatusUnauthorized)
			case r.URL.Path == "/token":
				_ = json.NewEncoder(w).Encode(map[string]string{"token": makeRegistryJWT(t, "valid", "runai/probe", []string{"pull", "push"})})
			case strings.HasPrefix(r.URL.Path, "/api/v2.0/projects/"):
				if projectExists {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(srv.Close)
		return srv, caFileFor(t, srv)
	}

	t.Run("exists", func(t *testing.T) {
		srv, ca := newHarbor(true)
		res, err := NewRegistryProber().Check(context.Background(), RegistryCheckRequest{
			Host: hostOf(srv), Repo: "runai/probe", Project: "runai", Username: "valid", Password: "valid", CAPath: ca, CheckPush: true,
		})
		if err != nil {
			t.Fatalf("check: %v", err)
		}
		if !res.ProjectChecked || !res.ProjectExists {
			t.Fatalf("expected project checked + exists, got %+v", res)
		}
	})
	t.Run("missing", func(t *testing.T) {
		srv, ca := newHarbor(false)
		res, err := NewRegistryProber().Check(context.Background(), RegistryCheckRequest{
			Host: hostOf(srv), Repo: "runai/probe", Project: "runai", Username: "valid", Password: "valid", CAPath: ca, CheckPush: true,
		})
		if err != nil {
			t.Fatalf("check: %v", err)
		}
		if !res.ProjectChecked || res.ProjectExists {
			t.Fatalf("expected project checked + missing, got %+v", res)
		}
	})
}

// TestRegistryProber_AnonymousAccess covers a registry that allows unauthenticated
// access (GET /v2/ returns 200).
func TestRegistryProber_AnonymousAccess(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)

	res, err := NewRegistryProber().Check(context.Background(), RegistryCheckRequest{Host: hostOf(srv), CAPath: caFileFor(t, srv)})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !res.Reachable || !res.Authenticated || !res.CanPull {
		t.Fatalf("expected reachable + authenticated + can pull, got %+v", res)
	}
}

// TestRegistryProber_Unreachable covers a host that cannot be reached.
func TestRegistryProber_Unreachable(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	host := hostOf(srv)
	ca := caFileFor(t, srv)
	srv.Close() // close so the connection is refused

	res, err := NewRegistryProber().Check(context.Background(), RegistryCheckRequest{Host: host, CAPath: ca})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if res.Reachable {
		t.Fatalf("expected not reachable, got %+v", res)
	}
	if res.Detail == "" {
		t.Fatalf("expected a detail explaining the failure, got %+v", res)
	}
}
