# InitializationTimeoutField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InitializationTimeoutSeconds** | Pointer to **NullableInt32** | Use &#x60;servingConfiguration.initializationTimeoutSeconds&#x60; instead.  If this field is set, it will be ignored and the value under &#x60;servingConfiguration&#x60; will be used. The maximum amount of time (in seconds) to wait for the container to become ready. | [optional] 

## Methods

### NewInitializationTimeoutField

`func NewInitializationTimeoutField() *InitializationTimeoutField`

NewInitializationTimeoutField instantiates a new InitializationTimeoutField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInitializationTimeoutFieldWithDefaults

`func NewInitializationTimeoutFieldWithDefaults() *InitializationTimeoutField`

NewInitializationTimeoutFieldWithDefaults instantiates a new InitializationTimeoutField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInitializationTimeoutSeconds

`func (o *InitializationTimeoutField) GetInitializationTimeoutSeconds() int32`

GetInitializationTimeoutSeconds returns the InitializationTimeoutSeconds field if non-nil, zero value otherwise.

### GetInitializationTimeoutSecondsOk

`func (o *InitializationTimeoutField) GetInitializationTimeoutSecondsOk() (*int32, bool)`

GetInitializationTimeoutSecondsOk returns a tuple with the InitializationTimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitializationTimeoutSeconds

`func (o *InitializationTimeoutField) SetInitializationTimeoutSeconds(v int32)`

SetInitializationTimeoutSeconds sets InitializationTimeoutSeconds field to given value.

### HasInitializationTimeoutSeconds

`func (o *InitializationTimeoutField) HasInitializationTimeoutSeconds() bool`

HasInitializationTimeoutSeconds returns a boolean if a field has been set.

### SetInitializationTimeoutSecondsNil

`func (o *InitializationTimeoutField) SetInitializationTimeoutSecondsNil(b bool)`

 SetInitializationTimeoutSecondsNil sets the value for InitializationTimeoutSeconds to be an explicit nil

### UnsetInitializationTimeoutSeconds
`func (o *InitializationTimeoutField) UnsetInitializationTimeoutSeconds()`

UnsetInitializationTimeoutSeconds ensures that no value is present for InitializationTimeoutSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


