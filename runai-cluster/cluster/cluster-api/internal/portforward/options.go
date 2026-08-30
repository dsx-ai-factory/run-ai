package portforward

import (
	"fmt"
	"net/http"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

type Options struct {
	Pod  string
	Port int32
}

type OptionsBuilder struct {
	options Options
}

func (o Options) ToK8sPfOptions() (*corev1.PodPortForwardOptions, error) {
	pfo := corev1.PodPortForwardOptions{
		Ports: nil,
	}

	if o.Port == 0 {
		return nil, fmt.Errorf("port cannot be 0")
	}

	pfo.Ports = append(pfo.Ports, o.Port)

	return &pfo, nil
}

func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{}
}

func (b *OptionsBuilder) WithPod(pod string) *OptionsBuilder {
	b.options.Pod = pod
	return b
}

func (b *OptionsBuilder) WithPort(port int32) *OptionsBuilder {
	b.options.Port = port
	return b
}

func (b *OptionsBuilder) Build() Options {
	return b.options
}

func (b *OptionsBuilder) FromOptions(options Options) *OptionsBuilder {
	b.options = options
	return b
}

func FromGetRequest(r *http.Request) (Options, error) {
	builder := NewOptionsBuilder()

	pod := r.URL.Query().Get("pod")
	if pod != "" {
		builder.WithPod(pod)
	}

	portStr := r.URL.Query().Get("port")
	if portStr == "" {
		return Options{}, fmt.Errorf("port query is missing")
	}

	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return Options{}, fmt.Errorf("invalid port value: %v", err)
	}

	builder.WithPort(int32(port))

	return builder.Build(), nil
}
