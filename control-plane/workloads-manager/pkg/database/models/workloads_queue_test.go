package models

import (
	"encoding/json"
	"testing"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
)

func TestToClusterUpdateIncludesCreatorIdentity(t *testing.T) {
	record := WorkloadsQueueRecord{
		ID:              "workload-id",
		Name:            "workload-name",
		CreatedBy:       "owner-uid",
		ResourceGroup:   "batch",
		ResourceVersion: "v1",
		ResourceKind:    "Job",
		DesiredPhase:    string(api.DESIREDPHASE_RUNNING),
	}

	update := record.ToClusterUpdate()

	if update.CreatedBy != "owner-uid" {
		t.Fatalf("unexpected creator identity: got %q, want %q", update.CreatedBy, "owner-uid")
	}
}

func TestNetworkingOwnershipContractJSON(t *testing.T) {
	configuration := WorkloadConfiguration{
		NetworkingOwnershipContract: &NetworkingOwnershipContract{Version: "v1"},
	}

	payload, err := json.Marshal(configuration)
	if err != nil {
		t.Fatalf("marshal networking ownership contract: %v", err)
	}
	if got, want := string(payload), `{"networkingOwnershipContract":{"version":"v1"}}`; got != want {
		t.Fatalf("unexpected networking ownership contract JSON: got %s, want %s", got, want)
	}

	var decoded WorkloadConfiguration
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal networking ownership contract: %v", err)
	}
	if decoded.NetworkingOwnershipContract == nil || decoded.NetworkingOwnershipContract.Version != "v1" {
		t.Fatalf("unexpected decoded networking ownership contract: %#v", decoded.NetworkingOwnershipContract)
	}
}

func TestNetworkingConfigurationUnmarshalSingleEndpointObject(t *testing.T) {
	var cfg WorkloadConfiguration
	err := json.Unmarshal([]byte(`{
		"networking": {
			"services": [
				{
					"ref": {
						"name": "existing-service",
						"port": 8000
					}
				}
			]
		}
	}`), &cfg)
	if err != nil {
		t.Fatalf("unmarshal legacy networking object: %v", err)
	}
	if cfg.Networking == nil {
		t.Fatal("expected networking to be set")
	}

	endpoints := cfg.Networking.ToAPI()
	if endpoints == nil || len(*endpoints) != 1 {
		t.Fatalf("expected one normalized endpoint, got %#v", endpoints)
	}
	services := (*endpoints)[0].Services
	if services == nil || len(*services) != 1 {
		t.Fatalf("expected one service, got %#v", services)
	}
	if (*services)[0].Ref == nil || (*services)[0].Ref.Name != "existing-service" || (*services)[0].Ref.Port != 8000 {
		t.Fatalf("unexpected service ref: %#v", (*services)[0].Ref)
	}
}

func TestNetworkingConfigurationUnmarshalLegacyCreate(t *testing.T) {
	tests := map[string]string{
		"single endpoint object": `{
			"networking": {
				"services": [
					{
						"create": {
							"port": 8080,
							"selector": {
								"app": "legacy-workload"
							}
						}
					}
				]
			}
		}`,
		"endpoint array": `{
			"networking": [
				{
					"services": [
						{
							"create": {
								"port": 8080,
								"selector": {
									"app": "legacy-workload"
								}
							}
						}
					]
				}
			]
		}`,
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			var cfg WorkloadConfiguration
			if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
				t.Fatalf("unmarshal legacy create networking: %v", err)
			}
			if cfg.Networking == nil {
				t.Fatal("expected networking to be set")
			}

			endpoints := cfg.Networking.ToAPI("legacy-workload")
			if endpoints == nil || len(*endpoints) != 1 {
				t.Fatalf("expected one normalized endpoint, got %#v", endpoints)
			}
			services := (*endpoints)[0].Services
			if services == nil || len(*services) != 1 {
				t.Fatalf("expected one service, got %#v", services)
			}
			spec := (*services)[0].Spec
			if spec == nil {
				t.Fatal("expected legacy create to be normalized into service spec")
			}
			if spec.Name != "legacy-workload" {
				t.Fatalf("expected workload name to be filled in service spec, got %q", spec.Name)
			}
			if spec.Type == nil || *spec.Type != api.SERVICETYPE_CLUSTER_IP {
				t.Fatalf("expected ClusterIP service type, got %#v", spec.Type)
			}
			if spec.Selector == nil || (*spec.Selector)["app"] != "legacy-workload" {
				t.Fatalf("unexpected selector: %#v", spec.Selector)
			}
			if len(spec.Ports) != 1 || spec.Ports[0].Port != 8080 {
				t.Fatalf("unexpected ports: %#v", spec.Ports)
			}

			agentConfig := cfg.Networking.ToAgentAPI("legacy-workload")
			if agentConfig == nil || agentConfig.Services == nil || len(*agentConfig.Services) != 1 {
				t.Fatalf("expected legacy create to be projected to agent networking, got %#v", agentConfig)
			}
			create := (*agentConfig.Services)[0].Create
			if create == nil || create.Port != 8080 || create.Selector == nil || (*create.Selector)["app"] != "legacy-workload" {
				t.Fatalf("unexpected projected create: %#v", create)
			}
		})
	}
}

func TestNetworkingConfigurationUnmarshalEndpointArray(t *testing.T) {
	var cfg WorkloadConfiguration
	err := json.Unmarshal([]byte(`{
		"networking": [
			{
				"services": [
					{
						"ref": {
							"name": "existing-service",
							"port": 8000
						}
					}
				]
			}
		]
	}`), &cfg)
	if err != nil {
		t.Fatalf("unmarshal networking endpoint array: %v", err)
	}
	if cfg.Networking == nil {
		t.Fatal("expected networking to be set")
	}

	endpoints := cfg.Networking.ToAPI()
	if endpoints == nil || len(*endpoints) != 1 {
		t.Fatalf("expected one endpoint, got %#v", endpoints)
	}
}

func TestNetworkingConfigurationToAgentAPI(t *testing.T) {
	portName := "http"
	nodePort := int32(30080)
	serviceType := api.SERVICETYPE_NODE_PORT
	selector := map[string]string{"app": "workload"}
	config := NetworkingConfiguration{{
		Services: &[]api.NetworkingService{{
			Spec: &api.ServiceSpec{
				Name:     "workload",
				Type:     &serviceType,
				Selector: &selector,
				Ports: []api.ServicePort{{
					Name:     &portName,
					Port:     8080,
					NodePort: &nodePort,
				}},
			},
		}},
	}}

	if agentConfig := config.ToAgentAPI("workload"); agentConfig != nil {
		t.Fatalf("expected rich service spec to be omitted from agent networking projection, got %#v", agentConfig)
	}

	config[0].Services = &[]api.NetworkingService{{
		Spec: &api.ServiceSpec{
			Name:     "workload",
			Selector: &selector,
			Ports: []api.ServicePort{{
				Port: 8080,
			}},
		},
		Ref: &api.ServiceRef{Name: "workload", Port: 8080},
	}}
	if err := config.AgentNetworkingCompatibilityError("workload"); err == nil {
		t.Fatal("expected service with both spec and ref to be incompatible with agent networking projection")
	}

	config[0].Services = &[]api.NetworkingService{{}}
	if err := config.AgentNetworkingCompatibilityError("workload"); err == nil {
		t.Fatal("expected service without spec or ref to be incompatible with agent networking projection")
	}

	serviceType = api.SERVICETYPE_CLUSTER_IP
	config[0].Services = &[]api.NetworkingService{{
		Spec: &api.ServiceSpec{
			Name:     "workload",
			Type:     &serviceType,
			Selector: &selector,
			Ports: []api.ServicePort{{
				Port: 8080,
			}},
		},
	}}

	agentConfig := config.ToAgentAPI("workload")
	if agentConfig == nil || agentConfig.Services == nil || len(*agentConfig.Services) != 1 {
		t.Fatalf("expected one projected legacy service, got %#v", agentConfig)
	}
	create := (*agentConfig.Services)[0].Create
	if create == nil || create.Port != 8080 || create.Selector == nil || (*create.Selector)["app"] != "workload" {
		t.Fatalf("unexpected projected service create: %#v", create)
	}
}

func TestNetworkingConfigurationServiceLabelsRoundTripWithoutMutation(t *testing.T) {
	payload := []byte(`{
		"networking": [
			{"services":[{"spec":{
				"name":"serving","selector":{"app":"model"},"labels":{
					"run.ai/nim-workload":"true",
					"run.ai/nim-workload-metrics-service":"true",
					"ordinary":"keep"
				},"ports":[{"port":8000}]
			}}]},
			{"services":[{"spec":{
				"name":"metrics","selector":{"app":"model"},"labels":{
					"run.ai/vllm-workload":"true",
					"run.ai/vllm-workload-metrics-service":"true",
					"shared":"metrics"
				},"ports":[{"name":"metrics","port":9090,"targetPort":9091}]
			}}]}
		]
	}`)

	var cfg WorkloadConfiguration
	if err := json.Unmarshal(payload, &cfg); err != nil {
		t.Fatalf("unmarshal networking labels: %v", err)
	}

	first := cfg.Networking.ToAPI("workload")
	if first == nil || len(*first) != 2 {
		t.Fatalf("expected two endpoints, got %#v", first)
	}
	if got := (*(*first)[0].Services)[0].Spec.Labels["ordinary"]; got != "keep" {
		t.Fatalf("ordinary label lost: %q", got)
	}
	if got := (*(*first)[1].Services)[0].Spec.Labels["run.ai/vllm-workload-metrics-service"]; got != "true" {
		t.Fatalf("vLLM metrics label lost: %q", got)
	}

	(*(*first)[0].Services)[0].Spec.Labels["ordinary"] = "mutated"
	(*(*first)[0].Services)[0].Spec.Ports[0].Port = 1
	second := cfg.Networking.ToAPI("workload")
	if got := (*(*second)[0].Services)[0].Spec.Labels["ordinary"]; got != "keep" {
		t.Fatalf("stored labels were mutated through API projection: %q", got)
	}
	if got := (*(*second)[0].Services)[0].Spec.Ports[0].Port; got != 8000 {
		t.Fatalf("stored ports were mutated through API projection: %d", got)
	}

	record := WorkloadsQueueRecord{Name: "workload", Configuration: &cfg}
	update := record.ToClusterUpdate()
	if update.Configuration.NetworkingEndpoints == nil || len(*update.Configuration.NetworkingEndpoints) != 2 {
		t.Fatalf("expected two endpoints in agent update, got %#v", update.Configuration.NetworkingEndpoints)
	}
	if got := (*(*update.Configuration.NetworkingEndpoints)[1].Services)[0].Spec.Labels["shared"]; got != "metrics" {
		t.Fatalf("ordinary shared metrics label lost from agent update: %q", got)
	}
}
