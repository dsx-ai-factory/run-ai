# ClusterCalculatedStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** | The cluster&#39;s state. | [optional] 
**LastConnected** | Pointer to **time.Time** | The last time the cluster was connected to the control plane. | [optional] 

## Methods

### NewClusterCalculatedStatus

`func NewClusterCalculatedStatus() *ClusterCalculatedStatus`

NewClusterCalculatedStatus instantiates a new ClusterCalculatedStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterCalculatedStatusWithDefaults

`func NewClusterCalculatedStatusWithDefaults() *ClusterCalculatedStatus`

NewClusterCalculatedStatusWithDefaults instantiates a new ClusterCalculatedStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *ClusterCalculatedStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ClusterCalculatedStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ClusterCalculatedStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ClusterCalculatedStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetLastConnected

`func (o *ClusterCalculatedStatus) GetLastConnected() time.Time`

GetLastConnected returns the LastConnected field if non-nil, zero value otherwise.

### GetLastConnectedOk

`func (o *ClusterCalculatedStatus) GetLastConnectedOk() (*time.Time, bool)`

GetLastConnectedOk returns a tuple with the LastConnected field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastConnected

`func (o *ClusterCalculatedStatus) SetLastConnected(v time.Time)`

SetLastConnected sets LastConnected field to given value.

### HasLastConnected

`func (o *ClusterCalculatedStatus) HasLastConnected() bool`

HasLastConnected returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


