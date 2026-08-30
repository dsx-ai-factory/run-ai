# FieldGroupDeleteItemSelectorsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ItemSelectors** | **[]string** | The names of the item selectors to delete (1-500 names). | 

## Methods

### NewFieldGroupDeleteItemSelectorsRequest

`func NewFieldGroupDeleteItemSelectorsRequest(itemSelectors []string, ) *FieldGroupDeleteItemSelectorsRequest`

NewFieldGroupDeleteItemSelectorsRequest instantiates a new FieldGroupDeleteItemSelectorsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFieldGroupDeleteItemSelectorsRequestWithDefaults

`func NewFieldGroupDeleteItemSelectorsRequestWithDefaults() *FieldGroupDeleteItemSelectorsRequest`

NewFieldGroupDeleteItemSelectorsRequestWithDefaults instantiates a new FieldGroupDeleteItemSelectorsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItemSelectors

`func (o *FieldGroupDeleteItemSelectorsRequest) GetItemSelectors() []string`

GetItemSelectors returns the ItemSelectors field if non-nil, zero value otherwise.

### GetItemSelectorsOk

`func (o *FieldGroupDeleteItemSelectorsRequest) GetItemSelectorsOk() (*[]string, bool)`

GetItemSelectorsOk returns a tuple with the ItemSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelectors

`func (o *FieldGroupDeleteItemSelectorsRequest) SetItemSelectors(v []string)`

SetItemSelectors sets ItemSelectors field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


