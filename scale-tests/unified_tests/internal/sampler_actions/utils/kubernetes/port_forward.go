package kubernetes

import (
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwarder struct {
	config *rest.Config
}

func NewPortForwarder(config *rest.Config) *PortForwarder {
	return &PortForwarder{
		config: config,
	}
}

func (pf *PortForwarder) PortForward(namespace, podName string, localPort, remotePort int, stopCh <-chan struct{}, readyCh chan struct{}, steams genericclioptions.IOStreams) error {
	url, err := url.Parse(pf.config.Host)
	if err != nil {
		return fmt.Errorf("failed to parse host %s URL: %w", pf.config.Host, err)
	}

	url.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)

	transport, upgrader, err := spdy.RoundTripperFor(pf.config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)

	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, remotePort)}, stopCh, readyCh, steams.Out, steams.ErrOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}
