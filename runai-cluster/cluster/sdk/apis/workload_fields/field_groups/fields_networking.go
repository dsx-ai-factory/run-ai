package field_groups

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type NetworkingFields struct {

	// Specifies that the created workload will use the host's ipc namespace.
	// +k8s:conversion-gen=false
	HostIpc *flds.BoolField `json:"hostIpc,omitempty"`

	// Specifies that the created workload will use the host's network stack inside its container.
	// For more information see the Docker Run Reference at
	// https://docs.docker.com/engine/reference/run/
	HostNetwork *flds.BoolField `json:"hostNetwork,omitempty"`

	// Specifies the default service exposure method for ports.
	// The default shall be used for ports which do not specify service type.
	// Options are: LoadBalancer, NodePort or ClusterIP.
	// For more information see the External Access to Containers guide on
	// https://docs.run.ai/admin/runai-setup/config/allow-external-access-to-containers/
	ServiceType *flds.StringField `json:"serviceType,omitempty"`

	// Specifies a set of ports exposed from the container running the created workload.
	// Used together with --service-type.
	Ports *flds.PortsField `json:"ports,omitempty"`

	// Specifies a set of exported url (e.g. ingress) from the container running the created workload.
	ExposedUrls *flds.ExposedUrlsField `json:"exposedUrls,omitempty"`

	// This field is for internal use only.
	IngressUrl *flds.StringField `json:"ingressUrl,omitempty"`
}

var ServiceTypeRules = flds.StringRules{
	Options: &[]flds.OptionRule{
		{
			Value: ptr.To(constants.ServiceTypeLoadBalancer),
			Name:  ptr.To("Load Balancer"),
		},
		{
			Value: ptr.To(constants.ServiceTypeNodePort),
			Name:  ptr.To("Node Port"),
		},
		{
			Value: ptr.To(constants.ServiceTypeClusterIP),
			Name:  ptr.To("Cluster IP"),
		},
		{
			Value: ptr.To(constants.ServiceTypeServingPort),
			Name:  ptr.To("Serving Port"),
		},
	},
}

var NetworkingRunaiRules = NetworkingFields{

	ServiceType: &flds.StringField{
		Rules: &ServiceTypeRules,
	},

	Ports: &flds.PortsField{
		ItemRules: &flds.PortItemRules{
			ServiceType: &ServiceTypeRules,
			ContainerPort: &flds.IntRules{
				Format: ptr.To(workload_fields.FormatPort),
				Min:    ptr.To(0),
			},
			HostPort: &flds.IntRules{
				Format: ptr.To(workload_fields.FormatPort),
				Min:    ptr.To(0),
			},
			Protocol: &flds.StringRules{
				Required: ptr.To(false),
			},
		},
	},

	ExposedUrls: &flds.ExposedUrlsField{
		ItemRules: &flds.ExposedUrlItemRules{
			ContainerPort: &flds.IntRules{
				Format:   ptr.To(workload_fields.FormatPort),
				Min:      ptr.To(0),
				Required: ptr.To(true),
			},
			Url: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatRequestUrl),
			},
		},
	},

	IngressUrl: &flds.StringField{
		Rules: &flds.StringRules{
			Format: ptr.To(workload_fields.FormatRequestUrl),
		},
	},
}
