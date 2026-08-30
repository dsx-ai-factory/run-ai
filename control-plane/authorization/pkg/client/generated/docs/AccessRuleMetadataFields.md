# AccessRuleMetadataFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** | The ID of the access rule entry. | 
**CreatedAt** | **time.Time** | The timestamp for when the access rule was created. | 
**UpdatedAt** | **time.Time** | The timestamp for the last time the access rule was updated. | 
**DeletedAt** | Pointer to **NullableTime** | The timestamp indicating when the access rule was deleted. | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**CreatedBy** | **string** | Identifier of the user who created the access rule. | 

## Methods

### NewAccessRuleMetadataFields

`func NewAccessRuleMetadataFields(id int32, createdAt time.Time, updatedAt time.Time, createdBy string, ) *AccessRuleMetadataFields`

NewAccessRuleMetadataFields instantiates a new AccessRuleMetadataFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleMetadataFieldsWithDefaults

`func NewAccessRuleMetadataFieldsWithDefaults() *AccessRuleMetadataFields`

NewAccessRuleMetadataFieldsWithDefaults instantiates a new AccessRuleMetadataFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AccessRuleMetadataFields) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccessRuleMetadataFields) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccessRuleMetadataFields) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *AccessRuleMetadataFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccessRuleMetadataFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccessRuleMetadataFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *AccessRuleMetadataFields) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AccessRuleMetadataFields) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AccessRuleMetadataFields) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *AccessRuleMetadataFields) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *AccessRuleMetadataFields) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *AccessRuleMetadataFields) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *AccessRuleMetadataFields) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *AccessRuleMetadataFields) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *AccessRuleMetadataFields) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *AccessRuleMetadataFields) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AccessRuleMetadataFields) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AccessRuleMetadataFields) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *AccessRuleMetadataFields) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetCreatedBy

`func (o *AccessRuleMetadataFields) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AccessRuleMetadataFields) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AccessRuleMetadataFields) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


