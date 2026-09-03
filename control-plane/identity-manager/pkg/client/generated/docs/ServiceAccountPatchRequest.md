# ServiceAccountPatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | **bool** | Indicates whether the service account is enabled. | 

## Methods

### NewServiceAccountPatchRequest

`func NewServiceAccountPatchRequest(enabled bool, ) *ServiceAccountPatchRequest`

NewServiceAccountPatchRequest instantiates a new ServiceAccountPatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServiceAccountPatchRequestWithDefaults

`func NewServiceAccountPatchRequestWithDefaults() *ServiceAccountPatchRequest`

NewServiceAccountPatchRequestWithDefaults instantiates a new ServiceAccountPatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *ServiceAccountPatchRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *ServiceAccountPatchRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *ServiceAccountPatchRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


