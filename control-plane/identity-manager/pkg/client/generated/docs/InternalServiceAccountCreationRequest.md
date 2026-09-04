# InternalServiceAccountCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the service account. The name must be unique within the organization and can only contain lowercase alphanumeric characters and hyphens. It must start and end with a letter. | 
**Realm** | **string** | The realm in which to create the service account | 

## Methods

### NewInternalServiceAccountCreationRequest

`func NewInternalServiceAccountCreationRequest(name string, realm string, ) *InternalServiceAccountCreationRequest`

NewInternalServiceAccountCreationRequest instantiates a new InternalServiceAccountCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalServiceAccountCreationRequestWithDefaults

`func NewInternalServiceAccountCreationRequestWithDefaults() *InternalServiceAccountCreationRequest`

NewInternalServiceAccountCreationRequestWithDefaults instantiates a new InternalServiceAccountCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *InternalServiceAccountCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InternalServiceAccountCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InternalServiceAccountCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetRealm

`func (o *InternalServiceAccountCreationRequest) GetRealm() string`

GetRealm returns the Realm field if non-nil, zero value otherwise.

### GetRealmOk

`func (o *InternalServiceAccountCreationRequest) GetRealmOk() (*string, bool)`

GetRealmOk returns a tuple with the Realm field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRealm

`func (o *InternalServiceAccountCreationRequest) SetRealm(v string)`

SetRealm sets Realm field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


