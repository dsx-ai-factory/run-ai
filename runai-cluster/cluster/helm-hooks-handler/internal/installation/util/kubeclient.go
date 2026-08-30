package util

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOrUpdate(ctx context.Context, c client.Client, obj client.Object) error {
	if err := c.Create(ctx, obj); !apierrors.IsAlreadyExists(err) {
		return err
	}
	return c.Update(ctx, obj)
}
