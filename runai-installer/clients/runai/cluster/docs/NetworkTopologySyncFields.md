# NetworkTopologySyncFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Source** | **string** | Identifier of the the source of the topology. Useful for tracking where the topology data originated (i.e. Run:ai, Cluster). | 
**Name** | **string** | Name of the network topology | 
**Levels** | **[]string** | Ordered array of network topology levels | 
**Action** | Pointer to [**NetworkTopologySyncAction**](NetworkTopologySyncAction.md) |  | [optional] 

## Methods

### NewNetworkTopologySyncFields

`func NewNetworkTopologySyncFields(source string, name string, levels []string, ) *NetworkTopologySyncFields`

NewNetworkTopologySyncFields instantiates a new NetworkTopologySyncFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologySyncFieldsWithDefaults

`func NewNetworkTopologySyncFieldsWithDefaults() *NetworkTopologySyncFields`

NewNetworkTopologySyncFieldsWithDefaults instantiates a new NetworkTopologySyncFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSource

`func (o *NetworkTopologySyncFields) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *NetworkTopologySyncFields) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *NetworkTopologySyncFields) SetSource(v string)`

SetSource sets Source field to given value.


### GetName

`func (o *NetworkTopologySyncFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NetworkTopologySyncFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NetworkTopologySyncFields) SetName(v string)`

SetName sets Name field to given value.


### GetLevels

`func (o *NetworkTopologySyncFields) GetLevels() []string`

GetLevels returns the Levels field if non-nil, zero value otherwise.

### GetLevelsOk

`func (o *NetworkTopologySyncFields) GetLevelsOk() (*[]string, bool)`

GetLevelsOk returns a tuple with the Levels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevels

`func (o *NetworkTopologySyncFields) SetLevels(v []string)`

SetLevels sets Levels field to given value.


### GetAction

`func (o *NetworkTopologySyncFields) GetAction() NetworkTopologySyncAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *NetworkTopologySyncFields) GetActionOk() (*NetworkTopologySyncAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *NetworkTopologySyncFields) SetAction(v NetworkTopologySyncAction)`

SetAction sets Action field to given value.

### HasAction

`func (o *NetworkTopologySyncFields) HasAction() bool`

HasAction returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


