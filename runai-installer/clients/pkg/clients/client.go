// Package clients provides a shared RunAI API client with automatic token rotation.
// This package can be used by both the installer and provisioner-service.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/run-ai/runai/runai-installer/clients/pkg/tls"
	"github.com/run-ai/runai/runai-installer/clients/runai/authorization"
	"github.com/run-ai/runai/runai-installer/clients/runai/cluster"
	"github.com/run-ai/runai/runai-installer/clients/runai/orgunit"
	idmclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"golang.org/x/oauth2"
	"k8s.io/utils/ptr"
)

// RunAIClient wraps the generated RunAI API clients with automatic token rotation
type RunAIClient struct {
	authClient    *authorization.APIClient
	clusterClient *cluster.APIClient
	orgunitClient *orgunit.APIClient
	idmClient     *idmclient.APIClient
	fqdn          string
}

// Credentials holds the username and password for authentication
type Credentials struct {
	Username string
	Password string
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	Name     string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s' not found", e.Resource, e.Name)
}

// Is allows errors.Is to match any NotFoundError
func (e *NotFoundError) Is(target error) bool {
	_, ok := target.(*NotFoundError)
	return ok
}

// Logger interface for optional logging
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// noopLogger is a no-op implementation of Logger
type noopLogger struct{}

func (noopLogger) Infof(format string, args ...interface{})  {}
func (noopLogger) Errorf(format string, args ...interface{}) {}

var (
	clientInstance *RunAIClient
	clientOnce     sync.Once
	clientMu       sync.Mutex
	clientFQDN     string
	logger         Logger = noopLogger{}
)

// SetLogger sets a custom logger for the client package
func SetLogger(l Logger) {
	logger = l
}

// TokenSource implements oauth2.TokenSource for automatic token refresh
type TokenSource struct {
	tokenApi   idmclient.TokensAPI
	user       string
	password   string
	oauthToken *oauth2.Token
	serverURL  string
	mu         sync.Mutex
}

// Token returns a valid OAuth token, refreshing if necessary
func (ts *TokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.oauthToken == nil || ts.oauthToken.Expiry.Before(time.Now()) {
		logger.Infof("Refreshing OAuth token for Run:AI Control Plane at %s (user: %s)", ts.serverURL, ts.user)
		request := idmclient.TokenRequest{
			Username:  ptr.To(ts.user),
			Password:  ptr.To(ts.password),
			GrantType: idmclient.TokenRequestGrantTypeEnum("password").Ptr(),
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "user-agent", "runai-client/1.0.0")
		token, _, err := ts.tokenApi.GrantToken(ctx).TokenRequest(request).Execute()
		if err != nil {
			logger.Errorf("Failed to refresh token: %v", err)
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		expiration, err := getTokenExpiration(token)
		if err != nil {
			logger.Errorf("Failed to get token expiration: %v", err)
			return nil, fmt.Errorf("failed to get token expiration: %w", err)
		}
		ts.oauthToken = &oauth2.Token{
			AccessToken:  token.GetAccessToken(),
			RefreshToken: token.GetRefreshToken(),
			Expiry:       *expiration,
		}
		logger.Infof("OAuth token refreshed successfully, expires at %s", expiration.Format(time.RFC3339))
	}
	return ts.oauthToken, nil
}

// getTokenExpiration extracts the expiration time from a JWT token.
// IMPORTANT: This function uses ParseUnverified which does NOT validate the token signature.
// This is SAFE because:
// 1. We only extract the 'exp' (expiration) claim to determine if token refresh is needed
// 2. We never use this token for authentication or authorization decisions
// 3. The actual token validation happens server-side when making API calls
// 4. The token was originally issued by the OAuth provider and we trust its content for timing purposes only
func getTokenExpiration(token *idmclient.TokenResponse) (*time.Time, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(*token.AccessToken, claims)
	if err != nil {
		logger.Errorf("Failed to parse token: %v", err)
		return nil, err
	}

	// Extract and validate the expiration claim
	exp, ok := claims["exp"].(float64)
	if !ok {
		logger.Errorf("Invalid or missing 'exp' claim in token")
		return nil, fmt.Errorf("exp claim not found or invalid type in JWT token")
	}

	timeObj := time.Unix(int64(exp), 0)
	return ptr.To(timeObj), nil
}

// GetClient returns a singleton RunAI client for the given FQDN
// The client handles automatic token rotation
func GetClient(fqdn string, creds Credentials) (*RunAIClient, error) {
	return GetClientWithCA(fqdn, creds, "")
}

// GetClientWithCA returns a singleton RunAI client for the given FQDN with optional CA certificate
// If caCertPath is provided, the client will trust that CA for TLS connections
func GetClientWithCA(fqdn string, creds Credentials, caCertPath string) (*RunAIClient, error) {
	clientMu.Lock()
	defer clientMu.Unlock()

	// If we have an existing client for a different FQDN, reset the singleton
	if clientInstance != nil && clientFQDN != fqdn {
		logger.Infof("FQDN changed from %s to %s, resetting RunAI client", clientFQDN, fqdn)
		clientInstance = nil
		clientOnce = sync.Once{}
	}

	var initErr error
	clientOnce.Do(func() {
		clientFQDN = fqdn
		clientInstance, initErr = initClient(fqdn, creds, caCertPath)
	})

	if initErr != nil {
		// Reset once so next call can try again
		clientOnce = sync.Once{}
		return nil, initErr
	}

	return clientInstance, nil
}

// NewClient creates a new RunAI client without singleton behavior
// Use this when you need a separate client instance
func NewClient(fqdn string, creds Credentials) (*RunAIClient, error) {
	return NewClientWithCA(fqdn, creds, "")
}

// NewClientWithCA creates a new RunAI client with optional CA certificate
// If caCertPath is provided, the client will trust that CA for TLS connections
func NewClientWithCA(fqdn string, creds Credentials, caCertPath string) (*RunAIClient, error) {
	return initClient(fqdn, creds, caCertPath)
}

// initClient creates and initializes the RunAI API client
func initClient(fqdn string, creds Credentials, caCertPath string) (*RunAIClient, error) {
	serverURL := normalizeURL(fqdn)
	logger.Infof("Initializing RunAI client for %s", serverURL)

	if caCertPath != "" {
		logger.Infof("Using custom CA certificate from %s", caCertPath)
	}

	// Check if the control plane is responsive
	if err := CheckHealth(serverURL, caCertPath); err != nil {
		return nil, fmt.Errorf("control plane not responsive: %w", err)
	}

	if creds.Username == "" || creds.Password == "" {
		return nil, fmt.Errorf("RunAI API credentials not provided")
	}

	// Create IDM client for token management
	idmClient := newIdmClientWithCA(serverURL, caCertPath)

	// Create token source for automatic token rotation
	tokenSource := &TokenSource{
		tokenApi:  idmClient.TokensAPI,
		user:      creds.Username,
		password:  creds.Password,
		serverURL: serverURL,
	}

	// Create HTTP transport with optional custom CA
	transport, err := tls.NewTransportWithCA(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP transport: %w", err)
	}

	// Create HTTP client with automatic token injection
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
			Base:   transport,
		},
	}

	// Create Authorization API client
	authCfg := authorization.NewConfiguration()
	authCfg.Servers = authorization.ServerConfigurations{
		authorization.ServerConfiguration{URL: serverURL},
	}
	authCfg.HTTPClient = httpClient

	// Create Cluster API client
	clusterCfg := cluster.NewConfiguration()
	clusterCfg.Servers = cluster.ServerConfigurations{
		cluster.ServerConfiguration{URL: serverURL},
	}
	clusterCfg.HTTPClient = httpClient

	// Create OrgUnit API client
	orgunitCfg := orgunit.NewConfiguration()
	orgunitCfg.Servers = orgunit.ServerConfigurations{
		orgunit.ServerConfiguration{URL: serverURL},
	}
	orgunitCfg.HTTPClient = httpClient

	// Create IDM client with the same authenticated HTTP client for user operations
	idmCfg := idmclient.NewConfiguration()
	idmCfg.Servers = idmclient.ServerConfigurations{{URL: serverURL}}
	idmCfg.HTTPClient = httpClient
	authenticatedIdmClient := idmclient.NewAPIClient(idmCfg)

	client := &RunAIClient{
		authClient:    authorization.NewAPIClient(authCfg),
		clusterClient: cluster.NewAPIClient(clusterCfg),
		orgunitClient: orgunit.NewAPIClient(orgunitCfg),
		idmClient:     authenticatedIdmClient,
		fqdn:          fqdn,
	}

	logger.Infof("RunAI client initialized successfully for %s", serverURL)
	return client, nil
}

// normalizeURL ensures the URL has the https:// prefix
func normalizeURL(fqdn string) string {
	if len(fqdn) > 0 && fqdn[0] != 'h' {
		return fmt.Sprintf("https://%s", fqdn)
	}
	return fqdn
}

// newIdmClient creates a new identity manager client
func newIdmClient(baseURL string) *idmclient.APIClient {
	return newIdmClientWithCA(baseURL, "")
}

// newIdmClientWithCA creates a new identity manager client with optional CA certificate
func newIdmClientWithCA(baseURL string, caCertPath string) *idmclient.APIClient {
	config := idmclient.NewConfiguration()
	config.Servers = idmclient.ServerConfigurations{{URL: baseURL}}

	if caCertPath != "" {
		httpClient, err := tls.NewHTTPClientWithCA(caCertPath, 30*time.Second)
		if err != nil {
			logger.Errorf("Failed to create HTTP client with custom CA: %v", err)
		} else {
			config.HTTPClient = httpClient
		}
	}

	return idmclient.NewAPIClient(config)
}

// CheckHealth verifies the control plane is responsive
func CheckHealth(serverURL string, caCertPath string) error {
	logger.Infof("Checking control plane health at %s", serverURL)

	httpClient, err := tls.NewHTTPClientWithCA(caCertPath, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	versionURL := fmt.Sprintf("%s/_version", serverURL)
	resp, err := httpClient.Get(versionURL)
	if err != nil {
		return fmt.Errorf("failed to connect to control plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("control plane returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var versionResponse struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &versionResponse); err != nil {
		return fmt.Errorf("failed to parse version response: %w", err)
	}

	logger.Infof("Control plane at %s is responsive, version: %s", serverURL, versionResponse.Version)
	return nil
}

// API accessor methods

// AccessRulesAPI returns the Access Rules API client
func (c *RunAIClient) AccessRulesAPI() authorization.AccessRulesAPI {
	return c.authClient.AccessRulesAPI
}

// RolesAPI returns the Roles API client
func (c *RunAIClient) RolesAPI() authorization.RolesAPI {
	return c.authClient.RolesAPI
}

// PermissionsAPI returns the Permissions API client
func (c *RunAIClient) PermissionsAPI() authorization.PermissionsAPI {
	return c.authClient.PermissionsAPI
}

// ClustersAPI returns the Clusters API client
func (c *RunAIClient) ClustersAPI() cluster.ClustersAPI {
	return c.clusterClient.ClustersAPI
}

// NodePoolsAPI returns the Node Pools API client
func (c *RunAIClient) NodePoolsAPI() cluster.NodePoolsAPI {
	return c.clusterClient.NodePoolsAPI
}

// NodesAPI returns the Nodes API client
func (c *RunAIClient) NodesAPI() cluster.NodesAPI {
	return c.clusterClient.NodesAPI
}

// ProjectsAPI returns the Projects API client
func (c *RunAIClient) ProjectsAPI() orgunit.ProjectsAPI {
	return c.orgunitClient.ProjectsAPI
}

// DepartmentsAPI returns the Departments API client
func (c *RunAIClient) DepartmentsAPI() orgunit.DepartmentsAPI {
	return c.orgunitClient.DepartmentsAPI
}

// IdmUsersAPI returns the Users API client from the IDM service
func (c *RunAIClient) IdmUsersAPI() idmclient.UsersAPI {
	return c.idmClient.UsersAPI
}

// AuthAPIClient returns the underlying authorization API client
func (c *RunAIClient) AuthAPIClient() *authorization.APIClient {
	return c.authClient
}

// ClusterAPIClient returns the underlying cluster API client
func (c *RunAIClient) ClusterAPIClient() *cluster.APIClient {
	return c.clusterClient
}

// OrgUnitAPIClient returns the underlying org unit API client
func (c *RunAIClient) OrgUnitAPIClient() *orgunit.APIClient {
	return c.orgunitClient
}

// IdmAPIClient returns the underlying IDM API client
// Use this for accessing IDM APIs not exposed through accessor methods
func (c *RunAIClient) IdmAPIClient() *idmclient.APIClient {
	return c.idmClient
}

// FQDN returns the FQDN this client was initialized with
func (c *RunAIClient) FQDN() string {
	return c.fqdn
}

// ResetClient resets the singleton client (useful for testing)
func ResetClient() {
	clientMu.Lock()
	defer clientMu.Unlock()
	clientInstance = nil
	clientOnce = sync.Once{}
	clientFQDN = ""
}
