package attach

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	ResourcePods      = "pods"
	SubResourceAttach = "attach"
)

type PodAttach struct {
	client kubernetes.Interface
}

func NewPodAttach(client kubernetes.Interface) *PodAttach {
	return &PodAttach{client: client}
}

func (p *PodAttach) CreateRequest(namespace string, options *stream.AttachOptions) (*rest.Request, error) {
	attachOptions, err := options.ToK8sPodAttachOptions()
	if err != nil {
		return nil, err
	}

	req := p.client.CoreV1().RESTClient().Post().
		Resource(ResourcePods).
		Namespace(namespace).
		Name(options.Pod).
		SubResource(SubResourceAttach)
	req.VersionedParams(
		attachOptions,
		scheme.ParameterCodec,
	)

	return req, nil
}
