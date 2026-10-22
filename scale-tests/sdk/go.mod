module github.com/dsx-ai-factory/run-ai/scale-tests/sdk

go 1.27.3

replace (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ../../runai-cluster/cluster/common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils => ../../runai-cluster/cluster/common/utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk => ../../runai-cluster/cluster/sdk
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller => ../../runai-cluster/cluster/workload-controller
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cert-utils => ../../runai-cluster/common/cert-utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants => ../../runai-cluster/common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/profiling => ../../runai-cluster/common/profiling
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner => ../../runai-cluster/common/top-owner
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils => ../../runai-cluster/common/utils
)

replace k8s.io/client-go => k8s.io/client-go v0.35.4

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 => ../../runai-cluster/cluster/dromos/api/v1alpha1
