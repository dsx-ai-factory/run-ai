package async_events

type ProjectDeleteRequest struct {
	ClusterId   string `json:"cluster_id"`
	ProjectId   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	// A flag that indicates if the receiver of the event should delete cluster sources objects, or leave them orphans behind
	LeaveOrphans bool `json:"leave_orphans"`
}

func NewProjectDeleteRequest(
	clusterId string,
	projectId string,
	projectName string,
	leaveOrphans bool,
) *ProjectDeleteRequest {
	return &ProjectDeleteRequest{
		ClusterId:    clusterId,
		ProjectId:    projectId,
		ProjectName:  projectName,
		LeaveOrphans: leaveOrphans,
	}
}
