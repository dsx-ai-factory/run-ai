# DatavolumesSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Datavolumes** | [**[]DatavolumeStatusForSync**](DatavolumeStatusForSync.md) |  | 
**ClusterId** | **string** | The id of the cluster. | 
**From** | Pointer to **string** | the first datavolume id to sync. if empty, all datavolumes before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | the last datavolume to sync. if empty, all datavolumes after the last one sent will be deleted. | [optional] 

## Methods

### NewDatavolumesSyncRequest

`func NewDatavolumesSyncRequest(datavolumes []DatavolumeStatusForSync, clusterId string, ) *DatavolumesSyncRequest`

NewDatavolumesSyncRequest instantiates a new DatavolumesSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumesSyncRequestWithDefaults

`func NewDatavolumesSyncRequestWithDefaults() *DatavolumesSyncRequest`

NewDatavolumesSyncRequestWithDefaults instantiates a new DatavolumesSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDatavolumes

`func (o *DatavolumesSyncRequest) GetDatavolumes() []DatavolumeStatusForSync`

GetDatavolumes returns the Datavolumes field if non-nil, zero value otherwise.

### GetDatavolumesOk

`func (o *DatavolumesSyncRequest) GetDatavolumesOk() (*[]DatavolumeStatusForSync, bool)`

GetDatavolumesOk returns a tuple with the Datavolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumes

`func (o *DatavolumesSyncRequest) SetDatavolumes(v []DatavolumeStatusForSync)`

SetDatavolumes sets Datavolumes field to given value.


### GetClusterId

`func (o *DatavolumesSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DatavolumesSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DatavolumesSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetFrom

`func (o *DatavolumesSyncRequest) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *DatavolumesSyncRequest) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *DatavolumesSyncRequest) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *DatavolumesSyncRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *DatavolumesSyncRequest) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *DatavolumesSyncRequest) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *DatavolumesSyncRequest) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *DatavolumesSyncRequest) HasTo() bool`

HasTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


