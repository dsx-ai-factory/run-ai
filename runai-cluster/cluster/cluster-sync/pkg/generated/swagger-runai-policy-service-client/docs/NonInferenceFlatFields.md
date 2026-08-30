# NonInferenceFlatFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Tty** | Pointer to **NullableBool** | Whether this container should allocate a TTY for itself, also requires &#39;stdin&#39; to be true. | [optional] 
**Stdin** | Pointer to **NullableBool** | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF | [optional] 

## Methods

### NewNonInferenceFlatFields

`func NewNonInferenceFlatFields() *NonInferenceFlatFields`

NewNonInferenceFlatFields instantiates a new NonInferenceFlatFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNonInferenceFlatFieldsWithDefaults

`func NewNonInferenceFlatFieldsWithDefaults() *NonInferenceFlatFields`

NewNonInferenceFlatFieldsWithDefaults instantiates a new NonInferenceFlatFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTty

`func (o *NonInferenceFlatFields) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *NonInferenceFlatFields) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *NonInferenceFlatFields) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *NonInferenceFlatFields) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *NonInferenceFlatFields) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *NonInferenceFlatFields) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetStdin

`func (o *NonInferenceFlatFields) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *NonInferenceFlatFields) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *NonInferenceFlatFields) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *NonInferenceFlatFields) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *NonInferenceFlatFields) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *NonInferenceFlatFields) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


