package utils

// ContextKey is a dedicated type for context.WithValue keys, so that values
// stored by this service cannot collide with keys set by other packages.
type ContextKey string

// context keys
const (
	FullHostKey ContextKey = "fullHost"
)

const (
	Authorization = "Authorization"
)

// aws cluster installastion query params
const (
	StackNameParam    = "stackName"
	EksClusterParam   = "param_EKSClusterName"
	TenantNameParam   = "param_RunaiTenantName"
	ClientIdParam     = "param_RunaiClientID"
	ClientSecretParam = "param_RunaiClientSecret"
	ClusterIdParam    = "param_RunaiClusterID"
	CpUrlParam        = "param_RunaiControlPlaneURL"
	ClusterUrlParam   = "param_RunaiClusterURL"
)

const (
	ClusterUrlNamePrefix = "runai-cluster"
)
