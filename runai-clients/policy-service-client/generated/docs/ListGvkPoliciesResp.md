# ListGvkPoliciesResp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | [**[]Policy**](Policy.md) | The policies matching the request, up to a maximum of 1000. | 

## Methods

### NewListGvkPoliciesResp

`func NewListGvkPoliciesResp(items []Policy, ) *ListGvkPoliciesResp`

NewListGvkPoliciesResp instantiates a new ListGvkPoliciesResp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListGvkPoliciesRespWithDefaults

`func NewListGvkPoliciesRespWithDefaults() *ListGvkPoliciesResp`

NewListGvkPoliciesRespWithDefaults instantiates a new ListGvkPoliciesResp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *ListGvkPoliciesResp) GetItems() []Policy`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *ListGvkPoliciesResp) GetItemsOk() (*[]Policy, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *ListGvkPoliciesResp) SetItems(v []Policy)`

SetItems sets Items field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


