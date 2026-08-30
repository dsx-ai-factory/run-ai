package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// HTTPRouteAccepted checks if any parent gateway has accepted the HTTPRoute.
func HTTPRouteAccepted(status gwv1.HTTPRouteStatus) (bool, string) {
	if len(status.Parents) == 0 {
		return false, "no gateway controller has reported status"
	}
	var rejectionMsg string
	for _, parent := range status.Parents {
		for _, cond := range parent.Conditions {
			if cond.Type == string(gwv1.RouteConditionAccepted) {
				if cond.Status == metav1.ConditionTrue {
					return true, ""
				}
				if cond.Status == metav1.ConditionFalse && rejectionMsg == "" {
					rejectionMsg = cond.Message
				}
			}
		}
	}
	if rejectionMsg != "" {
		return false, rejectionMsg
	}
	return false, "no gateway controller has accepted the HTTPRoute"
}
