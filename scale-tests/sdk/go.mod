module github.com/run-ai/runai/scale-tests/sdk

go 1.26.3

replace (
	github.com/run-ai/runai/runai-cluster/cluster/common/constants => ../../runai-cluster/cluster/common/constants
	github.com/run-ai/runai/runai-cluster/cluster/common/utils => ../../runai-cluster/cluster/common/utils
	github.com/run-ai/runai/runai-cluster/cluster/go-operator => ../../runai-cluster/cluster/go-operator
	github.com/run-ai/runai/runai-cluster/cluster/sdk => ../../runai-cluster/cluster/sdk
	github.com/run-ai/runai/runai-cluster/cluster/workload-controller => ../../runai-cluster/cluster/workload-controller
	github.com/run-ai/runai/runai-cluster/common/cert-utils => ../../runai-cluster/common/cert-utils
	github.com/run-ai/runai/runai-cluster/common/constants => ../../runai-cluster/common/constants
	github.com/run-ai/runai/runai-cluster/common/profiling => ../../runai-cluster/common/profiling
	github.com/run-ai/runai/runai-cluster/common/top-owner => ../../runai-cluster/common/top-owner
	github.com/run-ai/runai/runai-cluster/common/utils => ../../runai-cluster/common/utils
)

replace k8s.io/client-go => k8s.io/client-go v0.35.4

replace github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1 => ../../runai-cluster/cluster/dromos/api/v1alpha1
