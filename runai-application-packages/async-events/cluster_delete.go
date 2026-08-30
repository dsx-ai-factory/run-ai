package async_events

type ClusterDelete struct {
	ClusterId string `json:"cluster_id"`
}

func NewClusterDelete(clusterId string) *ClusterDelete {
	return &ClusterDelete{
		ClusterId: clusterId,
	}
}
