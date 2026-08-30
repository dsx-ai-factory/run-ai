package async_events

// PvcEvents are events that are emitted from the k8s-objects-tracker service when a PVC issues is the cluster are found or fixed.
type PvcEvents struct {
	Events []PvcEvent `json:"events"`
}

type PvcEvent struct {
	ClusterId string `json:"clusterId"`
	ScopeType string `json:"scopeType"`
	ScopeId   string `json:"scopeId"`
	ClaimName string `json:"claimName"`
	Status    string `json:"status"`
}

func NewPvcEvent(
	clusterId string,
	scopeType string,
	scopeId string,
	claimName string,
	status string,
) *PvcEvent {
	return &PvcEvent{
		ClusterId: clusterId,
		ScopeType: scopeType,
		ScopeId:   scopeId,
		ClaimName: claimName,
		Status:    status,
	}
}

func NewPvcEvents(
	events []PvcEvent,
) *PvcEvents {
	return &PvcEvents{
		Events: events,
	}
}
