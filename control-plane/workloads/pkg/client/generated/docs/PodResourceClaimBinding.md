# PodResourceClaimBinding

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Pod-local name of the claim entry, as listed in &#x60;pod.spec.resourceClaims[].name&#x60; and referenced by containers in &#x60;resources.claims&#x60;. | 
**ClaimName** | Pointer to **NullableString** | Name of the ResourceClaim bound to the pod for this entry. For a claim generated from a template, this is null until Kubernetes creates and binds the claim. | [optional] 

## Methods

### NewPodResourceClaimBinding

`func NewPodResourceClaimBinding(name string, ) *PodResourceClaimBinding`

NewPodResourceClaimBinding instantiates a new PodResourceClaimBinding object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodResourceClaimBindingWithDefaults

`func NewPodResourceClaimBindingWithDefaults() *PodResourceClaimBinding`

NewPodResourceClaimBindingWithDefaults instantiates a new PodResourceClaimBinding object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PodResourceClaimBinding) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PodResourceClaimBinding) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PodResourceClaimBinding) SetName(v string)`

SetName sets Name field to given value.


### GetClaimName

`func (o *PodResourceClaimBinding) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *PodResourceClaimBinding) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *PodResourceClaimBinding) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.

### HasClaimName

`func (o *PodResourceClaimBinding) HasClaimName() bool`

HasClaimName returns a boolean if a field has been set.

### SetClaimNameNil

`func (o *PodResourceClaimBinding) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *PodResourceClaimBinding) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


