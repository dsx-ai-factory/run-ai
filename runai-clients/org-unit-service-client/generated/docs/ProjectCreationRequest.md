# ProjectCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | The description of the project. | [optional] 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | An ordered list of node pools that defines the default scheduling preference for workloads in this project. The Scheduler will attempt to schedule workloads on node pools in this order. This setting is enforced if no node pool list is defined in the workload policy. | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**Name** | **string** | The unique name of the project within the cluster. Must be a valid Kubernetes name (lowercase, alphanumeric, hyphens allowed). | 
**ClusterId** | **string** | The id of the cluster. | 
**RequestedNamespace** | Pointer to **string** | The custom Kubernetes namespace name for the project. If not specified, a namespace will be auto-generated using the pattern runai-&lt;project-name&gt;. | [optional] 
**EnforceRunaiScheduler** | Pointer to **NullableBool** | When set to true, enforces that all pods in this project&#39;s namespace are scheduled using the NVIDIA Run:ai Scheduler (runai-scheduler). Supported for cluster version 2.21 and later. | [optional] [default to true]
**ParentId** | Pointer to **NullableString** | The ID of the parent department. When specified, the project is created as a child of this department and inherits its quota, the quota of other projects within the department and its scheduling rules. | [optional] 
**Resources** | [**[]ResourcesNullable**](ResourcesNullable.md) | An array of resource configurations for the project. Each entry defines GPU, CPU, and CPU memory assigned to a specific node pool. | 

## Methods

### NewProjectCreationRequest

`func NewProjectCreationRequest(name string, clusterId string, resources []ResourcesNullable, ) *ProjectCreationRequest`

NewProjectCreationRequest instantiates a new ProjectCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectCreationRequestWithDefaults

`func NewProjectCreationRequestWithDefaults() *ProjectCreationRequest`

NewProjectCreationRequestWithDefaults instantiates a new ProjectCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *ProjectCreationRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ProjectCreationRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ProjectCreationRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ProjectCreationRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *ProjectCreationRequest) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *ProjectCreationRequest) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *ProjectCreationRequest) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *ProjectCreationRequest) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *ProjectCreationRequest) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *ProjectCreationRequest) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *ProjectCreationRequest) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectCreationRequest) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectCreationRequest) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectCreationRequest) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *ProjectCreationRequest) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *ProjectCreationRequest) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *ProjectCreationRequest) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *ProjectCreationRequest) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *ProjectCreationRequest) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *ProjectCreationRequest) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetName

`func (o *ProjectCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *ProjectCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ProjectCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ProjectCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetRequestedNamespace

`func (o *ProjectCreationRequest) GetRequestedNamespace() string`

GetRequestedNamespace returns the RequestedNamespace field if non-nil, zero value otherwise.

### GetRequestedNamespaceOk

`func (o *ProjectCreationRequest) GetRequestedNamespaceOk() (*string, bool)`

GetRequestedNamespaceOk returns a tuple with the RequestedNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNamespace

`func (o *ProjectCreationRequest) SetRequestedNamespace(v string)`

SetRequestedNamespace sets RequestedNamespace field to given value.

### HasRequestedNamespace

`func (o *ProjectCreationRequest) HasRequestedNamespace() bool`

HasRequestedNamespace returns a boolean if a field has been set.

### GetEnforceRunaiScheduler

`func (o *ProjectCreationRequest) GetEnforceRunaiScheduler() bool`

GetEnforceRunaiScheduler returns the EnforceRunaiScheduler field if non-nil, zero value otherwise.

### GetEnforceRunaiSchedulerOk

`func (o *ProjectCreationRequest) GetEnforceRunaiSchedulerOk() (*bool, bool)`

GetEnforceRunaiSchedulerOk returns a tuple with the EnforceRunaiScheduler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforceRunaiScheduler

`func (o *ProjectCreationRequest) SetEnforceRunaiScheduler(v bool)`

SetEnforceRunaiScheduler sets EnforceRunaiScheduler field to given value.

### HasEnforceRunaiScheduler

`func (o *ProjectCreationRequest) HasEnforceRunaiScheduler() bool`

HasEnforceRunaiScheduler returns a boolean if a field has been set.

### SetEnforceRunaiSchedulerNil

`func (o *ProjectCreationRequest) SetEnforceRunaiSchedulerNil(b bool)`

 SetEnforceRunaiSchedulerNil sets the value for EnforceRunaiScheduler to be an explicit nil

### UnsetEnforceRunaiScheduler
`func (o *ProjectCreationRequest) UnsetEnforceRunaiScheduler()`

UnsetEnforceRunaiScheduler ensures that no value is present for EnforceRunaiScheduler, not even an explicit nil
### GetParentId

`func (o *ProjectCreationRequest) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *ProjectCreationRequest) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *ProjectCreationRequest) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *ProjectCreationRequest) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *ProjectCreationRequest) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *ProjectCreationRequest) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil
### GetResources

`func (o *ProjectCreationRequest) GetResources() []ResourcesNullable`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectCreationRequest) GetResourcesOk() (*[]ResourcesNullable, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectCreationRequest) SetResources(v []ResourcesNullable)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


