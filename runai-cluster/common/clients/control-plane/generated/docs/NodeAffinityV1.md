# NodeAffinityV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AffinityType** | Pointer to **string** | The type of affinity of the jobs on the nodes. | [optional] 
**SelectedTypes** | Pointer to [**[]NodeAffinityV1SelectedTypesInner**](NodeAffinityV1SelectedTypesInner.md) |  | [optional] 

## Methods

### NewNodeAffinityV1

`func NewNodeAffinityV1() *NodeAffinityV1`

NewNodeAffinityV1 instantiates a new NodeAffinityV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeAffinityV1WithDefaults

`func NewNodeAffinityV1WithDefaults() *NodeAffinityV1`

NewNodeAffinityV1WithDefaults instantiates a new NodeAffinityV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAffinityType

`func (o *NodeAffinityV1) GetAffinityType() string`

GetAffinityType returns the AffinityType field if non-nil, zero value otherwise.

### GetAffinityTypeOk

`func (o *NodeAffinityV1) GetAffinityTypeOk() (*string, bool)`

GetAffinityTypeOk returns a tuple with the AffinityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffinityType

`func (o *NodeAffinityV1) SetAffinityType(v string)`

SetAffinityType sets AffinityType field to given value.

### HasAffinityType

`func (o *NodeAffinityV1) HasAffinityType() bool`

HasAffinityType returns a boolean if a field has been set.

### GetSelectedTypes

`func (o *NodeAffinityV1) GetSelectedTypes() []NodeAffinityV1SelectedTypesInner`

GetSelectedTypes returns the SelectedTypes field if non-nil, zero value otherwise.

### GetSelectedTypesOk

`func (o *NodeAffinityV1) GetSelectedTypesOk() (*[]NodeAffinityV1SelectedTypesInner, bool)`

GetSelectedTypesOk returns a tuple with the SelectedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedTypes

`func (o *NodeAffinityV1) SetSelectedTypes(v []NodeAffinityV1SelectedTypesInner)`

SetSelectedTypes sets SelectedTypes field to given value.

### HasSelectedTypes

`func (o *NodeAffinityV1) HasSelectedTypes() bool`

HasSelectedTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


