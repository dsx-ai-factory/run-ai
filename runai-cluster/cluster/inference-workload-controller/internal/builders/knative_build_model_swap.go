package builders

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// this module is responsible for the model-swap feature, which builds a knative service that deploys multiple models
// on the same GPU resource, using the GPU swap Run:ai core technology.
// the feature is triggered by a special annotation on the workload, as it's an experimental feature and is not yet
// available as 1st class citizen via CRD and API.
//
// given N models, the builder creates N+1 containers: one for an LLM gateway and one LLM server for each of the N models.
// the gateway is the container that exposes the serving port and is the access point to the different servers (models).
// the LLM server can be of the following types: vllm, tgi and nim.
// for each of these types, a container spec is built according to the given workload spec as base spec + adjustments of
// the spec for the specific model that container should serve.
// for example, if using vllm, all the containers will use the same image, command, base args, storage, security, etc.
// but will vary on the extra args of --model and --port. each model will use different values for these two args.

const (
	llmGatewayResourcesCpuCoreEnv   = "RUNAI_LLM_GATEWAY_RESOURCES_CPU_CORE"
	llmGatewayResourcesCpuMemoryEnv = "RUNAI_LLM_GATEWAY_RESOURCES_CPU_MEMORY"
	llmGatewayImageEnv              = "RUNAI_LLM_GATEWAY_IMAGE"
	llmGatewayImageForTests         = "litellm:latest"
	llmGatewayContainerName         = "llm-gateway"
	llmGatewayPort                  = 4000
	llmServerBasePort               = 8000
	vllmArgModel                    = "--model"
	vllmArgPort                     = "--port"
	vllmEnvVarRpcTimeoutName        = "VLLM_RPC_TIMEOUT"
	vllmEnvVarRpcTimeoutValue       = "60000"
	tgiArgModel                     = "--model-id"
	tgiArgPort                      = "--port"
	nimImageFormat                  = "nvcr.io/nim/nvidia/%s"
	nimEnvVarPort                   = "NIM_SERVER_PORT"
)

type KnativeBuilderModelSwap struct {
}

func (kbms KnativeBuilderModelSwap) createContainersForModelSwap(workload *infv2.InferenceWorkload, resourceRequirements *corev1.ResourceRequirements,
	envs []corev1.EnvVar, command, args []string, sc *corev1.SecurityContext, mounts []corev1.VolumeMount) ([]corev1.Container, error) {

	annotation := kbms.getModelSwapAnnotation(workload.Spec.Annotations)
	params, err := kbms.getModelSwapParamsFromAnnotation(annotation)
	if err != nil {
		return nil, err
	}

	containers := make([]corev1.Container, 0)
	containers = append(containers, createGatewayContainer(workload, params.Models))

	for i, model := range params.Models {
		containers = append(containers, createLlmServerContainer(i, workload, resourceRequirements, envs, command, args, sc, mounts, model, params.ServerType))
	}

	return containers, nil
}

// ModelSwapParams defines the format of the JSON given in the model-swap annotation. it contains the list of models
// and the server type.
type ModelSwapParams struct {
	Models     []string `json:"models"`
	ServerType string   `json:"server-type"`
}

func (kbms KnativeBuilderModelSwap) getModelSwapAnnotation(annotations *flds.DictionaryField) string {
	if annotations == nil {
		return ""
	}

	for key, value := range annotations.Deref() {
		if key == constants.InferenceWorkloadModelSwapAnnotationVolumeKey && value.Value != nil {
			return *value.Value
		}
	}

	return ""
}

func validateModelName(model string) (bool, error) {
	pattern := `^[a-zA-Z0-9-_]+(\x2F[a-zA-Z0-9-_.]+)?$`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("model swap annotation invalid regex pattern: %v", err)
	}

	return regex.MatchString(model), nil
}

func validateServerType(serverType string) error {
	if serverType == "" {
		return fmt.Errorf("model swap annotation is missing server type")
	}

	validServerTypes := map[string]bool{
		"vllm": true,
		"tgi":  true,
		"nim":  true,
	}

	if !validServerTypes[serverType] {
		return fmt.Errorf("model swap annotation contains invalid server type: '%s'", serverType)
	}
	return nil
}

func (kbms KnativeBuilderModelSwap) getModelSwapParamsFromAnnotation(annotation string) (*ModelSwapParams, error) {
	if annotation == "" {
		return nil, fmt.Errorf("model swap annotation is empty")
	}

	var params ModelSwapParams
	if err := json.Unmarshal([]byte(annotation), &params); err != nil {
		return nil, fmt.Errorf("model swap annotation contains invalid JSON: %v", err)
	}

	if len(params.Models) == 0 {
		return nil, fmt.Errorf("model swap annotation must include at least one model")
	}

	for _, model := range params.Models {
		valid, err := validateModelName(model)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, fmt.Errorf("model swap annotation contains invalid model name: '%s'", model)
		}
	}

	if err := validateServerType(params.ServerType); err != nil {
		return nil, err
	}

	return &params, nil
}

func createGatewayContainer(workload *infv2.InferenceWorkload, models []string) corev1.Container {
	cpuCore := os.Getenv(llmGatewayResourcesCpuCoreEnv)
	if cpuCore == "" {
		cpuCore = common.WorkloadRunaiLlmGatewayResourcesCpuCore
	}

	cpuMemory := os.Getenv(llmGatewayResourcesCpuMemoryEnv)
	if cpuMemory == "" {
		cpuMemory = common.WorkloadRunaiLlmGatewayResourcesCpuMemory
	}

	resourceRequirements := corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse(cpuCore),
			corev1.ResourceMemory: resource.MustParse(cpuMemory),
		},
	}

	command := generateGatewayCommand(models, llmGatewayPort)

	image := os.Getenv(llmGatewayImageEnv)
	if image == "" {
		image = llmGatewayImageForTests
	}

	return corev1.Container{
		Name:            llmGatewayContainerName,
		Image:           image,
		ImagePullPolicy: corev1.PullPolicy(workload.Spec.ImagePullPolicy.Deref()),
		Resources:       resourceRequirements,
		Command:         command,
		ReadinessProbe:  getReadinessProbe(llmGatewayPort),
	}
}

// generateGatewayCommand generates a command for the litellm gateway. it first writes a configuration yaml file for
// the gateway to use, and then runs the gateway using the given port and the written config file.
// the configuration sets the given models for the gateway to route. it refers to the models as openai models,
// on localhost, each with its own port, and a dummy api key.
func generateGatewayCommand(models []string, port int) []string {
	var builder strings.Builder

	builder.WriteString("model_list:\n")

	for i, model := range models {
		modelPort := llmServerBasePort + i
		builder.WriteString(fmt.Sprintf("  - model_name: %s\n", model))
		builder.WriteString("    litellm_params:\n")
		builder.WriteString(fmt.Sprintf("      model: openai/%s\n", model))
		builder.WriteString(fmt.Sprintf("      api_base: http://0.0.0.0:%d/v1\n", modelPort))
		builder.WriteString("      api_key: sk-1234\n")
	}

	configFileName := "/tmp/proxy_server_config.yaml"

	command := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf("cat <<EOF > %s\n%sEOF\nlitellm --port %d --config %s",
			configFileName, builder.String(), port, configFileName),
	}

	return command
}

func createLlmServerContainer(index int, workload *infv2.InferenceWorkload, resourceRequirements *corev1.ResourceRequirements,
	envs []corev1.EnvVar, command, args []string, sc *corev1.SecurityContext, mounts []corev1.VolumeMount, model string, serverType string) corev1.Container {

	image := workload.Spec.Image.Deref()
	port := strconv.Itoa(llmServerBasePort + index)

	// we use the workload spec as our base spec, and change image/command/args/envs as needed.
	// - for vllm and tgi, we need to set model and port via args.
	//   note that we assume these args are NOT given as part of the workload spec args.
	// - for nim, we need to set the model via the image url, and the port via an env var.

	switch serverType {
	case "vllm":
		args = append([]string{}, args...)
		args = append(args, vllmArgModel, model)
		args = append(args, vllmArgPort, port)
		// the following env var is required when using swap, as the default timeout is too short
		envs = setOrAddEnvVar(envs, vllmEnvVarRpcTimeoutName, vllmEnvVarRpcTimeoutValue)
	case "tgi":
		args = append([]string{}, args...)
		args = append(args, tgiArgModel, model)
		args = append(args, tgiArgPort, port)
	case "nim":
		image = fmt.Sprintf(nimImageFormat, model)
		envs = append([]corev1.EnvVar{}, envs...)
		envs = setOrAddEnvVar(envs, nimEnvVarPort, port)
	}

	return corev1.Container{
		Name:            fmt.Sprintf("llm-%d", index),
		Image:           image,
		ImagePullPolicy: corev1.PullPolicy(workload.Spec.ImagePullPolicy.Deref()),
		Resources:       *resourceRequirements,
		Env:             envs,
		Command:         command,
		Args:            args,
		ReadinessProbe:  getReadinessProbe(llmServerBasePort + int32(index)),
		SecurityContext: sc,
		VolumeMounts:    mounts,
	}
}

func getReadinessProbe(port int32) *corev1.Probe {
	// we're setting a readiness probe for each container.
	// it requires the multi-container-probing knative feature, see here: https://knative.dev/docs/serving/configuration/feature-flags/#multiple-container-probing
	// available from knative 1.14.

	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: port,
				},
			},
		},
	}
}

// setOrAddEnvVar looks for an env var by name, if found - it sets the given value,
// otherwise, it adds a new env var with the given name and value
func setOrAddEnvVar(envs []corev1.EnvVar, envVarName, envVarValue string) []corev1.EnvVar {
	envVar := corev1.EnvVar{
		Name:  envVarName,
		Value: envVarValue,
	}

	for i, e := range envs {
		if e.Name == envVarName {
			envs[i] = envVar
			return envs
		}
	}

	// if we reached here, the desired env var wasn't found, so we add it
	envs = append(envs, envVar)
	return envs
}
