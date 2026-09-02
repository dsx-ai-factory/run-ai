# NetworkTopologyCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Levels** | **[]string** | Ordered array of network topology levels | 
**Name** | **string** | Name of the network topology | 
**ClusterId** | **string** | The id of the cluster. | 

## Methods

### NewNetworkTopologyCreateRequest

`func NewNetworkTopologyCreateRequest(levels []string, name string, clusterId string, ) *NetworkTopologyCreateRequest`

NewNetworkTopologyCreateRequest instantiates a new NetworkTopologyCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologyCreateRequestWithDefaults

`func NewNetworkTopologyCreateRequestWithDefaults() *NetworkTopologyCreateRequest`

NewNetworkTopologyCreateRequestWithDefaults instantiates a new NetworkTopologyCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLevels

`func (o *NetworkTopologyCreateRequest) GetLevels() []string`

GetLevels returns the Levels field if non-nil, zero value otherwise.

### GetLevelsOk

`func (o *NetworkTopologyCreateRequest) GetLevelsOk() (*[]string, bool)`

GetLevelsOk returns a tuple with the Levels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevels

`func (o *NetworkTopologyCreateRequest) SetLevels(v []string)`

SetLevels sets Levels field to given value.


### GetName

`func (o *NetworkTopologyCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NetworkTopologyCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NetworkTopologyCreateRequest) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *NetworkTopologyCreateRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NetworkTopologyCreateRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NetworkTopologyCreateRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


