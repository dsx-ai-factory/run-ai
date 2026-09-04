# UpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**UpdateSpecSpec**](UpdateSpecSpec.md) |  | [optional] 

## Methods

### NewUpdateRequest

`func NewUpdateRequest() *UpdateRequest`

NewUpdateRequest instantiates a new UpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateRequestWithDefaults

`func NewUpdateRequestWithDefaults() *UpdateRequest`

NewUpdateRequestWithDefaults instantiates a new UpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *UpdateRequest) GetSpec() UpdateSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *UpdateRequest) GetSpecOk() (*UpdateSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *UpdateRequest) SetSpec(v UpdateSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *UpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


