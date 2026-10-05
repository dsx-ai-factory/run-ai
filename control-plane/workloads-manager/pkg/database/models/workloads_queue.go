/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

const (
	WorkloadsQueueTableName               = "workloads_queue"
	WorkloadsQueueColumnId                = "id"
	WorkloadsQueueColumnUpdateCount       = "update_count"
	WorkloadsQueueColumnTenantId          = "tenant_id"
	WorkloadsQueueColumnClusterId         = "cluster_id"
	WorkloadsQueueColumnDepartmentId      = "department_id"
	WorkloadsQueueColumnProjectId         = "project_id"
	WorkloadsQueueColumnRawManifest       = "raw_manifest"
	WorkloadsQueueColumnEffectiveManifest = "effective_manifest"
	WorkloadsQueueColumnDesiredPhase      = "desired_phase"
	WorkloadsQueueColumnUpdatedAt         = "updated_at"
	WorkloadsQueueColumnUpdatedBy         = "updated_by"
	WorkloadsQueueColumnTranslatedFromV1  = "translated_from_v1"
)

// MnnvlConfiguration represents MNNVL configuration
type MnnvlConfiguration struct {
	Mode                            api.MnnvlMode `json:"mode"`
	GpuNetworkAccelerationLabelKeys []string      `json:"gpuNetworkAccelerationLabelKeys,omitempty"`
}

// UserCredentialsConfiguration carries personal-credential sources from CP to
// the agent, which forwards them as the run.ai/personal-credentials annotation on
// the WorkloadRunner.
type UserCredentialsConfiguration struct {
	Sources []string `json:"sources,omitempty"`
}

// NetworkingConfiguration stores the current endpoint-array shape in the DB.
// It also reads legacy single-endpoint object JSON written before networking
// became an array, so existing workloads_queue records remain readable.
type NetworkingConfiguration []api.NetworkingEndpoint

type networkingEndpointCompat struct {
	ExternalAccess *api.ExternalAccess        `json:"externalAccess,omitempty"`
	Services       *[]networkingServiceCompat `json:"services,omitempty"`
}

type networkingServiceCompat struct {
	Spec          *api.ServiceSpec             `json:"spec,omitempty"`
	Create        *api.AgentServiceCreate      `json:"create,omitempty"`
	Ref           *api.ServiceRef              `json:"ref,omitempty"`
	Authorization *api.Workloadv2Authorization `json:"authorization,omitempty"`
}

func (n *NetworkingConfiguration) UnmarshalJSON(data []byte) error {
	cfg, err := parseNetworkingConfiguration(data)
	if err != nil {
		return err
	}
	*n = cfg
	return nil
}

func parseNetworkingConfiguration(data []byte) (NetworkingConfiguration, error) {
	var endpoints []networkingEndpointCompat
	if err := json.Unmarshal(data, &endpoints); err == nil {
		return networkingEndpointsFromCompat(endpoints), nil
	}

	var legacyEndpoint networkingEndpointCompat
	if err := json.Unmarshal(data, &legacyEndpoint); err != nil {
		return nil, fmt.Errorf("networking: expected endpoint array or legacy endpoint object: %w", err)
	}

	return networkingEndpointsFromCompat([]networkingEndpointCompat{legacyEndpoint}), nil
}

func networkingEndpointsFromCompat(endpoints []networkingEndpointCompat) NetworkingConfiguration {
	result := make(NetworkingConfiguration, len(endpoints))
	for i, endpoint := range endpoints {
		result[i] = api.NetworkingEndpoint{
			ExternalAccess: endpoint.ExternalAccess,
			Services:       networkingServicesFromCompat(endpoint.Services),
		}
	}
	return result
}

func networkingServicesFromCompat(services *[]networkingServiceCompat) *[]api.NetworkingService {
	if services == nil {
		return nil
	}
	result := make([]api.NetworkingService, len(*services))
	for i, service := range *services {
		result[i] = api.NetworkingService{
			Spec:          service.Spec,
			Ref:           service.Ref,
			Authorization: service.Authorization,
		}
		if result[i].Spec == nil && service.Create != nil {
			result[i].Spec = serviceCreateToSpec(service.Create)
		}
	}
	return &result
}

func serviceCreateToSpec(create *api.AgentServiceCreate) *api.ServiceSpec {
	if create == nil {
		return nil
	}
	serviceType := api.SERVICETYPE_CLUSTER_IP
	return &api.ServiceSpec{
		Type:     &serviceType,
		Selector: cloneStringMapPtr(create.Selector),
		Ports: []api.ServicePort{{
			Port: create.Port,
		}},
	}
}

func (n *NetworkingConfiguration) ToAPI(workloadName ...string) *[]api.NetworkingEndpoint {
	if n == nil {
		return nil
	}
	endpoints := make([]api.NetworkingEndpoint, len(*n))
	for i, endpoint := range *n {
		endpoints[i] = cloneNetworkingEndpoint(endpoint)
	}
	if len(workloadName) > 0 && workloadName[0] != "" {
		fillMissingServiceSpecNames(endpoints, workloadName[0])
	}
	return &endpoints
}

// cloneNetworkingEndpoint creates an independent API projection so callers
// cannot mutate the stored configuration through nested pointers, maps, or slices.
func cloneNetworkingEndpoint(endpoint api.NetworkingEndpoint) api.NetworkingEndpoint {
	return api.NetworkingEndpoint{
		ExternalAccess: cloneExternalAccess(endpoint.ExternalAccess),
		Services:       cloneNetworkingServices(endpoint.Services),
	}
}

func cloneNetworkingServices(services *[]api.NetworkingService) *[]api.NetworkingService {
	if services == nil {
		return nil
	}
	out := make([]api.NetworkingService, len(*services))
	for i, service := range *services {
		out[i] = api.NetworkingService{
			Spec:          cloneServiceSpec(service.Spec),
			Ref:           clonePtr(service.Ref),
			Authorization: cloneWorkloadv2Authorization(service.Authorization),
		}
	}
	return &out
}

func cloneServiceSpec(spec *api.ServiceSpec) *api.ServiceSpec {
	if spec == nil {
		return nil
	}
	out := &api.ServiceSpec{
		Name:     spec.Name,
		Type:     clonePtr(spec.Type),
		Selector: cloneStringMapPtr(spec.Selector),
		Labels:   maps.Clone(spec.Labels),
		Ports:    make([]api.ServicePort, len(spec.Ports)),
	}
	for i, port := range spec.Ports {
		out.Ports[i] = api.ServicePort{
			Name:       clonePtr(port.Name),
			Port:       port.Port,
			TargetPort: clonePtr(port.TargetPort),
			NodePort:   clonePtr(port.NodePort),
			Protocol:   clonePtr(port.Protocol),
		}
	}
	return out
}

func fillMissingServiceSpecNames(endpoints []api.NetworkingEndpoint, workloadName string) {
	for i := range endpoints {
		if endpoints[i].Services == nil {
			continue
		}
		services := append([]api.NetworkingService(nil), (*endpoints[i].Services)...)
		endpoints[i].Services = &services
		for j := range services {
			if services[j].Spec == nil || services[j].Spec.Name != "" {
				continue
			}
			spec := *services[j].Spec
			spec.Name = workloadName
			services[j].Spec = &spec
		}
	}
}

func (n *NetworkingConfiguration) ToAgentAPI(workloadName string) *api.AgentNetworkingConfiguration {
	if n == nil || len(*n) == 0 {
		return nil
	}
	if err := n.AgentNetworkingCompatibilityError(workloadName); err != nil {
		return nil
	}

	config := &api.AgentNetworkingConfiguration{}
	endpoint := (*n)[0]
	if endpoint.ExternalAccess != nil {
		config.ExternalAccess = cloneExternalAccess(endpoint.ExternalAccess)
	}
	if endpoint.Services != nil && len(*endpoint.Services) > 0 {
		services := []api.AgentNetworkingService{networkingServiceToAgentAPI((*endpoint.Services)[0])}
		config.Services = &services
	}

	if config.ExternalAccess == nil && config.Services == nil {
		return nil
	}
	return config
}

func (n NetworkingConfiguration) AgentNetworkingCompatibilityError(workloadName string) error {
	if len(n) > 1 {
		return fmt.Errorf("multiple networking endpoints cannot be represented for this cluster version")
	}
	if len(n) == 0 {
		return nil
	}

	endpoint := n[0]
	if endpoint.Services == nil || len(*endpoint.Services) == 0 {
		return nil
	}
	if len(*endpoint.Services) > 1 {
		return fmt.Errorf("multiple networking services cannot be represented for this cluster version")
	}

	service := (*endpoint.Services)[0]
	hasSpec := service.Spec != nil
	hasRef := service.Ref != nil
	if hasSpec && hasRef {
		return fmt.Errorf("services[0]: spec and ref are mutually exclusive")
	}
	if !hasSpec && !hasRef {
		return fmt.Errorf("services[0]: either spec or ref must be set")
	}
	if !hasSpec {
		return nil
	}
	if err := agentServiceSpecCompatibilityError(service.Spec, workloadName); err != nil {
		return fmt.Errorf("services[0].spec: %w", err)
	}
	return nil
}

func agentServiceSpecCompatibilityError(spec *api.ServiceSpec, workloadName string) error {
	if spec == nil {
		return nil
	}
	if spec.Name != "" && spec.Name != workloadName {
		return fmt.Errorf("name must match the workload name for this cluster version")
	}
	if spec.Type != nil && *spec.Type != api.SERVICETYPE_CLUSTER_IP {
		return fmt.Errorf("type %q cannot be represented for this cluster version", *spec.Type)
	}
	if spec.Selector == nil || len(*spec.Selector) == 0 {
		return fmt.Errorf("selector is required for this cluster version")
	}
	if len(spec.Ports) != 1 {
		return fmt.Errorf("exactly one port can be represented for this cluster version")
	}

	port := spec.Ports[0]
	if port.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	if port.Name != nil {
		return fmt.Errorf("port name cannot be represented for this cluster version")
	}
	if port.NodePort != nil {
		return fmt.Errorf("nodePort cannot be represented for this cluster version")
	}
	if port.TargetPort != nil && *port.TargetPort != port.Port {
		return fmt.Errorf("targetPort must be omitted or equal to port for this cluster version")
	}
	if port.Protocol != nil && *port.Protocol != api.SERVICEPROTOCOL_TCP {
		return fmt.Errorf("protocol %q cannot be represented for this cluster version", *port.Protocol)
	}
	return nil
}

func networkingServiceToAgentAPI(service api.NetworkingService) api.AgentNetworkingService {
	agentService := api.AgentNetworkingService{
		Ref:           clonePtr(service.Ref),
		Authorization: cloneWorkloadv2Authorization(service.Authorization),
	}
	if service.Spec != nil && len(service.Spec.Ports) > 0 {
		agentService.Create = &api.AgentServiceCreate{
			Port:     service.Spec.Ports[0].Port,
			Selector: cloneStringMapPtr(service.Spec.Selector),
		}
	}
	return agentService
}

func cloneExternalAccess(access *api.ExternalAccess) *api.ExternalAccess {
	if access == nil {
		return nil
	}
	return &api.ExternalAccess{
		Ingress:   cloneInterfaceMapPtr(access.Ingress),
		Route:     cloneInterfaceMapPtr(access.Route),
		HttpRoute: cloneInterfaceMapPtr(access.HttpRoute),
	}
}

func cloneInterfaceMapPtr(in *map[string]interface{}) *map[string]interface{} {
	if in == nil {
		return nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return &out
}

func cloneStringMapPtr(in *map[string]string) *map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(*in))
	for k, v := range *in {
		out[k] = v
	}
	return &out
}

func cloneWorkloadv2Authorization(auth *api.Workloadv2Authorization) *api.Workloadv2Authorization {
	if auth == nil {
		return nil
	}
	return &api.Workloadv2Authorization{
		AuthorizationType: clonePtr(auth.AuthorizationType),
		AuthorizedUsers:   cloneStringSlicePtr(auth.AuthorizedUsers),
		AuthorizedGroups:  cloneStringSlicePtr(auth.AuthorizedGroups),
		ForceJsonResponse: clonePtr(auth.ForceJsonResponse),
	}
}

func cloneStringSlicePtr(in *[]string) *[]string {
	if in == nil {
		return nil
	}
	out := append([]string(nil), (*in)...)
	return &out
}

func clonePtr[T any](in *T) *T {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

// WorkloadConfiguration represents all workload configuration
type WorkloadConfiguration struct {
	Mnnvl      *MnnvlConfiguration      `json:"mnnvl,omitempty"`
	Networking *NetworkingConfiguration `json:"networking,omitempty"`
	// InferenceEndpoint is resolved by the server, never accepted from the public
	// workload configuration. Nil preserves the behavior of existing workloads.
	InferenceEndpoint *api.AgentInferenceEndpoint `json:"inferenceEndpoint,omitempty"`
	// NetworkingOwnershipContract is set only when Workloads Manager applies
	// the binding during creation. It remains nil for older records so an
	// update cannot add a selector that their existing Pods do not carry.
	NetworkingOwnershipContract *NetworkingOwnershipContract                 `json:"networkingOwnershipContract,omitempty"`
	UserCredentials             *UserCredentialsConfiguration                `json:"userCredentials,omitempty"`
	Pvcs                        []api.WorkloadV2UpdateConfigurationPvc       `json:"pvcs,omitempty"`
	PvcsConfig                  *api.WorkloadV2UpdateConfigurationPvcsConfig `json:"pvcsConfig,omitempty"`
}

// NetworkingOwnershipContract describes the versioned server-side binding
// applied to a workload's networking configuration and Pod templates.
type NetworkingOwnershipContract struct {
	Version string `json:"version"`
}

type WorkloadsQueueRecord struct {
	ID string `gorm:"primaryKey"`

	CreatedAt time.Time `gorm:"not null;"`
	CreatedBy string    `gorm:"not null;"`
	UpdatedAt time.Time `gorm:"not null;"`
	UpdatedBy string    `gorm:"not null;"`
	DeletedAt gorm.DeletedAt
	DeletedBy *string

	TenantId       int32  `gorm:"not null;"`
	ClusterId      string `gorm:"not null;"`
	DepartmentId   string `gorm:"not null;"`
	DepartmentName string `gorm:"not null;"`
	ProjectId      string `gorm:"not null;"`
	ProjectName    string `gorm:"not null;"`

	Name string `gorm:"not null"`

	Category       *string
	PriorityClass  *string
	Preemptibility *string
	Configuration  *WorkloadConfiguration `gorm:"type:jsonb;serializer:json"`

	ResourceGroup   string `gorm:"not null"`
	ResourceVersion string `gorm:"not null"`
	ResourceKind    string `gorm:"not null"`

	DesiredPhase string

	// TranslatedFromV1 indicates the workload carries the run.ai/converted-from-native
	// label, meaning it was created via the v1 native API and cast to v2.
	TranslatedFromV1 bool `gorm:"not null;default:false"`

	RawManifest       dbmodels.JsonBMap
	EffectiveManifest dbmodels.JsonBMap

	UpdateCount int64 `gorm:"not null"`

	InputRequestPath string `gorm:"not null"`
	InputRequestBody dbmodels.JsonBMap
}

func (*WorkloadsQueueRecord) TableName() string {
	return WorkloadsQueueTableName
}

func (q *WorkloadsQueueRecord) BeforeCreate(_ *gorm.DB) (err error) {
	if q.ID == "" {
		q.ID = uuid.NewString()
	}
	return
}

func (q *WorkloadsQueueRecord) ToWorkloadV2() api.WorkloadV2 {
	result := api.WorkloadV2{
		Metadata: api.WorkloadV2Metadata{
			Id: q.ID,

			CreatedBy: q.CreatedBy,
			CreatedAt: q.CreatedAt,
			UpdatedAt: q.UpdatedAt,
			UpdatedBy: q.UpdatedBy,
			DeletedBy: q.DeletedBy,

			Name: q.Name,

			TenantId:       q.TenantId,
			ClusterId:      q.ClusterId,
			DepartmentId:   q.DepartmentId,
			DepartmentName: q.DepartmentName,
			ProjectId:      q.ProjectId,
			ProjectName:    q.ProjectName,

			Priority: q.PriorityClass,
			Category: q.Category,

			Gvk: api.Gvk{
				Group:   convutils.Ref(q.ResourceGroup),
				Kind:    q.ResourceKind,
				Version: q.ResourceVersion,
			},
		},

		DesiredPhase: api.DesiredPhase(q.DesiredPhase),

		RawManifest: q.RawManifest,

		WorkloadManifest: q.EffectiveManifest,
	}

	if q.Preemptibility != nil {
		result.Metadata.Preemptibility = convutils.Ref(api.Preemptibility(*q.Preemptibility))
	}

	if q.Configuration != nil {
		result.Metadata.Configuration = api.WorkloadV2Configuration{}
		if q.Configuration.Mnnvl != nil {
			result.Metadata.Configuration.Mnnvl = &q.Configuration.Mnnvl.Mode
		}
		if q.Configuration.Networking != nil {
			result.Metadata.Configuration.Networking = q.Configuration.Networking.ToAPI(q.Name)
		}
		if len(q.Configuration.Pvcs) > 0 {
			result.Metadata.Configuration.Pvcs = &q.Configuration.Pvcs
		}
		if q.Configuration.PvcsConfig != nil {
			result.Metadata.Configuration.PvcsConfig = q.Configuration.PvcsConfig
		}
	}

	if q.DeletedAt.Valid {
		result.Metadata.DeletedAt = &q.DeletedAt.Time
	}

	return result
}

func (q *WorkloadsQueueRecord) ToGetResponse() api.WorkloadV2GetResponse {
	result := api.WorkloadV2GetResponse{
		Workload: q.ToWorkloadV2(),
	}

	return result
}

func (q *WorkloadsQueueRecord) ToClusterUpdate() api.WorkloadV2Update {
	result := api.WorkloadV2Update{
		Id:        q.ID,
		Name:      q.Name,
		CreatedBy: q.CreatedBy,
		ProjectId: q.ProjectId,
		Gvk: api.Gvk{
			Group:   convutils.Ref(q.ResourceGroup),
			Kind:    q.ResourceKind,
			Version: q.ResourceVersion,
		},
		UpdateCount:  q.UpdateCount,
		DesiredPhase: api.DesiredPhase(q.DesiredPhase),
	}

	if q.Configuration != nil {
		cfg := api.WorkloadV2UpdateConfiguration{}
		if endpoint := q.Configuration.InferenceEndpoint; endpoint != nil {
			cfg.InferenceEndpoint = &api.AgentInferenceEndpoint{SchemaVersion: endpoint.SchemaVersion}
			if spec := cloneInterfaceMapPtr(&endpoint.Spec); spec != nil {
				cfg.InferenceEndpoint.Spec = *spec
			}
		}
		if q.Configuration.Networking != nil {
			cfg.Networking = q.Configuration.Networking.ToAgentAPI(q.Name)
			cfg.NetworkingEndpoints = q.Configuration.Networking.ToAPI(q.Name)
		}
		if q.Configuration.Mnnvl != nil {
			cfg.Mnnvl = api.WorkloadV2UpdateConfigurationMnnvl{
				Mode:                            q.Configuration.Mnnvl.Mode,
				GpuNetworkAccelerationLabelKeys: q.Configuration.Mnnvl.GpuNetworkAccelerationLabelKeys,
			}
		}
		if pc := q.Configuration.UserCredentials; pc != nil && len(pc.Sources) > 0 {
			sources := make([]api.WorkloadV2UpdateConfigurationUserCredentialSource, 0, len(pc.Sources))
			for _, s := range pc.Sources {
				sources = append(sources, api.WorkloadV2UpdateConfigurationUserCredentialSource{Ref: s})
			}
			cfg.UserCredentials = &api.WorkloadV2UpdateConfigurationUserCredentials{
				Sources: sources,
			}
		}
		if len(q.Configuration.Pvcs) > 0 {
			cfg.Pvcs = &q.Configuration.Pvcs
		}
		if q.Configuration.PvcsConfig != nil {
			cfg.PvcsConfig = q.Configuration.PvcsConfig
		}
		result.Configuration = cfg
	}

	if result.DesiredPhase != api.DESIREDPHASE_DELETED {
		result.WorkloadManifest = q.EffectiveManifest
	}

	return result
}

func (q *WorkloadsQueueRecord) DebugInfo() string {
	return fmt.Sprintf("[%d-%s-%s-%s] %s(%s)",
		q.TenantId, q.ClusterId, q.DepartmentId, q.ProjectId, q.ProjectName, q.Name)
}
