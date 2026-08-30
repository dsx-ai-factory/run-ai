package client

import (
	"context"
	"fmt"
	"io"

	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	QPS   = 1000
	Burst = 1000
)

type Client struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
}

func NewClient() (*Client, error) {
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := configLoader.ClientConfig()
	if err != nil {
		return nil, err
	}

	restConfig.QPS = QPS
	restConfig.Burst = Burst

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientset:  clientset,
		restConfig: restConfig,
	}, nil
}

func (c *Client) ListPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (c *Client) ListNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (c *Client) GetPodLogStream(ctx context.Context, podName string, namespace string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	return req.Stream(ctx)
}

func (c *Client) VerifyAccess(ctx context.Context, sar *authorizationv1.SelfSubjectAccessReview) (bool, error) {
	result, err := c.clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to verify access permissions: %w", err)
	}
	return result.Status.Allowed, nil
}

func (c *Client) CheckConnectivity(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to check connectivity: %w", err)
	}
	return nil
}

func (c *Client) GetServerUrl(ctx context.Context) string {
	return c.restConfig.Host
}
