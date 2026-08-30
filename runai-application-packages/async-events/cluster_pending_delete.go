package async_events

type ClusterPendingDelete struct {
	ClusterId string `json:"cluster_id"`
}

func NewClusterPendingDelete(clusterId string) *ClusterPendingDelete {
	return &ClusterPendingDelete{
		ClusterId: clusterId,
	}
}
