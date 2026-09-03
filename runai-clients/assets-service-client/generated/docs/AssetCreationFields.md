# AssetCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Scope** | [**Scope**](Scope.md) |  | 
**ClusterId** | Pointer to **NullableString** | The id of the cluster. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department. Must be specified for department scoped assets. | [optional] 
**ProjectId** | Pointer to **NullableInt32** | The id of the project. Must be specified for project scoped assets. | [optional] 
**AutoDelete** | Pointer to **NullableBool** | The asset will be deleted automatically. This is intended for internal use. | [optional] [default to false]
**WorkloadSupportedTypes** | Pointer to [**NullableWorkloadSupportedTypes**](WorkloadSupportedTypes.md) |  | [optional] 
**Private** | Pointer to **NullableBool** | The asset will be private to the user who created it. supported only for docker registry and generic secrets | [optional] [default to false]

## Methods

### NewAssetCreationFields

`func NewAssetCreationFields(scope Scope, ) *AssetCreationFields`

NewAssetCreationFields instantiates a new AssetCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetCreationFieldsWithDefaults

`func NewAssetCreationFieldsWithDefaults() *AssetCreationFields`

NewAssetCreationFieldsWithDefaults instantiates a new AssetCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *AssetCreationFields) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AssetCreationFields) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AssetCreationFields) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *AssetCreationFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AssetCreationFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AssetCreationFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AssetCreationFields) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *AssetCreationFields) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *AssetCreationFields) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *AssetCreationFields) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *AssetCreationFields) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *AssetCreationFields) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *AssetCreationFields) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *AssetCreationFields) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *AssetCreationFields) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *AssetCreationFields) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AssetCreationFields) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AssetCreationFields) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AssetCreationFields) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *AssetCreationFields) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *AssetCreationFields) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetAutoDelete

`func (o *AssetCreationFields) GetAutoDelete() bool`

GetAutoDelete returns the AutoDelete field if non-nil, zero value otherwise.

### GetAutoDeleteOk

`func (o *AssetCreationFields) GetAutoDeleteOk() (*bool, bool)`

GetAutoDeleteOk returns a tuple with the AutoDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDelete

`func (o *AssetCreationFields) SetAutoDelete(v bool)`

SetAutoDelete sets AutoDelete field to given value.

### HasAutoDelete

`func (o *AssetCreationFields) HasAutoDelete() bool`

HasAutoDelete returns a boolean if a field has been set.

### SetAutoDeleteNil

`func (o *AssetCreationFields) SetAutoDeleteNil(b bool)`

 SetAutoDeleteNil sets the value for AutoDelete to be an explicit nil

### UnsetAutoDelete
`func (o *AssetCreationFields) UnsetAutoDelete()`

UnsetAutoDelete ensures that no value is present for AutoDelete, not even an explicit nil
### GetWorkloadSupportedTypes

`func (o *AssetCreationFields) GetWorkloadSupportedTypes() WorkloadSupportedTypes`

GetWorkloadSupportedTypes returns the WorkloadSupportedTypes field if non-nil, zero value otherwise.

### GetWorkloadSupportedTypesOk

`func (o *AssetCreationFields) GetWorkloadSupportedTypesOk() (*WorkloadSupportedTypes, bool)`

GetWorkloadSupportedTypesOk returns a tuple with the WorkloadSupportedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadSupportedTypes

`func (o *AssetCreationFields) SetWorkloadSupportedTypes(v WorkloadSupportedTypes)`

SetWorkloadSupportedTypes sets WorkloadSupportedTypes field to given value.

### HasWorkloadSupportedTypes

`func (o *AssetCreationFields) HasWorkloadSupportedTypes() bool`

HasWorkloadSupportedTypes returns a boolean if a field has been set.

### SetWorkloadSupportedTypesNil

`func (o *AssetCreationFields) SetWorkloadSupportedTypesNil(b bool)`

 SetWorkloadSupportedTypesNil sets the value for WorkloadSupportedTypes to be an explicit nil

### UnsetWorkloadSupportedTypes
`func (o *AssetCreationFields) UnsetWorkloadSupportedTypes()`

UnsetWorkloadSupportedTypes ensures that no value is present for WorkloadSupportedTypes, not even an explicit nil
### GetPrivate

`func (o *AssetCreationFields) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *AssetCreationFields) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *AssetCreationFields) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *AssetCreationFields) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### SetPrivateNil

`func (o *AssetCreationFields) SetPrivateNil(b bool)`

 SetPrivateNil sets the value for Private to be an explicit nil

### UnsetPrivate
`func (o *AssetCreationFields) UnsetPrivate()`

UnsetPrivate ensures that no value is present for Private, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


