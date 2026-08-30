package utils

import (
	"context"

	"github.com/rs/zerolog/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsCRDInstalled(ctx context.Context, reader client.Reader, crdName string) bool {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := reader.Get(ctx, client.ObjectKey{Name: crdName}, crd); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error().Msgf("Failed to check CRD <%s>, assuming not installed, err: %v", crdName, err.Error())
		}
		return false
	}
	for _, cond := range crd.Status.Conditions {
		if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
			return true
		}
	}
	return false
}
