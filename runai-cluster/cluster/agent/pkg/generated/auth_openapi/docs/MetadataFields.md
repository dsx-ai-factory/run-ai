# MetadataFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**TenantId** | Pointer to **int32** |  | [optional] 
**CreatedBy** | **string** |  | 

## Methods

### NewMetadataFields

`func NewMetadataFields(id int32, createdAt time.Time, updatedAt time.Time, createdBy string, ) *MetadataFields`

NewMetadataFields instantiates a new MetadataFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetadataFieldsWithDefaults

`func NewMetadataFieldsWithDefaults() *MetadataFields`

NewMetadataFieldsWithDefaults instantiates a new MetadataFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *MetadataFields) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *MetadataFields) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *MetadataFields) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *MetadataFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *MetadataFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *MetadataFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *MetadataFields) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *MetadataFields) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *MetadataFields) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *MetadataFields) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *MetadataFields) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *MetadataFields) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *MetadataFields) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *MetadataFields) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *MetadataFields) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *MetadataFields) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *MetadataFields) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *MetadataFields) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *MetadataFields) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetCreatedBy

`func (o *MetadataFields) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *MetadataFields) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *MetadataFields) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


