# ClusterControlPlaneConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Agent** | Pointer to [**ClusterControlPlaneConfigAgent**](ClusterControlPlaneConfigAgent.md) |  | [optional] 
**Workloads** | Pointer to [**ClusterControlPlaneConfigWorkloads**](ClusterControlPlaneConfigWorkloads.md) |  | [optional] 
**MetricsRemoteWrite** | Pointer to [**ClusterControlPlaneConfigMetricsRemoteWrite**](ClusterControlPlaneConfigMetricsRemoteWrite.md) |  | [optional] 
**MetricsRemoteWrites** | Pointer to [**[]ClusterControlPlaneConfigMetricsRemoteWrite**](ClusterControlPlaneConfigMetricsRemoteWrite.md) | Additional remote write configurations. | [optional] 
**ResearcherService** | Pointer to [**ClusterControlPlaneConfigResearcherService**](ClusterControlPlaneConfigResearcherService.md) |  | [optional] 
**ClusterName** | Pointer to **string** | The cluster&#39;s name. | [optional] 
**TenantId** | Pointer to **int** | the tenant&#39;s identifier | [optional] 
**ReleaseName** | Pointer to **string** | the name of the release | [optional] 
**Environment** | Pointer to **string** | test/staging/production/self-hosted | [optional] 
**DockerConfigJson** | Pointer to **string** | The cluster&#39;s docker config json. | [optional] 

## Methods

### NewClusterControlPlaneConfig

`func NewClusterControlPlaneConfig() *ClusterControlPlaneConfig`

NewClusterControlPlaneConfig instantiates a new ClusterControlPlaneConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterControlPlaneConfigWithDefaults

`func NewClusterControlPlaneConfigWithDefaults() *ClusterControlPlaneConfig`

NewClusterControlPlaneConfigWithDefaults instantiates a new ClusterControlPlaneConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAgent

`func (o *ClusterControlPlaneConfig) GetAgent() ClusterControlPlaneConfigAgent`

GetAgent returns the Agent field if non-nil, zero value otherwise.

### GetAgentOk

`func (o *ClusterControlPlaneConfig) GetAgentOk() (*ClusterControlPlaneConfigAgent, bool)`

GetAgentOk returns a tuple with the Agent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAgent

`func (o *ClusterControlPlaneConfig) SetAgent(v ClusterControlPlaneConfigAgent)`

SetAgent sets Agent field to given value.

### HasAgent

`func (o *ClusterControlPlaneConfig) HasAgent() bool`

HasAgent returns a boolean if a field has been set.

### GetWorkloads

`func (o *ClusterControlPlaneConfig) GetWorkloads() ClusterControlPlaneConfigWorkloads`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *ClusterControlPlaneConfig) GetWorkloadsOk() (*ClusterControlPlaneConfigWorkloads, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *ClusterControlPlaneConfig) SetWorkloads(v ClusterControlPlaneConfigWorkloads)`

SetWorkloads sets Workloads field to given value.

### HasWorkloads

`func (o *ClusterControlPlaneConfig) HasWorkloads() bool`

HasWorkloads returns a boolean if a field has been set.

### GetMetricsRemoteWrite

`func (o *ClusterControlPlaneConfig) GetMetricsRemoteWrite() ClusterControlPlaneConfigMetricsRemoteWrite`

GetMetricsRemoteWrite returns the MetricsRemoteWrite field if non-nil, zero value otherwise.

### GetMetricsRemoteWriteOk

`func (o *ClusterControlPlaneConfig) GetMetricsRemoteWriteOk() (*ClusterControlPlaneConfigMetricsRemoteWrite, bool)`

GetMetricsRemoteWriteOk returns a tuple with the MetricsRemoteWrite field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricsRemoteWrite

`func (o *ClusterControlPlaneConfig) SetMetricsRemoteWrite(v ClusterControlPlaneConfigMetricsRemoteWrite)`

SetMetricsRemoteWrite sets MetricsRemoteWrite field to given value.

### HasMetricsRemoteWrite

`func (o *ClusterControlPlaneConfig) HasMetricsRemoteWrite() bool`

HasMetricsRemoteWrite returns a boolean if a field has been set.

### GetMetricsRemoteWrites

`func (o *ClusterControlPlaneConfig) GetMetricsRemoteWrites() []ClusterControlPlaneConfigMetricsRemoteWrite`

GetMetricsRemoteWrites returns the MetricsRemoteWrites field if non-nil, zero value otherwise.

### GetMetricsRemoteWritesOk

`func (o *ClusterControlPlaneConfig) GetMetricsRemoteWritesOk() (*[]ClusterControlPlaneConfigMetricsRemoteWrite, bool)`

GetMetricsRemoteWritesOk returns a tuple with the MetricsRemoteWrites field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricsRemoteWrites

`func (o *ClusterControlPlaneConfig) SetMetricsRemoteWrites(v []ClusterControlPlaneConfigMetricsRemoteWrite)`

SetMetricsRemoteWrites sets MetricsRemoteWrites field to given value.

### HasMetricsRemoteWrites

`func (o *ClusterControlPlaneConfig) HasMetricsRemoteWrites() bool`

HasMetricsRemoteWrites returns a boolean if a field has been set.

### GetResearcherService

`func (o *ClusterControlPlaneConfig) GetResearcherService() ClusterControlPlaneConfigResearcherService`

GetResearcherService returns the ResearcherService field if non-nil, zero value otherwise.

### GetResearcherServiceOk

`func (o *ClusterControlPlaneConfig) GetResearcherServiceOk() (*ClusterControlPlaneConfigResearcherService, bool)`

GetResearcherServiceOk returns a tuple with the ResearcherService field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResearcherService

`func (o *ClusterControlPlaneConfig) SetResearcherService(v ClusterControlPlaneConfigResearcherService)`

SetResearcherService sets ResearcherService field to given value.

### HasResearcherService

`func (o *ClusterControlPlaneConfig) HasResearcherService() bool`

HasResearcherService returns a boolean if a field has been set.

### GetClusterName

`func (o *ClusterControlPlaneConfig) GetClusterName() string`

GetClusterName returns the ClusterName field if non-nil, zero value otherwise.

### GetClusterNameOk

`func (o *ClusterControlPlaneConfig) GetClusterNameOk() (*string, bool)`

GetClusterNameOk returns a tuple with the ClusterName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterName

`func (o *ClusterControlPlaneConfig) SetClusterName(v string)`

SetClusterName sets ClusterName field to given value.

### HasClusterName

`func (o *ClusterControlPlaneConfig) HasClusterName() bool`

HasClusterName returns a boolean if a field has been set.

### GetTenantId

`func (o *ClusterControlPlaneConfig) GetTenantId() int`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ClusterControlPlaneConfig) GetTenantIdOk() (*int, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ClusterControlPlaneConfig) SetTenantId(v int)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ClusterControlPlaneConfig) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetReleaseName

`func (o *ClusterControlPlaneConfig) GetReleaseName() string`

GetReleaseName returns the ReleaseName field if non-nil, zero value otherwise.

### GetReleaseNameOk

`func (o *ClusterControlPlaneConfig) GetReleaseNameOk() (*string, bool)`

GetReleaseNameOk returns a tuple with the ReleaseName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseName

`func (o *ClusterControlPlaneConfig) SetReleaseName(v string)`

SetReleaseName sets ReleaseName field to given value.

### HasReleaseName

`func (o *ClusterControlPlaneConfig) HasReleaseName() bool`

HasReleaseName returns a boolean if a field has been set.

### GetEnvironment

`func (o *ClusterControlPlaneConfig) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *ClusterControlPlaneConfig) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *ClusterControlPlaneConfig) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.

### HasEnvironment

`func (o *ClusterControlPlaneConfig) HasEnvironment() bool`

HasEnvironment returns a boolean if a field has been set.

### GetDockerConfigJson

`func (o *ClusterControlPlaneConfig) GetDockerConfigJson() string`

GetDockerConfigJson returns the DockerConfigJson field if non-nil, zero value otherwise.

### GetDockerConfigJsonOk

`func (o *ClusterControlPlaneConfig) GetDockerConfigJsonOk() (*string, bool)`

GetDockerConfigJsonOk returns a tuple with the DockerConfigJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerConfigJson

`func (o *ClusterControlPlaneConfig) SetDockerConfigJson(v string)`

SetDockerConfigJson sets DockerConfigJson field to given value.

### HasDockerConfigJson

`func (o *ClusterControlPlaneConfig) HasDockerConfigJson() bool`

HasDockerConfigJson returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


