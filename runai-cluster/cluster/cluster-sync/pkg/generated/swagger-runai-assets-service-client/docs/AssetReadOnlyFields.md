# AssetReadOnlyFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the asset. | 
**Kind** | [**AssetKind**](AssetKind.md) |  | 
**TenantId** | Pointer to **NullableInt32** | The id of the tenant. | [optional] 
**CreatedBy** | **string** | The user who created the asset. | 
**CreatedAt** | **time.Time** | The time at which the asset were created | 
**UpdatedBy** | **string** | The user who updated the asset. | 
**UpdatedAt** | **time.Time** | The time at which the asset has been updated | 
**ProjectName** | Pointer to **NullableString** | The name of the project that the asset is associated with, for project scoped assets. | [optional] 

## Methods

### NewAssetReadOnlyFields

`func NewAssetReadOnlyFields(id string, kind AssetKind, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time, ) *AssetReadOnlyFields`

NewAssetReadOnlyFields instantiates a new AssetReadOnlyFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetReadOnlyFieldsWithDefaults

`func NewAssetReadOnlyFieldsWithDefaults() *AssetReadOnlyFields`

NewAssetReadOnlyFieldsWithDefaults instantiates a new AssetReadOnlyFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssetReadOnlyFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetReadOnlyFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetReadOnlyFields) SetId(v string)`

SetId sets Id field to given value.


### GetKind

`func (o *AssetReadOnlyFields) GetKind() AssetKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *AssetReadOnlyFields) GetKindOk() (*AssetKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *AssetReadOnlyFields) SetKind(v AssetKind)`

SetKind sets Kind field to given value.


### GetTenantId

`func (o *AssetReadOnlyFields) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AssetReadOnlyFields) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AssetReadOnlyFields) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *AssetReadOnlyFields) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### SetTenantIdNil

`func (o *AssetReadOnlyFields) SetTenantIdNil(b bool)`

 SetTenantIdNil sets the value for TenantId to be an explicit nil

### UnsetTenantId
`func (o *AssetReadOnlyFields) UnsetTenantId()`

UnsetTenantId ensures that no value is present for TenantId, not even an explicit nil
### GetCreatedBy

`func (o *AssetReadOnlyFields) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AssetReadOnlyFields) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AssetReadOnlyFields) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *AssetReadOnlyFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AssetReadOnlyFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AssetReadOnlyFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedBy

`func (o *AssetReadOnlyFields) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *AssetReadOnlyFields) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *AssetReadOnlyFields) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetUpdatedAt

`func (o *AssetReadOnlyFields) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AssetReadOnlyFields) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AssetReadOnlyFields) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetProjectName

`func (o *AssetReadOnlyFields) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *AssetReadOnlyFields) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *AssetReadOnlyFields) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *AssetReadOnlyFields) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### SetProjectNameNil

`func (o *AssetReadOnlyFields) SetProjectNameNil(b bool)`

 SetProjectNameNil sets the value for ProjectName to be an explicit nil

### UnsetProjectName
`func (o *AssetReadOnlyFields) UnsetProjectName()`

UnsetProjectName ensures that no value is present for ProjectName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


