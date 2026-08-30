package orchestrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testClusterName = "test-cluster"
	testFQDN        = "google.com" //reachable domain for tests
)

// setupMockK8sClients creates a MockK8sClients with a fake client for testing
func setupMockK8sClients(t *testing.T) *mocks.MockK8sClients {
	fakeClient := fake.NewClientBuilder().Build()
	mockK8s := mocks.NewMockK8sClients(t)
	mockK8s.On("GetClient").Return(fakeClient).Maybe()
	mockK8s.On("GetConfig").Return(&rest.Config{}).Maybe()
	mockK8s.On("GetContextName").Return("test-context").Maybe()
	return mockK8s
}

// setupTestCerts creates temporary SSL certificate files for testing
func setupTestCerts(t *testing.T) (certPath, keyPath string, cleanup func()) {
	tmpDir, err := os.MkdirTemp("", "installer-test-*")
	require.NoError(t, err)

	certPath = filepath.Join(tmpDir, "cert.pem")
	keyPath = filepath.Join(tmpDir, "key.pem")

	// Create dummy cert and key files
	err = os.WriteFile(certPath, []byte("dummy cert"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(keyPath, []byte("dummy key"), 0600)
	require.NoError(t, err)

	cleanup = func() {
		os.RemoveAll(tmpDir)
	}

	return certPath, keyPath, cleanup
}

func TestInit(t *testing.T) {
	// Reset global state before test
	Reset()

	// Setup temporary SSL certificate files
	certPath, keyPath, cleanup := setupTestCerts(t)
	defer cleanup()

	cfg := &config.InstallConfig{
		Mode:    config.ModeCLI,
		GPUType: config.GPUTypeReal,
		Env:     config.EnvProd,
		DevMode: true, // Enable dev mode to skip some validations
	}
	runaiData, err := model.NewRunAIInstallData(cfg)
	require.NoError(t, err)

	// Add required mandatory inputs
	runaiData.Inputs[configinputs.InstallInputClusterName] = testClusterName
	runaiData.Inputs[configinputs.InstallInputFQDN] = testFQDN // Use a reachable domain for tests
	runaiData.Inputs[configinputs.InstallInputSSLPath] = certPath
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = keyPath

	// Create mock K8s client for testing with proper expectations
	mockK8s := setupMockK8sClients(t)

	// Test InitWithK8sClient
	err = InitWithK8sClient(runaiData, mockK8s)
	assert.NoError(t, err)
	assert.True(t, IsInitialized())

	// Test Get
	inst := Get()
	assert.NotNil(t, inst)
	assert.Equal(t, mockK8s, inst.k8sClient)

	// Test that calling Init again doesn't change the instance (sync.Once behavior)
	Reset()
	mockK8s2 := setupMockK8sClients(t)
	err = InitWithK8sClient(runaiData, mockK8s2)
	assert.NoError(t, err)
	inst2 := Get()
	assert.NotNil(t, inst2)
}

func TestInitWithNilData(t *testing.T) {
	// Reset global state before test
	Reset()
	mockK8s := setupMockK8sClients(t)
	// Test InitWithK8sClient with nil data
	err := InitWithK8sClient(nil, mockK8s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runaiData cannot be nil")
}

func TestInitWithNilK8sClient(t *testing.T) {
	// Reset global state before test
	Reset()
	// Setup temporary SSL certificate files
	certPath, keyPath, cleanup := setupTestCerts(t)
	defer cleanup()

	cfg := &config.InstallConfig{
		Mode:    config.ModeCLI,
		GPUType: config.GPUTypeReal,
		Env:     config.EnvProd,
		DevMode: true,
	}
	runaiData, err := model.NewRunAIInstallData(cfg)
	require.NoError(t, err)

	// Add required mandatory inputs
	runaiData.Inputs[configinputs.InstallInputClusterName] = testClusterName
	runaiData.Inputs[configinputs.InstallInputFQDN] = testFQDN // Use a reachable domain for tests
	runaiData.Inputs[configinputs.InstallInputSSLPath] = certPath
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = keyPath

	// Test InitWithK8sClient with nil K8s client
	err = InitWithK8sClient(runaiData, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "k8sClients cannot be nil")
}

func TestGetBeforeInit(t *testing.T) {
	// Reset global state before test
	Reset()
	// Test that Get panics when called before Init
	assert.Panics(t, func() {
		Get()
	})
}

func TestIsInitialized(t *testing.T) {
	// Reset global state before test
	Reset()
	// Test IsInitialized before init:ns
	assert.False(t, IsInitialized())

	// Setup temporary SSL certificate files
	certPath, keyPath, cleanup := setupTestCerts(t)
	defer cleanup()

	cfg := &config.InstallConfig{
		Mode:    config.ModeCLI,
		GPUType: config.GPUTypeReal,
		Env:     config.EnvProd,
		DevMode: true,
	}
	runaiData, err := model.NewRunAIInstallData(cfg)
	require.NoError(t, err)

	// Add required mandatory inputs
	runaiData.Inputs[configinputs.InstallInputClusterName] = testClusterName
	runaiData.Inputs[configinputs.InstallInputFQDN] = testFQDN // Use a reachable domain for tests
	runaiData.Inputs[configinputs.InstallInputSSLPath] = certPath
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = keyPath

	mockK8s := setupMockK8sClients(t)
	// Initialize
	err = InitWithK8sClient(runaiData, mockK8s)
	require.NoError(t, err)

	// Test IsInitialized after init
	assert.True(t, IsInitialized())
}

func TestReset(t *testing.T) {
	// Setup temporary SSL certificate files
	certPath, keyPath, cleanup := setupTestCerts(t)
	defer cleanup()

	cfg := &config.InstallConfig{
		Mode:    config.ModeCLI,
		GPUType: config.GPUTypeReal,
		Env:     config.EnvProd,
		DevMode: true,
	}
	runaiData, err := model.NewRunAIInstallData(cfg)
	require.NoError(t, err)

	// Add required mandatory inputs
	runaiData.Inputs[configinputs.InstallInputClusterName] = testClusterName
	runaiData.Inputs[configinputs.InstallInputFQDN] = testFQDN // Use a reachable domain for tests
	runaiData.Inputs[configinputs.InstallInputSSLPath] = certPath
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = keyPath

	mockK8s := setupMockK8sClients(t)
	// Initialize
	err = InitWithK8sClient(runaiData, mockK8s)
	require.NoError(t, err)
	assert.True(t, IsInitialized())

	// Reset
	Reset()
	assert.False(t, IsInitialized())

	// Test that Get panics after reset
	assert.Panics(t, func() {
		Get()
	})

	// Create a new mock for the second initialization
	mockK8s2 := setupMockK8sClients(t)
	// Test that we can initialize again after reset
	err = InitWithK8sClient(runaiData, mockK8s2)
	assert.NoError(t, err)
	assert.True(t, IsInitialized())
}
