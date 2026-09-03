# GetInternalClusterConfigResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**ClusterId** | Pointer to **string** |  | [optional] 
**Config** | [**ClusterFullConfigData**](ClusterFullConfigData.md) |  | 

## Methods

### NewGetInternalClusterConfigResponse

`func NewGetInternalClusterConfigResponse(id string, createdAt time.Time, updatedAt time.Time, config ClusterFullConfigData, ) *GetInternalClusterConfigResponse`

NewGetInternalClusterConfigResponse instantiates a new GetInternalClusterConfigResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetInternalClusterConfigResponseWithDefaults

`func NewGetInternalClusterConfigResponseWithDefaults() *GetInternalClusterConfigResponse`

NewGetInternalClusterConfigResponseWithDefaults instantiates a new GetInternalClusterConfigResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GetInternalClusterConfigResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetInternalClusterConfigResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetInternalClusterConfigResponse) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *GetInternalClusterConfigResponse) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetInternalClusterConfigResponse) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetInternalClusterConfigResponse) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *GetInternalClusterConfigResponse) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GetInternalClusterConfigResponse) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GetInternalClusterConfigResponse) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *GetInternalClusterConfigResponse) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *GetInternalClusterConfigResponse) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *GetInternalClusterConfigResponse) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *GetInternalClusterConfigResponse) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *GetInternalClusterConfigResponse) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *GetInternalClusterConfigResponse) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *GetInternalClusterConfigResponse) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *GetInternalClusterConfigResponse) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *GetInternalClusterConfigResponse) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *GetInternalClusterConfigResponse) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterId

`func (o *GetInternalClusterConfigResponse) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *GetInternalClusterConfigResponse) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *GetInternalClusterConfigResponse) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *GetInternalClusterConfigResponse) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetConfig

`func (o *GetInternalClusterConfigResponse) GetConfig() ClusterFullConfigData`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *GetInternalClusterConfigResponse) GetConfigOk() (*ClusterFullConfigData, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *GetInternalClusterConfigResponse) SetConfig(v ClusterFullConfigData)`

SetConfig sets Config field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


