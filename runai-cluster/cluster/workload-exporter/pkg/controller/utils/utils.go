package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RequeueTimeout = time.Second * 2

	RateLimiterBaseDelay = time.Second
	RateLimiterMaxDelay  = time.Second * 10
)

type Reconciler interface {
	SetupWithManager(mgr ctrl.Manager) error
	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

func GetObjectForRequest(ctx context.Context, req ctrl.Request, kind string,
	object client.Object, reader client.Reader) (isObjectFound bool, err error) {
	err = reader.Get(ctx, req.NamespacedName, object)
	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		log.Ctx(ctx).Warn().Msgf("Can't find %s <%s> in cluster, this might also mean it was deleted",
			kind, getNameStr(req))
		// Nothing really more to do.
		return false, nil
	}

	log.Ctx(ctx).Error().Msgf("Failed to get %s <%s> from incoming request, error: %s",
		kind, getNameStr(req), err.Error())
	return false, err
}

func getNameStr(req ctrl.Request) string {
	if req.Namespace == "" {
		return req.Name
	}

	return fmt.Sprintf("%s/%s", req.Namespace, req.Name)
}
