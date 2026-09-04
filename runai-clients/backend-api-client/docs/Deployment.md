# Deployment

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterUuid** | **string** |  | 
**Id** | **string** |  | 
**Name** | **string** |  | 
**Namespace** | **string** |  | 
**Project** | Pointer to **string** |  | [optional] 
**Owners** | [**[]Owner**](Owner.md) |  | 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**Replicas** | **int32** |  | 
**Template** | [**PodTemplate**](PodTemplate.md) |  | 
**SchedulerName** | **string** |  | 
**Username** | Pointer to **string** |  | [optional] 
**Connections** | **[]string** |  | 
**ServiceId** | Pointer to **string** |  | [optional] 
**RevisionId** | Pointer to **string** |  | [optional] 
**InferenceworkloadName** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to [**[]DeploymentCondition**](DeploymentCondition.md) |  | [optional] 
**NodePool** | Pointer to **string** |  | [optional] 

## Methods

### NewDeployment

`func NewDeployment(clusterUuid string, id string, name string, namespace string, owners []Owner, replicas int32, template PodTemplate, schedulerName string, connections []string, ) *Deployment`

NewDeployment instantiates a new Deployment object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeploymentWithDefaults

`func NewDeploymentWithDefaults() *Deployment`

NewDeploymentWithDefaults instantiates a new Deployment object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterUuid

`func (o *Deployment) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Deployment) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Deployment) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetId

`func (o *Deployment) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Deployment) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Deployment) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *Deployment) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Deployment) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Deployment) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *Deployment) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Deployment) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Deployment) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetProject

`func (o *Deployment) GetProject() string`

GetProject returns the Project field if non-nil, zero value otherwise.

### GetProjectOk

`func (o *Deployment) GetProjectOk() (*string, bool)`

GetProjectOk returns a tuple with the Project field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProject

`func (o *Deployment) SetProject(v string)`

SetProject sets Project field to given value.

### HasProject

`func (o *Deployment) HasProject() bool`

HasProject returns a boolean if a field has been set.

### GetOwners

`func (o *Deployment) GetOwners() []Owner`

GetOwners returns the Owners field if non-nil, zero value otherwise.

### GetOwnersOk

`func (o *Deployment) GetOwnersOk() (*[]Owner, bool)`

GetOwnersOk returns a tuple with the Owners field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwners

`func (o *Deployment) SetOwners(v []Owner)`

SetOwners sets Owners field to given value.


### GetCreatedAt

`func (o *Deployment) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Deployment) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Deployment) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Deployment) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetReplicas

`func (o *Deployment) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *Deployment) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *Deployment) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.


### GetTemplate

`func (o *Deployment) GetTemplate() PodTemplate`

GetTemplate returns the Template field if non-nil, zero value otherwise.

### GetTemplateOk

`func (o *Deployment) GetTemplateOk() (*PodTemplate, bool)`

GetTemplateOk returns a tuple with the Template field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplate

`func (o *Deployment) SetTemplate(v PodTemplate)`

SetTemplate sets Template field to given value.


### GetSchedulerName

`func (o *Deployment) GetSchedulerName() string`

GetSchedulerName returns the SchedulerName field if non-nil, zero value otherwise.

### GetSchedulerNameOk

`func (o *Deployment) GetSchedulerNameOk() (*string, bool)`

GetSchedulerNameOk returns a tuple with the SchedulerName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulerName

`func (o *Deployment) SetSchedulerName(v string)`

SetSchedulerName sets SchedulerName field to given value.


### GetUsername

`func (o *Deployment) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *Deployment) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *Deployment) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *Deployment) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetConnections

`func (o *Deployment) GetConnections() []string`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *Deployment) GetConnectionsOk() (*[]string, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *Deployment) SetConnections(v []string)`

SetConnections sets Connections field to given value.


### GetServiceId

`func (o *Deployment) GetServiceId() string`

GetServiceId returns the ServiceId field if non-nil, zero value otherwise.

### GetServiceIdOk

`func (o *Deployment) GetServiceIdOk() (*string, bool)`

GetServiceIdOk returns a tuple with the ServiceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceId

`func (o *Deployment) SetServiceId(v string)`

SetServiceId sets ServiceId field to given value.

### HasServiceId

`func (o *Deployment) HasServiceId() bool`

HasServiceId returns a boolean if a field has been set.

### GetRevisionId

`func (o *Deployment) GetRevisionId() string`

GetRevisionId returns the RevisionId field if non-nil, zero value otherwise.

### GetRevisionIdOk

`func (o *Deployment) GetRevisionIdOk() (*string, bool)`

GetRevisionIdOk returns a tuple with the RevisionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevisionId

`func (o *Deployment) SetRevisionId(v string)`

SetRevisionId sets RevisionId field to given value.

### HasRevisionId

`func (o *Deployment) HasRevisionId() bool`

HasRevisionId returns a boolean if a field has been set.

### GetInferenceworkloadName

`func (o *Deployment) GetInferenceworkloadName() string`

GetInferenceworkloadName returns the InferenceworkloadName field if non-nil, zero value otherwise.

### GetInferenceworkloadNameOk

`func (o *Deployment) GetInferenceworkloadNameOk() (*string, bool)`

GetInferenceworkloadNameOk returns a tuple with the InferenceworkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInferenceworkloadName

`func (o *Deployment) SetInferenceworkloadName(v string)`

SetInferenceworkloadName sets InferenceworkloadName field to given value.

### HasInferenceworkloadName

`func (o *Deployment) HasInferenceworkloadName() bool`

HasInferenceworkloadName returns a boolean if a field has been set.

### GetConditions

`func (o *Deployment) GetConditions() []DeploymentCondition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *Deployment) GetConditionsOk() (*[]DeploymentCondition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *Deployment) SetConditions(v []DeploymentCondition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *Deployment) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetNodePool

`func (o *Deployment) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *Deployment) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *Deployment) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *Deployment) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


