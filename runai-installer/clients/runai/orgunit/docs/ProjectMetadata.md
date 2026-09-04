# ProjectMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**Namespace** | Pointer to **NullableString** |  | [optional] 
**EnforceRunaiScheduler** | Pointer to **bool** | a flag to enforce that all pods in this namespace will be scheduled with runai-scheduler. Supported for cluster v2.21+ | [optional] 
**Phase** | **NullableString** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**CreatedAt** | **time.Time** |  | 
**Allocated** | [**QuotaStatusResource**](QuotaStatusResource.md) |  | 
**Deserved** | [**QuotaStatusResource**](QuotaStatusResource.md) |  | 

## Methods

### NewProjectMetadata

`func NewProjectMetadata(id string, name string, phase NullableString, clusterId string, createdAt time.Time, allocated QuotaStatusResource, deserved QuotaStatusResource, ) *ProjectMetadata`

NewProjectMetadata instantiates a new ProjectMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectMetadataWithDefaults

`func NewProjectMetadataWithDefaults() *ProjectMetadata`

NewProjectMetadataWithDefaults instantiates a new ProjectMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectMetadata) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectMetadata) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectMetadata) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *ProjectMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectMetadata) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *ProjectMetadata) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectMetadata) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectMetadata) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectMetadata) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### SetNamespaceNil

`func (o *ProjectMetadata) SetNamespaceNil(b bool)`

 SetNamespaceNil sets the value for Namespace to be an explicit nil

### UnsetNamespace
`func (o *ProjectMetadata) UnsetNamespace()`

UnsetNamespace ensures that no value is present for Namespace, not even an explicit nil
### GetEnforceRunaiScheduler

`func (o *ProjectMetadata) GetEnforceRunaiScheduler() bool`

GetEnforceRunaiScheduler returns the EnforceRunaiScheduler field if non-nil, zero value otherwise.

### GetEnforceRunaiSchedulerOk

`func (o *ProjectMetadata) GetEnforceRunaiSchedulerOk() (*bool, bool)`

GetEnforceRunaiSchedulerOk returns a tuple with the EnforceRunaiScheduler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforceRunaiScheduler

`func (o *ProjectMetadata) SetEnforceRunaiScheduler(v bool)`

SetEnforceRunaiScheduler sets EnforceRunaiScheduler field to given value.

### HasEnforceRunaiScheduler

`func (o *ProjectMetadata) HasEnforceRunaiScheduler() bool`

HasEnforceRunaiScheduler returns a boolean if a field has been set.

### GetPhase

`func (o *ProjectMetadata) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ProjectMetadata) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ProjectMetadata) SetPhase(v string)`

SetPhase sets Phase field to given value.


### SetPhaseNil

`func (o *ProjectMetadata) SetPhaseNil(b bool)`

 SetPhaseNil sets the value for Phase to be an explicit nil

### UnsetPhase
`func (o *ProjectMetadata) UnsetPhase()`

UnsetPhase ensures that no value is present for Phase, not even an explicit nil
### GetClusterId

`func (o *ProjectMetadata) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ProjectMetadata) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ProjectMetadata) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetCreatedAt

`func (o *ProjectMetadata) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ProjectMetadata) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ProjectMetadata) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetAllocated

`func (o *ProjectMetadata) GetAllocated() QuotaStatusResource`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *ProjectMetadata) GetAllocatedOk() (*QuotaStatusResource, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *ProjectMetadata) SetAllocated(v QuotaStatusResource)`

SetAllocated sets Allocated field to given value.


### GetDeserved

`func (o *ProjectMetadata) GetDeserved() QuotaStatusResource`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *ProjectMetadata) GetDeservedOk() (*QuotaStatusResource, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *ProjectMetadata) SetDeserved(v QuotaStatusResource)`

SetDeserved sets Deserved field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


