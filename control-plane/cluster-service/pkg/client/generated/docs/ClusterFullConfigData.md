# ClusterFullConfigData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SystemConfig** | Pointer to [**ClusterSystemConfigData**](ClusterSystemConfigData.md) |  | [optional] 
**UserConfig** | Pointer to [**ClusterUserConfigData**](ClusterUserConfigData.md) |  | [optional] 

## Methods

### NewClusterFullConfigData

`func NewClusterFullConfigData() *ClusterFullConfigData`

NewClusterFullConfigData instantiates a new ClusterFullConfigData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterFullConfigDataWithDefaults

`func NewClusterFullConfigDataWithDefaults() *ClusterFullConfigData`

NewClusterFullConfigDataWithDefaults instantiates a new ClusterFullConfigData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSystemConfig

`func (o *ClusterFullConfigData) GetSystemConfig() ClusterSystemConfigData`

GetSystemConfig returns the SystemConfig field if non-nil, zero value otherwise.

### GetSystemConfigOk

`func (o *ClusterFullConfigData) GetSystemConfigOk() (*ClusterSystemConfigData, bool)`

GetSystemConfigOk returns a tuple with the SystemConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemConfig

`func (o *ClusterFullConfigData) SetSystemConfig(v ClusterSystemConfigData)`

SetSystemConfig sets SystemConfig field to given value.

### HasSystemConfig

`func (o *ClusterFullConfigData) HasSystemConfig() bool`

HasSystemConfig returns a boolean if a field has been set.

### GetUserConfig

`func (o *ClusterFullConfigData) GetUserConfig() ClusterUserConfigData`

GetUserConfig returns the UserConfig field if non-nil, zero value otherwise.

### GetUserConfigOk

`func (o *ClusterFullConfigData) GetUserConfigOk() (*ClusterUserConfigData, bool)`

GetUserConfigOk returns a tuple with the UserConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserConfig

`func (o *ClusterFullConfigData) SetUserConfig(v ClusterUserConfigData)`

SetUserConfig sets UserConfig field to given value.

### HasUserConfig

`func (o *ClusterFullConfigData) HasUserConfig() bool`

HasUserConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


