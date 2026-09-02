# ClusterCustomConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **int64** |  | [optional] [readonly] 
**TenantId** | Pointer to **float32** | The tenant id this cluster belongs to. | [optional] 
**ClusterId** | Pointer to **string** | Unique identifier of the cluster. | [optional] 
**Config** | Pointer to **map[string]interface{}** | The cluster custom configuration. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The time the cluster custom configuration was created. | [optional] [readonly] 

## Methods

### NewClusterCustomConfig

`func NewClusterCustomConfig() *ClusterCustomConfig`

NewClusterCustomConfig instantiates a new ClusterCustomConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterCustomConfigWithDefaults

`func NewClusterCustomConfigWithDefaults() *ClusterCustomConfig`

NewClusterCustomConfigWithDefaults instantiates a new ClusterCustomConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ClusterCustomConfig) GetId() int64`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ClusterCustomConfig) GetIdOk() (*int64, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ClusterCustomConfig) SetId(v int64)`

SetId sets Id field to given value.

### HasId

`func (o *ClusterCustomConfig) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTenantId

`func (o *ClusterCustomConfig) GetTenantId() float32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ClusterCustomConfig) GetTenantIdOk() (*float32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ClusterCustomConfig) SetTenantId(v float32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ClusterCustomConfig) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterId

`func (o *ClusterCustomConfig) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ClusterCustomConfig) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ClusterCustomConfig) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *ClusterCustomConfig) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetConfig

`func (o *ClusterCustomConfig) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ClusterCustomConfig) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ClusterCustomConfig) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.

### HasConfig

`func (o *ClusterCustomConfig) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ClusterCustomConfig) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ClusterCustomConfig) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ClusterCustomConfig) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ClusterCustomConfig) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


