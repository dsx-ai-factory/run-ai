package util

import (
	"context"
	"fmt"

	ocpconf "github.com/openshift/api/config/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdate(ctx context.Context, c client.Client, obj client.Object) error {
	if err := c.Create(ctx, obj); !apierrors.IsAlreadyExists(err) {
		return err
	}
	return c.Update(ctx, obj)
}

func IsOpenshift(ctx context.Context, c client.Reader) (bool, error) {
	cv := &ocpconf.ClusterVersionList{}
	err := c.List(ctx, cv, &client.ListOptions{})

	if meta.IsNoMatchError(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to determine if running on openshift or not: %w", err)
	}

	if len(cv.Items) == 0 {
		return false, nil
	}

	return true, nil
}
