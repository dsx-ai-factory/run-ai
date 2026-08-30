# NimServiceUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | [**NimServiceSpec**](NimServiceSpec.md) |  | 
**FieldsToRemove** | Pointer to [**[]NimServiceRemovableFields**](NimServiceRemovableFields.md) | List of field names to remove from the NIM service spec. Note: replicas and autoscaling are mutually exclusive - adding one automatically removes the other.  | [optional] 

## Methods

### NewNimServiceUpdateRequest

`func NewNimServiceUpdateRequest(spec NimServiceSpec, ) *NimServiceUpdateRequest`

NewNimServiceUpdateRequest instantiates a new NimServiceUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceUpdateRequestWithDefaults

`func NewNimServiceUpdateRequestWithDefaults() *NimServiceUpdateRequest`

NewNimServiceUpdateRequestWithDefaults instantiates a new NimServiceUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *NimServiceUpdateRequest) GetSpec() NimServiceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *NimServiceUpdateRequest) GetSpecOk() (*NimServiceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *NimServiceUpdateRequest) SetSpec(v NimServiceSpec)`

SetSpec sets Spec field to given value.


### GetFieldsToRemove

`func (o *NimServiceUpdateRequest) GetFieldsToRemove() []NimServiceRemovableFields`

GetFieldsToRemove returns the FieldsToRemove field if non-nil, zero value otherwise.

### GetFieldsToRemoveOk

`func (o *NimServiceUpdateRequest) GetFieldsToRemoveOk() (*[]NimServiceRemovableFields, bool)`

GetFieldsToRemoveOk returns a tuple with the FieldsToRemove field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldsToRemove

`func (o *NimServiceUpdateRequest) SetFieldsToRemove(v []NimServiceRemovableFields)`

SetFieldsToRemove sets FieldsToRemove field to given value.

### HasFieldsToRemove

`func (o *NimServiceUpdateRequest) HasFieldsToRemove() bool`

HasFieldsToRemove returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


