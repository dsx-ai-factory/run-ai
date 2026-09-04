# AssetMeta

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
**Private** | Pointer to **NullableBool** | The asset will be private to the user who created it. supported only for docker registry and generic secrets | [optional] [default to false]
**Id** | **string** | Unique identifier of the asset. | 
**Kind** | [**AssetKind**](AssetKind.md) |  | 
**TenantId** | Pointer to **NullableInt32** | The id of the tenant. | [optional] 
**CreatedBy** | **string** | The user who created the asset. | 
**CreatedAt** | **time.Time** | The time at which the asset were created | 
**UpdatedBy** | **string** | The user who updated the asset. | 
**UpdatedAt** | **time.Time** | The time at which the asset has been updated | 
**DeletedAt** | Pointer to **time.Time** | Deletion time of the asset (relevant for assets that sync to the cluster) | [optional] 
**DeletedBy** | Pointer to **string** | The user who deleted the asset | [optional] 
**ProjectName** | Pointer to **NullableString** | The name of the project that the asset is associated with, for project scoped assets. | [optional] 
**UpdateCount** | Pointer to **int64** | the value of an update counter of a data in the system. | [optional] 

## Methods

### NewAssetMeta

`func NewAssetMeta(name string, scope Scope, id string, kind AssetKind, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time, ) *AssetMeta`

NewAssetMeta instantiates a new AssetMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetMetaWithDefaults

`func NewAssetMetaWithDefaults() *AssetMeta`

NewAssetMetaWithDefaults instantiates a new AssetMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AssetMeta) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetMeta) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetMeta) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *AssetMeta) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *AssetMeta) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *AssetMeta) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *AssetMeta) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *AssetMeta) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *AssetMeta) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetScope

`func (o *AssetMeta) GetScope() Scope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *AssetMeta) GetScopeOk() (*Scope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *AssetMeta) SetScope(v Scope)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *AssetMeta) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AssetMeta) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AssetMeta) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AssetMeta) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *AssetMeta) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *AssetMeta) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *AssetMeta) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *AssetMeta) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *AssetMeta) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *AssetMeta) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *AssetMeta) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *AssetMeta) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *AssetMeta) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *AssetMeta) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *AssetMeta) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *AssetMeta) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *AssetMeta) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *AssetMeta) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil
### GetAutoDelete

`func (o *AssetMeta) GetAutoDelete() bool`

GetAutoDelete returns the AutoDelete field if non-nil, zero value otherwise.

### GetAutoDeleteOk

`func (o *AssetMeta) GetAutoDeleteOk() (*bool, bool)`

GetAutoDeleteOk returns a tuple with the AutoDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDelete

`func (o *AssetMeta) SetAutoDelete(v bool)`

SetAutoDelete sets AutoDelete field to given value.

### HasAutoDelete

`func (o *AssetMeta) HasAutoDelete() bool`

HasAutoDelete returns a boolean if a field has been set.

### SetAutoDeleteNil

`func (o *AssetMeta) SetAutoDeleteNil(b bool)`

 SetAutoDeleteNil sets the value for AutoDelete to be an explicit nil

### UnsetAutoDelete
`func (o *AssetMeta) UnsetAutoDelete()`

UnsetAutoDelete ensures that no value is present for AutoDelete, not even an explicit nil
### GetWorkloadSupportedTypes

`func (o *AssetMeta) GetWorkloadSupportedTypes() WorkloadSupportedTypes`

GetWorkloadSupportedTypes returns the WorkloadSupportedTypes field if non-nil, zero value otherwise.

### GetWorkloadSupportedTypesOk

`func (o *AssetMeta) GetWorkloadSupportedTypesOk() (*WorkloadSupportedTypes, bool)`

GetWorkloadSupportedTypesOk returns a tuple with the WorkloadSupportedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadSupportedTypes

`func (o *AssetMeta) SetWorkloadSupportedTypes(v WorkloadSupportedTypes)`

SetWorkloadSupportedTypes sets WorkloadSupportedTypes field to given value.

### HasWorkloadSupportedTypes

`func (o *AssetMeta) HasWorkloadSupportedTypes() bool`

HasWorkloadSupportedTypes returns a boolean if a field has been set.

### SetWorkloadSupportedTypesNil

`func (o *AssetMeta) SetWorkloadSupportedTypesNil(b bool)`

 SetWorkloadSupportedTypesNil sets the value for WorkloadSupportedTypes to be an explicit nil

### UnsetWorkloadSupportedTypes
`func (o *AssetMeta) UnsetWorkloadSupportedTypes()`

UnsetWorkloadSupportedTypes ensures that no value is present for WorkloadSupportedTypes, not even an explicit nil
### GetPrivate

`func (o *AssetMeta) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *AssetMeta) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *AssetMeta) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *AssetMeta) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### SetPrivateNil

`func (o *AssetMeta) SetPrivateNil(b bool)`

 SetPrivateNil sets the value for Private to be an explicit nil

### UnsetPrivate
`func (o *AssetMeta) UnsetPrivate()`

UnsetPrivate ensures that no value is present for Private, not even an explicit nil
### GetId

`func (o *AssetMeta) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetMeta) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetMeta) SetId(v string)`

SetId sets Id field to given value.


### GetKind

`func (o *AssetMeta) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AssetMeta) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AssetMeta) SetKind(v AssetKind)`

SetKind sets Kind field to given value.


### GetTenantId

`func (o *AssetMeta) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AssetMeta) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AssetMeta) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *AssetMeta) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### SetTenantIdNil

`func (o *AssetMeta) SetTenantIdNil(b bool)`

 SetTenantIdNil sets the value for TenantId to be an explicit nil

### UnsetTenantId
`func (o *AssetMeta) UnsetTenantId()`

UnsetTenantId ensures that no value is present for TenantId, not even an explicit nil
### GetCreatedBy

`func (o *AssetMeta) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AssetMeta) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AssetMeta) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *AssetMeta) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AssetMeta) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AssetMeta) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedBy

`func (o *AssetMeta) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *AssetMeta) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *AssetMeta) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetUpdatedAt

`func (o *AssetMeta) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AssetMeta) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AssetMeta) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *AssetMeta) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *AssetMeta) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *AssetMeta) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *AssetMeta) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### GetDeletedBy

`func (o *AssetMeta) GetDeletedBy() string`

GetDeletedBy returns the DeletedBy field if non-nil, zero value otherwise.

### GetDeletedByOk

`func (o *AssetMeta) GetDeletedByOk() (*string, bool)`

GetDeletedByOk returns a tuple with the DeletedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedBy

`func (o *AssetMeta) SetDeletedBy(v string)`

SetDeletedBy sets DeletedBy field to given value.

### HasDeletedBy

`func (o *AssetMeta) HasDeletedBy() bool`

HasDeletedBy returns a boolean if a field has been set.

### GetProjectName

`func (o *AssetMeta) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *AssetMeta) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *AssetMeta) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *AssetMeta) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### SetProjectNameNil

`func (o *AssetMeta) SetProjectNameNil(b bool)`

 SetProjectNameNil sets the value for ProjectName to be an explicit nil

### UnsetProjectName
`func (o *AssetMeta) UnsetProjectName()`

UnsetProjectName ensures that no value is present for ProjectName, not even an explicit nil
### GetUpdateCount

`func (o *AssetMeta) GetUpdateCount() int64`

GetUpdateCount returns the UpdateCount field if non-nil, zero value otherwise.

### GetUpdateCountOk

`func (o *AssetMeta) GetUpdateCountOk() (*int64, bool)`

GetUpdateCountOk returns a tuple with the UpdateCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCount

`func (o *AssetMeta) SetUpdateCount(v int64)`

SetUpdateCount sets UpdateCount field to given value.

### HasUpdateCount

`func (o *AssetMeta) HasUpdateCount() bool`

HasUpdateCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


