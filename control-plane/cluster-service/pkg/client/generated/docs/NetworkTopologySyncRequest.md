# NetworkTopologySyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FullSync** | Pointer to **bool** | Upsert all items and delete all missing items | [optional] 
**NetworkTopologies** | Pointer to [**[]NetworkTopologySyncFields**](NetworkTopologySyncFields.md) |  | [optional] 

## Methods

### NewNetworkTopologySyncRequest

`func NewNetworkTopologySyncRequest() *NetworkTopologySyncRequest`

NewNetworkTopologySyncRequest instantiates a new NetworkTopologySyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologySyncRequestWithDefaults

`func NewNetworkTopologySyncRequestWithDefaults() *NetworkTopologySyncRequest`

NewNetworkTopologySyncRequestWithDefaults instantiates a new NetworkTopologySyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFullSync

`func (o *NetworkTopologySyncRequest) GetFullSync() bool`

GetFullSync returns the FullSync field if non-nil, zero value otherwise.

### GetFullSyncOk

`func (o *NetworkTopologySyncRequest) GetFullSyncOk() (*bool, bool)`

GetFullSyncOk returns a tuple with the FullSync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullSync

`func (o *NetworkTopologySyncRequest) SetFullSync(v bool)`

SetFullSync sets FullSync field to given value.

### HasFullSync

`func (o *NetworkTopologySyncRequest) HasFullSync() bool`

HasFullSync returns a boolean if a field has been set.

### GetNetworkTopologies

`func (o *NetworkTopologySyncRequest) GetNetworkTopologies() []NetworkTopologySyncFields`

GetNetworkTopologies returns the NetworkTopologies field if non-nil, zero value otherwise.

### GetNetworkTopologiesOk

`func (o *NetworkTopologySyncRequest) GetNetworkTopologiesOk() (*[]NetworkTopologySyncFields, bool)`

GetNetworkTopologiesOk returns a tuple with the NetworkTopologies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologies

`func (o *NetworkTopologySyncRequest) SetNetworkTopologies(v []NetworkTopologySyncFields)`

SetNetworkTopologies sets NetworkTopologies field to given value.

### HasNetworkTopologies

`func (o *NetworkTopologySyncRequest) HasNetworkTopologies() bool`

HasNetworkTopologies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


