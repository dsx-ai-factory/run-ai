module github.com/dsx-ai-factory/run-ai/runai-cluster/common/topology

go 1.27.2

require (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk v0.0.0-20250907093507-4e419875bd3d
	k8s.io/utils v0.0.0-20260319190234-28399d86e0b5
)

require (
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils v0.0.0-20250408104737-6a14b23817ad // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.36.2 // indirect
	k8s.io/apimachinery v0.36.2 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
)

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk => ../../sdk

replace k8s.io/client-go => k8s.io/client-go v0.35.4

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils => ./../../../common/utils
