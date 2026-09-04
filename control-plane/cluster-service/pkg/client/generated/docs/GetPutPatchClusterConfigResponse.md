# GetPutPatchClusterConfigResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**ClusterId** | Pointer to **string** |  | [optional] 
**Config** | [**ClusterUserConfigData**](ClusterUserConfigData.md) |  | 
**Status** | Pointer to [**ClusterConfigStatus**](ClusterConfigStatus.md) |  | [optional] 

## Methods

### NewGetPutPatchClusterConfigResponse

`func NewGetPutPatchClusterConfigResponse(id string, createdAt time.Time, updatedAt time.Time, config ClusterUserConfigData, ) *GetPutPatchClusterConfigResponse`

NewGetPutPatchClusterConfigResponse instantiates a new GetPutPatchClusterConfigResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetPutPatchClusterConfigResponseWithDefaults

`func NewGetPutPatchClusterConfigResponseWithDefaults() *GetPutPatchClusterConfigResponse`

NewGetPutPatchClusterConfigResponseWithDefaults instantiates a new GetPutPatchClusterConfigResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GetPutPatchClusterConfigResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetPutPatchClusterConfigResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetPutPatchClusterConfigResponse) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *GetPutPatchClusterConfigResponse) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetPutPatchClusterConfigResponse) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetPutPatchClusterConfigResponse) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *GetPutPatchClusterConfigResponse) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GetPutPatchClusterConfigResponse) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GetPutPatchClusterConfigResponse) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *GetPutPatchClusterConfigResponse) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *GetPutPatchClusterConfigResponse) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *GetPutPatchClusterConfigResponse) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *GetPutPatchClusterConfigResponse) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *GetPutPatchClusterConfigResponse) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *GetPutPatchClusterConfigResponse) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *GetPutPatchClusterConfigResponse) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *GetPutPatchClusterConfigResponse) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *GetPutPatchClusterConfigResponse) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *GetPutPatchClusterConfigResponse) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterId

`func (o *GetPutPatchClusterConfigResponse) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *GetPutPatchClusterConfigResponse) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *GetPutPatchClusterConfigResponse) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *GetPutPatchClusterConfigResponse) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetConfig

`func (o *GetPutPatchClusterConfigResponse) GetConfig() ClusterUserConfigData`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *GetPutPatchClusterConfigResponse) GetConfigOk() (*ClusterUserConfigData, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *GetPutPatchClusterConfigResponse) SetConfig(v ClusterUserConfigData)`

SetConfig sets Config field to given value.


### GetStatus

`func (o *GetPutPatchClusterConfigResponse) GetStatus() ClusterConfigStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GetPutPatchClusterConfigResponse) GetStatusOk() (*ClusterConfigStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GetPutPatchClusterConfigResponse) SetStatus(v ClusterConfigStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *GetPutPatchClusterConfigResponse) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


