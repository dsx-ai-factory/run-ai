package kubeConfig

import (
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InitWatchClient(scheme *runtime.Scheme) (client.WithWatch, error) {
	watchClient, err := client.NewWithWatch(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		zap.S().Errorf("failed creating watchClient: %+v", err)
		return nil, err
	}
	zap.S().Debug("the k8's watchClient were created successfully ")
	return watchClient, nil
}
