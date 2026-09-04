# AssetCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the asset. | 
**Description** | Pointer to **NullableString** | The description of the asset. | [optional] 
**Scope** | [**Scope**](Scope.md) |  | 
**ClusterId** | Pointer to **NullableString** | The id of the cluster. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department. Must be specified for department scoped assets. | [optional] 
**ProjectId** | Pointer to **NullableInt32** | The id of the project. Must be specified for project scoped assets. | [optional] 
**AutoDelete** | Pointer to **NullableBool** | The asset will be deleted automatically. This is intended for internal use. | [optional] [default to false]
**WorkloadSupportedTypes** | Pointer to [**NullableWorkloadSupportedTypes**](WorkloadSupportedTypes.md) |  | [optional] 

## Methods

### NewAssetCreationRequest

`func NewAssetCreationRequest(name string, scope Scope, ) *AssetCreationRequest`

NewAssetCreationRequest instantiates a new AssetCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetCreationRequestWithDefaults

`func NewAssetCreationRequestWithDefaults() *AssetCreationRequest`

NewAssetCreationRequestWithDefaults instantiates a new AssetCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AssetCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *AssetCreationRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *AssetCreationRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *AssetCreationRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *AssetCreationRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *AssetCreationRequest) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *AssetCreationRequest) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetScope

`func (o *AssetCreationRequest) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AssetCreationRequest) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AssetCreationRequest) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *AssetCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AssetCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AssetCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AssetCreationRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *AssetCreationRequest) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *AssetCreationRequest) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *AssetCreationRequest) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *AssetCreationRequest) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *AssetCreationRequest) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *AssetCreationRequest) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *AssetCreationRequest) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *AssetCreationRequest) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *AssetCreationRequest) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AssetCreationRequest) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AssetCreationRequest) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AssetCreationRequest) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *AssetCreationRequest) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *AssetCreationRequest) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetAutoDelete

`func (o *AssetCreationRequest) GetAutoDelete() bool`

GetAutoDelete returns the AutoDelete field if non-nil, zero value otherwise.

### GetAutoDeleteOk

`func (o *AssetCreationRequest) GetAutoDeleteOk() (*bool, bool)`

GetAutoDeleteOk returns a tuple with the AutoDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDelete

`func (o *AssetCreationRequest) SetAutoDelete(v bool)`

SetAutoDelete sets AutoDelete field to given value.

### HasAutoDelete

`func (o *AssetCreationRequest) HasAutoDelete() bool`

HasAutoDelete returns a boolean if a field has been set.

### SetAutoDeleteNil

`func (o *AssetCreationRequest) SetAutoDeleteNil(b bool)`

 SetAutoDeleteNil sets the value for AutoDelete to be an explicit nil

### UnsetAutoDelete
`func (o *AssetCreationRequest) UnsetAutoDelete()`

UnsetAutoDelete ensures that no value is present for AutoDelete, not even an explicit nil
### GetWorkloadSupportedTypes

`func (o *AssetCreationRequest) GetWorkloadSupportedTypes() WorkloadSupportedTypes`

GetWorkloadSupportedTypes returns the WorkloadSupportedTypes field if non-nil, zero value otherwise.

### GetWorkloadSupportedTypesOk

`func (o *AssetCreationRequest) GetWorkloadSupportedTypesOk() (*WorkloadSupportedTypes, bool)`

GetWorkloadSupportedTypesOk returns a tuple with the WorkloadSupportedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadSupportedTypes

`func (o *AssetCreationRequest) SetWorkloadSupportedTypes(v WorkloadSupportedTypes)`

SetWorkloadSupportedTypes sets WorkloadSupportedTypes field to given value.

### HasWorkloadSupportedTypes

`func (o *AssetCreationRequest) HasWorkloadSupportedTypes() bool`

HasWorkloadSupportedTypes returns a boolean if a field has been set.

### SetWorkloadSupportedTypesNil

`func (o *AssetCreationRequest) SetWorkloadSupportedTypesNil(b bool)`

 SetWorkloadSupportedTypesNil sets the value for WorkloadSupportedTypes to be an explicit nil

### UnsetWorkloadSupportedTypes
`func (o *AssetCreationRequest) UnsetWorkloadSupportedTypes()`

UnsetWorkloadSupportedTypes ensures that no value is present for WorkloadSupportedTypes, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


