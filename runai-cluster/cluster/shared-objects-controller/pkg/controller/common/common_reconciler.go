package common

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectReconciler interface {
	// ReconcileObject - returns the object with an updated status
	ReconcileObject(ctx context.Context, object client.Object) (client.Object, error)

	UpdateStatus(ctx context.Context, object client.Object, objectWithUpdatedStatus client.Object) error

	// HandleDeletion - returns the object with an updated status, if an error occurred
	HandleDeletion(ctx context.Context, object client.Object) (client.Object, error)
}

type CommonReconciler struct {
	client           client.Client
	finalizerName    string
	objectReconciler ObjectReconciler
}

func NewCommonReconciler(client client.Client, finalizerName string, baseReconciler ObjectReconciler) CommonReconciler {
	return CommonReconciler{
		client:           client,
		finalizerName:    finalizerName,
		objectReconciler: baseReconciler,
	}
}

func (c *CommonReconciler) Reconcile(ctx context.Context, object client.Object) error {
	var err error

	objectWithUpdatedStatus, innerErr := c.objectReconciler.ReconcileObject(ctx, object)
	if innerErr != nil {
		log.Error().Msgf("Failed to reconcile %s, error: %s", ObjectLogStr(object), innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	innerErr = c.objectReconciler.UpdateStatus(ctx, object, objectWithUpdatedStatus)
	if innerErr != nil {
		log.Error().Msgf("Failed to update status of %s, error: %s", ObjectLogStr(object), innerErr.Error())
		err = multierror.Append(err, innerErr)
	}

	return err
}

func (c *CommonReconciler) AddObjects(ctx context.Context, objectsToReconcile *ObjectsToReconcile) error {
	var err error

	for _, obj := range objectsToReconcile.ToAdd {
		innerErr := c.Create(ctx, obj)
		if innerErr != nil {
			if !errors.IsConflict(innerErr) {
				err = multierror.Append(err, innerErr)
			}
		}
	}

	return err
}

func (c *CommonReconciler) UpdateObjects(ctx context.Context, objectsToReconcile *ObjectsToReconcile) error {
	var err error

	for _, obj := range objectsToReconcile.ToUpdate {
		innerErr := c.Update(ctx, obj)
		if innerErr != nil {
			if !errors.IsConflict(innerErr) {
				err = multierror.Append(err, innerErr)
			}
		}
	}

	return err
}

func (c *CommonReconciler) DeleteObjects(ctx context.Context, objectsToReconcile *ObjectsToReconcile) error {
	var err error

	for _, obj := range objectsToReconcile.ToRemove {
		innerErr := c.Delete(ctx, obj)
		if innerErr != nil {
			if !errors.IsNotFound(innerErr) {
				err = multierror.Append(err, innerErr)
			}
		}
	}

	return err
}

func (c *CommonReconciler) Create(ctx context.Context, obj client.Object) error {
	log.Info().Msgf("Creating %s", ObjectLogStrWithOwner(obj))

	err := c.client.Create(ctx, obj)
	if err != nil {
		log.Error().Msgf("Failed to create %s, error: %s", ObjectLogStrWithOwner(obj), err.Error())
		return err
	}
	return nil
}

func (c *CommonReconciler) Update(ctx context.Context, obj client.Object) error {
	log.Info().Msgf("Updating %s", ObjectLogStrWithOwner(obj))

	err := c.client.Update(ctx, obj)
	if err != nil {
		log.Error().Msgf("Failed to update %s, error: %s", ObjectLogStrWithOwner(obj), err.Error())
		return err
	}
	return nil
}

func (c *CommonReconciler) Delete(ctx context.Context, obj client.Object) error {
	log.Info().Msgf("Deleting %s", ObjectLogStrWithOwner(obj))

	err := c.client.Delete(ctx, obj)
	if err != nil {
		log.Error().Msgf("Failed to delete %s, error: %s", ObjectLogStrWithOwner(obj), err.Error())
		return err
	}
	return nil
}
