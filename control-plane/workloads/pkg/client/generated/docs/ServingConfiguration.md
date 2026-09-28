# ServingConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InitializationTimeoutSeconds** | Pointer to **NullableInt32** | The maximum time (in seconds) allowed for a workload to initialize and become ready. If the workload does not start within this time, it will be moved to failed state. | [optional] 
**RequestTimeoutSeconds** | Pointer to **NullableInt32** | The maximum time (in seconds) allowed to process an end-user request. If no response is returned within this time, the request will be ignored.Supported from Cluster version 2.22 | [optional] 

## Methods

### NewServingConfiguration

`func NewServingConfiguration() *ServingConfiguration`

NewServingConfiguration instantiates a new ServingConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingConfigurationWithDefaults

`func NewServingConfigurationWithDefaults() *ServingConfiguration`

NewServingConfigurationWithDefaults instantiates a new ServingConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInitializationTimeoutSeconds

`func (o *ServingConfiguration) GetInitializationTimeoutSeconds() int32`

GetInitializationTimeoutSeconds returns the InitializationTimeoutSeconds field if non-nil, zero value otherwise.

### GetInitializationTimeoutSecondsOk

`func (o *ServingConfiguration) GetInitializationTimeoutSecondsOk() (*int32, bool)`

GetInitializationTimeoutSecondsOk returns a tuple with the InitializationTimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitializationTimeoutSeconds

`func (o *ServingConfiguration) SetInitializationTimeoutSeconds(v int32)`

SetInitializationTimeoutSeconds sets InitializationTimeoutSeconds field to given value.

### HasInitializationTimeoutSeconds

`func (o *ServingConfiguration) HasInitializationTimeoutSeconds() bool`

HasInitializationTimeoutSeconds returns a boolean if a field has been set.

### SetInitializationTimeoutSecondsNil

`func (o *ServingConfiguration) SetInitializationTimeoutSecondsNil(b bool)`

 SetInitializationTimeoutSecondsNil sets the value for InitializationTimeoutSeconds to be an explicit nil

### UnsetInitializationTimeoutSeconds
`func (o *ServingConfiguration) UnsetInitializationTimeoutSeconds()`

UnsetInitializationTimeoutSeconds ensures that no value is present for InitializationTimeoutSeconds, not even an explicit nil
### GetRequestTimeoutSeconds

`func (o *ServingConfiguration) GetRequestTimeoutSeconds() int32`

GetRequestTimeoutSeconds returns the RequestTimeoutSeconds field if non-nil, zero value otherwise.

### GetRequestTimeoutSecondsOk

`func (o *ServingConfiguration) GetRequestTimeoutSecondsOk() (*int32, bool)`

GetRequestTimeoutSecondsOk returns a tuple with the RequestTimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestTimeoutSeconds

`func (o *ServingConfiguration) SetRequestTimeoutSeconds(v int32)`

SetRequestTimeoutSeconds sets RequestTimeoutSeconds field to given value.

### HasRequestTimeoutSeconds

`func (o *ServingConfiguration) HasRequestTimeoutSeconds() bool`

HasRequestTimeoutSeconds returns a boolean if a field has been set.

### SetRequestTimeoutSecondsNil

`func (o *ServingConfiguration) SetRequestTimeoutSecondsNil(b bool)`

 SetRequestTimeoutSecondsNil sets the value for RequestTimeoutSeconds to be an explicit nil

### UnsetRequestTimeoutSeconds
`func (o *ServingConfiguration) UnsetRequestTimeoutSeconds()`

UnsetRequestTimeoutSeconds ensures that no value is present for RequestTimeoutSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


