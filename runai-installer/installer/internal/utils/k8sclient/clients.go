package k8sclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const k8sConnectionTimeout = 5 * time.Second

var once sync.Once
var instance *k8sClientImpl
var initError error

type K8sClients interface {
	GetClient() client.Client
	GetConfig() *rest.Config
	GetContextName() string
}

// k8sClientImpl is the real implementation of K8sClient.
type k8sClientImpl struct {
	client.Client
	cfg         *rest.Config
	contextName string
}

// GetK8sClients creates a new real K8sClient implementation using default
// kubeconfig discovery.
func GetK8sClients() (K8sClients, error) {
	return GetK8sClientsWithKubeconfig("")
}

// GetK8sClientsWithKubeconfig creates the real K8sClient implementation. When
// kubeconfigPath is non-empty the REST config is built from that explicit path;
// otherwise default discovery is used.
func GetK8sClientsWithKubeconfig(kubeconfigPath string) (K8sClients, error) {
	once.Do(func() {
		logging.Infof("[K8s] Initializing Kubernetes client...")
		cfg, err := loadRESTConfig(kubeconfigPath)
		if err != nil {
			initError = fmt.Errorf("failed to get kubeconfig: %w", err)
			logging.Errorf("[K8s] Failed to get kubeconfig: %v", err)
			return
		}
		// Get the current context name from kubeconfig
		contextName := getCurrentContextName(kubeconfigPath)
		logging.Infof("[K8s] Kubeconfig loaded successfully (context: %s, server: %s)", contextName, cfg.Host)

		// Set shorter timeouts for faster failure when k8s is unavailable
		cfg.Timeout = k8sConnectionTimeout

		// Route Kubernetes API HTTP warning headers (299 status) through our logging system
		// Note: klog output (Info/Error logs from client-go) is configured separately in main.go
		// This prevents warnings from leaking to stderr and breaking the TUI
		cfg.WarningHandler = rest.NewWarningWriter(logging.InfoWriter(), rest.WarningWriterOptions{
			Deduplicate: true, // Avoid duplicate warnings
		})

		scheme := runtime.NewScheme()
		if err := appsv1.AddToScheme(scheme); err != nil {
			initError = fmt.Errorf("failed to add appsv1 to scheme: %w", err)
			logging.Errorf("[K8s] Failed to add appsv1 to scheme: %v", err)
			return
		}
		if err := corev1.AddToScheme(scheme); err != nil {
			initError = fmt.Errorf("failed to add corev1 to scheme: %w", err)
			logging.Errorf("[K8s] Failed to add corev1 to scheme: %v", err)
			return
		}
		k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
		if err != nil {
			initError = fmt.Errorf("failed to create client: %w", err)
			logging.Errorf("[K8s] Failed to create Kubernetes client: %v", err)
			return
		}

		logging.Infof("[K8s] Kubernetes client initialized for context: %s", contextName)

		instance = &k8sClientImpl{
			Client:      k8sClient,
			cfg:         cfg,
			contextName: contextName,
		}
	})
	return instance, initError
}

// ValidateK8sConnection validates that we can connect to the k8s cluster
func ValidateK8sConnection(k8sClient K8sClients) error {
	logging.Infof("[K8s] Validating connection to Kubernetes cluster by listing namespaces...")
	ctx, cancel := context.WithTimeout(context.Background(), k8sConnectionTimeout)
	defer cancel()

	// Try to list namespaces as a simple connectivity check
	namespaceList := &corev1.NamespaceList{}
	err := k8sClient.GetClient().List(ctx, namespaceList, &client.ListOptions{Limit: 1})
	if err != nil {
		logging.Errorf("[K8s] Connection validation failed: %v", err)
		return fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}
	logging.Infof("[K8s] Connection to Kubernetes cluster validated successfully")
	return nil
}

func loadRESTConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return config.GetConfig()
}

// getCurrentContextName gets the current context name from kubeconfig
func getCurrentContextName(kubeconfigPath string) string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		loadingRules.ExplicitPath = kubeconfigPath
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return "unknown"
	}
	return rawConfig.CurrentContext
}

func (k *k8sClientImpl) GetClient() client.Client {
	return k.Client
}

func (k *k8sClientImpl) GetConfig() *rest.Config {
	return k.cfg
}

func (k *k8sClientImpl) GetContextName() string {
	return k.contextName
}
