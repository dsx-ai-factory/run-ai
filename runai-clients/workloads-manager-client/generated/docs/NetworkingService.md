# NetworkingService

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Spec** | Pointer to [**NullableServiceSpec**](ServiceSpec.md) |  | [optional] 
**Ref** | Pointer to [**NullableServiceRef**](ServiceRef.md) |  | [optional] 
**Authorization** | Pointer to [**NullableWorkloadv2Authorization**](Workloadv2Authorization.md) |  | [optional] 

## Methods

### NewNetworkingService

`func NewNetworkingService() *NetworkingService`

NewNetworkingService instantiates a new NetworkingService object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkingServiceWithDefaults

`func NewNetworkingServiceWithDefaults() *NetworkingService`

NewNetworkingServiceWithDefaults instantiates a new NetworkingService object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpec

`func (o *NetworkingService) GetSpec() ServiceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *NetworkingService) GetSpecOk() (*ServiceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *NetworkingService) SetSpec(v ServiceSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *NetworkingService) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### SetSpecNil

`func (o *NetworkingService) SetSpecNil(b bool)`

 SetSpecNil sets the value for Spec to be an explicit nil

### UnsetSpec
`func (o *NetworkingService) UnsetSpec()`

UnsetSpec ensures that no value is present for Spec, not even an explicit nil
### GetRef

`func (o *NetworkingService) GetRef() ServiceRef`

GetRef returns the Ref field if non-nil, zero value otherwise.

### GetRefOk

`func (o *NetworkingService) GetRefOk() (*ServiceRef, bool)`

GetRefOk returns a tuple with the Ref field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRef

`func (o *NetworkingService) SetRef(v ServiceRef)`

SetRef sets Ref field to given value.

### HasRef

`func (o *NetworkingService) HasRef() bool`

HasRef returns a boolean if a field has been set.

### SetRefNil

`func (o *NetworkingService) SetRefNil(b bool)`

 SetRefNil sets the value for Ref to be an explicit nil

### UnsetRef
`func (o *NetworkingService) UnsetRef()`

UnsetRef ensures that no value is present for Ref, not even an explicit nil
### GetAuthorization

`func (o *NetworkingService) GetAuthorization() Workloadv2Authorization`

GetAuthorization returns the Authorization field if non-nil, zero value otherwise.

### GetAuthorizationOk

`func (o *NetworkingService) GetAuthorizationOk() (*Workloadv2Authorization, bool)`

GetAuthorizationOk returns a tuple with the Authorization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorization

`func (o *NetworkingService) SetAuthorization(v Workloadv2Authorization)`

SetAuthorization sets Authorization field to given value.

### HasAuthorization

`func (o *NetworkingService) HasAuthorization() bool`

HasAuthorization returns a boolean if a field has been set.

### SetAuthorizationNil

`func (o *NetworkingService) SetAuthorizationNil(b bool)`

 SetAuthorizationNil sets the value for Authorization to be an explicit nil

### UnsetAuthorization
`func (o *NetworkingService) UnsetAuthorization()`

UnsetAuthorization ensures that no value is present for Authorization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


