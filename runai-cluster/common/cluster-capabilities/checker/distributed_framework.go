package checker

import (
	"fmt"
	workloadsService "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	MPIDependency        = "mpi"
	PyTorchDependency    = "pytorch"
	TensorflowDependency = "tensorflow"
	XGBoostDependency    = "xgboost"
	JaxDependency        = "jax"

	DistributedSuspendComponent = "suspend"
)

var DistributedFrameworksDependencies = []string{MPIDependency, PyTorchDependency, TensorflowDependency, XGBoostDependency, JaxDependency}

func getDistributedFrameworkKey(frameworkDependency string) (workloadsService.DistributedFramework, error) {
	switch frameworkDependency {
	case MPIDependency:
		return workloadsService.DISTRIBUTEDFRAMEWORK_MPI, nil
	case PyTorchDependency:
		return workloadsService.DISTRIBUTEDFRAMEWORK_PY_TORCH, nil
	case TensorflowDependency:
		return workloadsService.DISTRIBUTEDFRAMEWORK_TF, nil
	case XGBoostDependency:
		return workloadsService.DISTRIBUTEDFRAMEWORK_XG_BOOST, nil
	case JaxDependency:
		return workloadsService.DISTRIBUTEDFRAMEWORK_JAX, nil
	default:
		return "", fmt.Errorf("unknown distributed framework dependency %s", frameworkDependency)
	}
}
