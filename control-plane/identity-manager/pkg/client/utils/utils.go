package utils

const (
	ClientTypeAttr = "runai.client_type"
	ClusterIdAttr  = "runai.cluster_id"
)

func GetClusterIdAttr() string {
	return ClusterIdAttr
}

func GetClientTypeAttr() string {
	return ClientTypeAttr
}
