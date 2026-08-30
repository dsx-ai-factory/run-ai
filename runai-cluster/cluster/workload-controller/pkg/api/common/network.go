package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NetworkStatus represents the network status of a workload
// +k8s:deepcopy-gen=true
type NetworkStatus struct {
	// ServicesStatus The list of services that are exposed by this workload.
	// +optional
	// patchMergeKey=name
	// patchStrategy=merge
	// +listType=map
	// +listMapKey=name
	ServiceStatuses []ServiceStatus `json:"services,omitempty" protobuf:"bytes,1,opt,name=services"`
}

// ServiceStatus represents the status of a service exposed by the workload
// +k8s:deepcopy-gen=true
type ServiceStatus struct {
	// Name The name of the service.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Type The type of the service.
	Type corev1.ServiceType `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
	// Ports The list of ports that are exposed by this service.
	// +optional
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	Ports []ExposePort `json:"ports,omitempty" protobuf:"bytes,3,opt,name=ports"`
	// LoadBalancer contains the current status of the load-balancer, if one is present.
	// +optional
	LoadBalancer *LoadBalancerStatus `json:"loadBalancer,omitempty" protobuf:"bytes,4,opt,name=loadBalancer"`
}

// ExposePort represents a port exposed by the workload
// +k8s:deepcopy-gen=true
type ExposePort struct {
	// The name of this port within the service. This must be a DNS_LABEL.
	// All ports within a ServiceSpec must have unique names. When considering
	// the endpoints for a Service, this must match the 'name' field in the
	// EndpointPort.
	// Optional if only one ExposePort is defined on this service.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// TargetPort Number or name of the port to access on the pods targeted by the service.
	// Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
	// If this is a string, it will be looked up as a named port in the
	// target Pod's container ports. If this is not specified, the value
	// of the 'port' field is used (an identity map).
	// This field is ignored for services with clusterIP=None, and should be
	// omitted or set equal to the 'port' field.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service
	// +optional
	TargetPort intstr.IntOrString `json:"targetPort,omitempty" protobuf:"bytes,2,opt,name=targetPort"`
	// Port The port that will be exposed by this service.
	// kubebuilder:validation:Required
	Port int32 `json:"port" protobuf:"varint,3,opt,name=port"`
	// Protocol The IP protocol for this port. Supports "TCP", "UDP", and "SCTP".
	// Default is TCP.
	// +default="TCP"
	// +optional
	// +kubebuilder:validation:Enum:=TCP;UDP;SCTP
	Protocol corev1.Protocol `json:"protocol,omitempty" protobuf:"bytes,4,opt,name=protocol"`
}

// LoadBalancerStatus represents the status of a load-balancer.
// +k8s:deepcopy-gen=true
type LoadBalancerStatus struct {
	// Ingress is a list containing ingress points for the load-balancer.
	// Traffic intended for the service should be sent to these ingress points.
	// +optional
	// +listType=atomic
	Ingress []LoadBalancerIngress `json:"ingress,omitempty" protobuf:"bytes,1,rep,name=ingress"`
}

// LoadBalancerIngress represents the status of a load-balancer ingress point:
// traffic intended for the service should be sent to an ingress point.
// +k8s:deepcopy-gen=true
type LoadBalancerIngress struct {
	// IP is set for load-balancer ingress points that are IP based
	// (typically GCE or OpenStack load-balancers)
	// +optional
	IP string `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`

	// Hostname is set for load-balancer ingress points that are DNS based
	// (typically AWS load-balancers)
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,2,opt,name=hostname"`
}
