# AssetMetaForSync

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
**Name** | **string** | The name of the asset. | 
**Description** | Pointer to **NullableString** | The description of the asset. | [optional] 
**Id** | **string** | Unique identifier of the asset. | 
**Kind** | [**AssetKind**](AssetKind.md) |  | 
**PrevName** | Pointer to **string** | The name of the asset. | [optional] 

## Methods

### NewAssetMetaForSync

`func NewAssetMetaForSync(scope Scope, name string, id string, kind AssetKind, ) *AssetMetaForSync`

NewAssetMetaForSync instantiates a new AssetMetaForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetMetaForSyncWithDefaults

`func NewAssetMetaForSyncWithDefaults() *AssetMetaForSync`

NewAssetMetaForSyncWithDefaults instantiates a new AssetMetaForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetScope

`func (o *AssetMetaForSync) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AssetMetaForSync) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AssetMetaForSync) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *AssetMetaForSync) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AssetMetaForSync) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AssetMetaForSync) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AssetMetaForSync) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *AssetMetaForSync) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *AssetMetaForSync) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *AssetMetaForSync) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *AssetMetaForSync) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *AssetMetaForSync) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *AssetMetaForSync) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *AssetMetaForSync) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *AssetMetaForSync) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *AssetMetaForSync) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AssetMetaForSync) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AssetMetaForSync) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AssetMetaForSync) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *AssetMetaForSync) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *AssetMetaForSync) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetAutoDelete

`func (o *AssetMetaForSync) GetAutoDelete() bool`

GetAutoDelete returns the AutoDelete field if non-nil, zero value otherwise.

### GetAutoDeleteOk

`func (o *AssetMetaForSync) GetAutoDeleteOk() (*bool, bool)`

GetAutoDeleteOk returns a tuple with the AutoDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDelete

`func (o *AssetMetaForSync) SetAutoDelete(v bool)`

SetAutoDelete sets AutoDelete field to given value.

### HasAutoDelete

`func (o *AssetMetaForSync) HasAutoDelete() bool`

HasAutoDelete returns a boolean if a field has been set.

### SetAutoDeleteNil

`func (o *AssetMetaForSync) SetAutoDeleteNil(b bool)`

 SetAutoDeleteNil sets the value for AutoDelete to be an explicit nil

### UnsetAutoDelete
`func (o *AssetMetaForSync) UnsetAutoDelete()`

UnsetAutoDelete ensures that no value is present for AutoDelete, not even an explicit nil
### GetWorkloadSupportedTypes

`func (o *AssetMetaForSync) GetWorkloadSupportedTypes() WorkloadSupportedTypes`

GetWorkloadSupportedTypes returns the WorkloadSupportedTypes field if non-nil, zero value otherwise.

### GetWorkloadSupportedTypesOk

`func (o *AssetMetaForSync) GetWorkloadSupportedTypesOk() (*WorkloadSupportedTypes, bool)`

GetWorkloadSupportedTypesOk returns a tuple with the WorkloadSupportedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadSupportedTypes

`func (o *AssetMetaForSync) SetWorkloadSupportedTypes(v WorkloadSupportedTypes)`

SetWorkloadSupportedTypes sets WorkloadSupportedTypes field to given value.

### HasWorkloadSupportedTypes

`func (o *AssetMetaForSync) HasWorkloadSupportedTypes() bool`

HasWorkloadSupportedTypes returns a boolean if a field has been set.

### SetWorkloadSupportedTypesNil

`func (o *AssetMetaForSync) SetWorkloadSupportedTypesNil(b bool)`

 SetWorkloadSupportedTypesNil sets the value for WorkloadSupportedTypes to be an explicit nil

### UnsetWorkloadSupportedTypes
`func (o *AssetMetaForSync) UnsetWorkloadSupportedTypes()`

UnsetWorkloadSupportedTypes ensures that no value is present for WorkloadSupportedTypes, not even an explicit nil
### GetPrivate

`func (o *AssetMetaForSync) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *AssetMetaForSync) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *AssetMetaForSync) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *AssetMetaForSync) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### SetPrivateNil

`func (o *AssetMetaForSync) SetPrivateNil(b bool)`

 SetPrivateNil sets the value for Private to be an explicit nil

### UnsetPrivate
`func (o *AssetMetaForSync) UnsetPrivate()`

UnsetPrivate ensures that no value is present for Private, not even an explicit nil
### GetName

`func (o *AssetMetaForSync) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetMetaForSync) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetMetaForSync) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *AssetMetaForSync) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *AssetMetaForSync) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *AssetMetaForSync) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *AssetMetaForSync) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *AssetMetaForSync) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *AssetMetaForSync) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetId

`func (o *AssetMetaForSync) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetMetaForSync) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetMetaForSync) SetId(v string)`

SetId sets Id field to given value.


### GetKind

`func (o *AssetMetaForSync) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AssetMetaForSync) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AssetMetaForSync) SetKind(v AssetKind)`

SetKind sets Kind field to given value.


### GetPrevName

`func (o *AssetMetaForSync) GetPrevName() string`

GetPrevName returns the PrevName field if non-nil, zero value otherwise.

### GetPrevNameOk

`func (o *AssetMetaForSync) GetPrevNameOk() (*string, bool)`

GetPrevNameOk returns a tuple with the PrevName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrevName

`func (o *AssetMetaForSync) SetPrevName(v string)`

SetPrevName sets PrevName field to given value.

### HasPrevName

`func (o *AssetMetaForSync) HasPrevName() bool`

HasPrevName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


