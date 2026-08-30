package validator

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TODO delete when merged (Function exist in PR)
func CreateDumbServiceWithLabel(name, namespace string, ownerUID types.UID, ownerKind string, labels map[string]string) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	if ownerUID != "" && ownerKind != "" {
		service.OwnerReferences = []metav1.OwnerReference{
			{
				UID:  ownerUID,
				Kind: ownerKind,
			},
		}
	}
	return service
}
