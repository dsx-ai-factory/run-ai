# ClusterHistory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**Version** | **string** | the cluster version | 
**UpdateTimestamp** | **int64** | the cluster version update timestamp in unix format | 

## Methods

### NewClusterHistory

`func NewClusterHistory(clusterId string, version string, updateTimestamp int64, ) *ClusterHistory`

NewClusterHistory instantiates a new ClusterHistory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterHistoryWithDefaults

`func NewClusterHistoryWithDefaults() *ClusterHistory`

NewClusterHistoryWithDefaults instantiates a new ClusterHistory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *ClusterHistory) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ClusterHistory) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ClusterHistory) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetVersion

`func (o *ClusterHistory) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ClusterHistory) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ClusterHistory) SetVersion(v string)`

SetVersion sets Version field to given value.


### GetUpdateTimestamp

`func (o *ClusterHistory) GetUpdateTimestamp() int64`

GetUpdateTimestamp returns the UpdateTimestamp field if non-nil, zero value otherwise.

### GetUpdateTimestampOk

`func (o *ClusterHistory) GetUpdateTimestampOk() (*int64, bool)`

GetUpdateTimestampOk returns a tuple with the UpdateTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateTimestamp

`func (o *ClusterHistory) SetUpdateTimestamp(v int64)`

SetUpdateTimestamp sets UpdateTimestamp field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


