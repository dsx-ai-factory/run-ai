package utils

import (
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"k8s.io/utils/ptr"
)

func GetBasicRealmArray() []*models.Realm {
	realmArray := []*models.Realm{
		{
			Enabled: ptr.To(true),
			ID:      ptr.To("123"),
			Name:    ptr.To(REALM_NAME),
		},
	}
	return realmArray
}
