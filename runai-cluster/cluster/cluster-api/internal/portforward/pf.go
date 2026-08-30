package portforward

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	ResourcePods        = "pods"
	ResourcePortforward = "portforward"
	MethodPost          = "POST"
)

type PodForwarder struct {
	rest *restclient.Config
	core kubernetes.Interface
}

func NewPodForwarder(rest *restclient.Config, core kubernetes.Interface, proxy ProxyFactoryFn) *PodForwarder {
	log.Debug().Msg("Creating new PodForwarder")
	return &PodForwarder{rest: rest, core: core}
}

func (pf *PodForwarder) Create(namespace string, options *Options) (httpstream.Connection, error) {
	log.Info().Str("namespace", namespace).Int32("port", options.Port).Str("pod", options.Pod).Msg("Creating port forward")

	roundTripper, upgrader, err := spdy.RoundTripperFor(pf.rest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create round tripper")
		return nil, err
	}

	req := pf.core.CoreV1().RESTClient().Post().
		Resource(ResourcePods).
		Namespace(namespace).
		Name(options.Pod).
		SubResource(ResourcePortforward)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, MethodPost, req.URL())

	dial, _, err := dialer.Dial(portforward.PortForwardProtocolV1Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to dial")
		return nil, err
	}

	log.Info().Msg("Port forward created successfully")

	return dial, nil
}
