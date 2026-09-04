# StorageClassConfigurationListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StorageClassConfigurations** | [**[]StorageClassConfiguration**](StorageClassConfiguration.md) | A list of storage class configuration objects. Each object defines the configuration details for a specific storage class. | 
**Next** | Pointer to **int32** |  | [optional] 

## Methods

### NewStorageClassConfigurationListResponse

`func NewStorageClassConfigurationListResponse(storageClassConfigurations []StorageClassConfiguration, ) *StorageClassConfigurationListResponse`

NewStorageClassConfigurationListResponse instantiates a new StorageClassConfigurationListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassConfigurationListResponseWithDefaults

`func NewStorageClassConfigurationListResponseWithDefaults() *StorageClassConfigurationListResponse`

NewStorageClassConfigurationListResponseWithDefaults instantiates a new StorageClassConfigurationListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStorageClassConfigurations

`func (o *StorageClassConfigurationListResponse) GetStorageClassConfigurations() []StorageClassConfiguration`

GetStorageClassConfigurations returns the StorageClassConfigurations field if non-nil, zero value otherwise.

### GetStorageClassConfigurationsOk

`func (o *StorageClassConfigurationListResponse) GetStorageClassConfigurationsOk() (*[]StorageClassConfiguration, bool)`

GetStorageClassConfigurationsOk returns a tuple with the StorageClassConfigurations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassConfigurations

`func (o *StorageClassConfigurationListResponse) SetStorageClassConfigurations(v []StorageClassConfiguration)`

SetStorageClassConfigurations sets StorageClassConfigurations field to given value.


### GetNext

`func (o *StorageClassConfigurationListResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *StorageClassConfigurationListResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *StorageClassConfigurationListResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *StorageClassConfigurationListResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


