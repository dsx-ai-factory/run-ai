/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg = "submit {{.AOrAn}} {{.Type}} workload"
	FailMsg  = "{{.AOrAn}} {{.Type}} submission failed"

	DistributedInferenceLongMsg = `Before using the flags, keep in mind:

By default (--worker-as-leader=true), worker pods inherit their configuration from the leader:
- You only need to specify worker-specific flags if you want worker pods to differ from the leader
- When --worker-as-leader=false is set, workers require their own explicit --worker-* flags

Command and Arguments:
- Use --command to set the leader's entrypoint command
- Use --arguments OR positional args (-- arg1 arg2) to set the leader's arguments
- Note: If both --arguments and -- are used, positional args (--) take precedence
- Use --worker-command and --worker-arguments to override worker's command/args`

	BaseExampleMsg = `# Submit a workload
runai {{.CommandType}} submit <workload-name> -p <project-name> -i runai.jfrog.io/demo/quickstart-demo

# Submit a workload with arguments
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ubuntu -- ls -la

# Submit a workload with a template
runai {{.CommandType}} submit <workload-name> -p <project-name> --template <template-name>

# Submit a workload with an environment asset
runai {{.CommandType}} submit <workload-name> -p <project-name> --environment <environment-asset-name>

# Submit a workload with a custom command
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ubuntu --command -- echo "Hello, World"

# Submit a workload with a field reference 
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ubuntu --env-pod-field-ref "PROJECT=metadata.labels['project']"`

	JupyterExampleMsg = `

# Submit the workload jupyter notebook
runai {{.CommandType}} submit <workload-name> -p <project-name> -i jupyter/scipy-notebook --gpu-devices-request 1 --external-url container=8888 --name-prefix jupyter --command -- start-notebook.sh --NotebookApp.base_url='/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}' --NotebookApp.token='` //nolint:lll // example no way to shorten it

	InferenceExampleMsg = `
# Submit a workload with scale to zero
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ghcr.io/knative/helloworld-go --gpu-devices-request 1 
--serving-port=8080 --min-replicas=0 --max-replicas=1 

# Submit a workload with arguments
runai {{.CommandType}} submit <workload-name> -p <project-name> -i vllm/vllm-openai:latest --gpu-devices-request 1 --serving-port=8000 -- --model Qwen/Qwen3-0.6B

# Submit a workload with a template
runai {{.CommandType}} submit <workload-name> -p <project-name> --template <template-name>

# Submit a workload with an asset
runai {{.CommandType}} submit <workload-name> -p <project-name> --environment <environment-asset-name>

# Submit a workload with autoscaling and authorization
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ghcr.io/knative/helloworld-go  --gpu-devices-request 1
--serving-port=container=8080,authorization-type=authorizedUsersOrGroups,authorized-users=user1:user2:app1,protocol=http 
--min-replicas=1 --max-replicas=4 --metric=concurrency  --metric-threshold=100 ` //nolint:lll // example no way to shorten it

	// DistributedInferenceExampleMsg intentionally omits the environment asset example
	// because environment assets are not yet supported for distributed inference workloads.
	DistributedInferenceExampleMsg = `
# Submit a workload
runai {{.CommandType}} submit <workload-name> -p <project-name> -i tiangolo/uvicorn-gunicorn-fastapi:python3.9 --replicas 2 --workers 3 -g 1 --serving-port 80

# Submit a workload with a template
runai {{.CommandType}} submit <workload-name> -p <project-name> --template <template-name>

# Submit a workload where workers inherit leader settings but override specific fields
runai {{.CommandType}} submit <workload-name> -p <project-name> -i tiangolo/uvicorn-gunicorn-fastapi:python3.9 --environment ROLE=LEADER --replicas 3 --workers 7 -g 2 --serving-port 80 --worker-environment ROLE=WORKER

# Submit a workload with a different worker setup (workers not based on leader configuration)
runai {{.CommandType}} submit <workload-name> -p <project-name> -i tiangolo/uvicorn-gunicorn-fastapi:python3.9 --replicas 4 --workers 1 -g 8 --serving-port 80 --existing-pvc claimname=my-pvc,path=/data --worker-as-leader=false --worker-image tiangolo/uvicorn-gunicorn-fastapi:python3.9 --worker-gpu-devices-request 8` //nolint:lll // example no way to shorten it

	// NimInferenceLongMsg provides guidance for NIM submit command
	NimInferenceLongMsg = `Before using the flags, keep in mind:

Model Store Configuration:
- You must select exactly one model store option: --model-existing-pvc, --model-new-pvc, or --model-nim-cache.

Autoscaling vs Fixed Replicas:
- Use --replicas for a fixed number of instances.
- Use --min-replicas, --max-replicas, --metric for autoscaling.

Multi-Node Inference:
- Use --workers to specify the number of worker nodes per leader.
- Setting workers > 0 automatically enables multi-node mode (1 Leader + N Workers).
- Note: Multi-node (--workers) cannot be used with autoscaling (--min-replicas/--max-replicas).`

	//nolint:lll // example no way to shorten it
	NimInferenceExampleMsg = `
# Submit a Llama3 model using an existing PVC
runai {{.CommandType}} submit <workload-name> -p <project-name> -i nvcr.io/nim/meta/llama3-8b-instruct:latest --ngc-auth-secret my-ngc-secret --model-existing-pvc claimname=my-model-pvc --serving-port 8000 -g 1

# Submit with autoscaling
runai {{.CommandType}} submit <workload-name> -p <project-name> -i nvcr.io/nim/meta/llama3-8b-instruct:latest --ngc-auth-secret my-ngc-secret --model-nim-cache "name=my-cache,profile=10gb" --min-replicas 1 --max-replicas 5 --metric concurrency --metric-threshold 10 --serving-port 8000

# Submit multi-node model (1 Leader + 2 Workers)
runai {{.CommandType}} submit <workload-name> -p <project-name> -i nvcr.io/nim/meta/llama3-8b-instruct:latest --ngc-auth-secret my-ngc-secret --model-new-pvc "claimname=my-model-pvc,size=100Gi,accessmode-rwo" --workers 2 --replicas 1 --serving-port 8000` //nolint:lll // example no way to shorten it

	MasterArgsAndCommandExampleMsg = `

# Submit a workload master args with worker args
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ubuntu --master-args "-a master_arg_a -b master-arg_b'" -- '-a worker_arg_a'

# Submit a workload master command with worker args
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ubuntu --master-command "echo -e 'master command'" -- '-a worker_arg_a'

# Submit a workload master command with worker command
runai {{.CommandType}} submit <workload-name> -p <project-name> -i ubuntu --master-command "echo -e 'master command'" --command -- echo -e 'worker command'`
)

var (
	baseWithJupyterExample              = BaseExampleMsg + JupyterExampleMsg
	baseWIthMasterArgsAndCommandExample = BaseExampleMsg + MasterArgsAndCommandExampleMsg
)

func NewFrameworkText(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	if workloadType != api.Distributed {
		return cmd.NewBaseCommandTexts(ShortMsg, "", baseWithJupyterExample, FailMsg, "", nil, replacementTokens)
	} else if framework != api.Jax {
		return cmd.NewBaseCommandTexts(ShortMsg, "", baseWIthMasterArgsAndCommandExample, FailMsg, "", nil, replacementTokens)
	}
	return cmd.NewBaseCommandTexts(ShortMsg, "", BaseExampleMsg, FailMsg, "", nil, replacementTokens)
}

func NewNimSubmitText() *cmd.BaseCommandTexts {
	return cmd.NewBaseCommandTexts(ShortMsg, NimInferenceLongMsg, NimInferenceExampleMsg, FailMsg, "", nil, cmd.GetNimReplacementTokens())
}

func NewWorkloadOnlyText(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	var replacementTokens []cmd.TokenPair

	// For inference workloads, use inference replacement tokens to get correct path
	if workloadType == api.Inference || workloadType == api.DistributedInference {
		replacementTokens = cmd.GetInferenceReplacementTokens(workloadType)
	} else {
		replacementTokens = cmd.GetWorkloadOnlyReplacementTokens(workloadType)
	}

	var example string
	var longMsg string

	switch workloadType {
	case api.Inference:
		example = InferenceExampleMsg
	case api.DistributedInference:
		example = DistributedInferenceExampleMsg
		longMsg = DistributedInferenceLongMsg
	case api.Distributed:
		example = BaseExampleMsg
	default:
		example = BaseExampleMsg + JupyterExampleMsg
	}

	return cmd.NewBaseCommandTexts(ShortMsg, longMsg, example, FailMsg, "", nil, replacementTokens)
}
