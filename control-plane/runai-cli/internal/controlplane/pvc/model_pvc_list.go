package pvc

import (
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

var (
	ManagedTrue  = true
	ManagedFalse = false
)

type Brief struct {
	// name of the pvc.
	Name    string `json:"name"`
	Managed *bool  `json:"managed"`
}

func newBriefFromPvc(info openapi.PVCListInfo) Brief {
	if !info.Origin.IsSet() {
		return Brief{Name: info.Name, Managed: nil}
	}
	if info.GetOrigin() == openapi.ORIGIN_ASSETS {
		return Brief{info.Name, &ManagedTrue}
	}
	return Brief{info.Name, &ManagedFalse}
}

func BriefListFromResponse(pvcList *openapi.PVCListResponse) ([]Brief, error) {
	if pvcList == nil {
		return nil, fmt.Errorf("list attribute is missing")
	}
	var briefList []Brief

	for _, pvc := range pvcList.Pvcs {
		briefList = append(briefList, newBriefFromPvc(pvc))
	}

	return briefList, nil
}
