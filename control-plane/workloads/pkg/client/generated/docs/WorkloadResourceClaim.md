# WorkloadResourceClaim

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Pod-local name of the claim entry, as listed in &#x60;pod.spec.resourceClaims[].name&#x60; and referenced by containers in &#x60;resources.claims&#x60;. | 
**ClaimName** | Pointer to **NullableString** | Name of the ResourceClaim bound to the pod for this entry. For a claim generated from a template, this is null until Kubernetes creates and binds the claim. | [optional] 
**Template** | Pointer to [**NullableResourceClaimTemplateRef**](ResourceClaimTemplateRef.md) |  | [optional] 

## Methods

### NewWorkloadResourceClaim

`func NewWorkloadResourceClaim(name string, ) *WorkloadResourceClaim`

NewWorkloadResourceClaim instantiates a new WorkloadResourceClaim object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadResourceClaimWithDefaults

`func NewWorkloadResourceClaimWithDefaults() *WorkloadResourceClaim`

NewWorkloadResourceClaimWithDefaults instantiates a new WorkloadResourceClaim object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadResourceClaim) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadResourceClaim) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadResourceClaim) SetName(v string)`

SetName sets Name field to given value.


### GetClaimName

`func (o *WorkloadResourceClaim) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *WorkloadResourceClaim) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *WorkloadResourceClaim) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.

### HasClaimName

`func (o *WorkloadResourceClaim) HasClaimName() bool`

HasClaimName returns a boolean if a field has been set.

### SetClaimNameNil

`func (o *WorkloadResourceClaim) SetClaimNameNil(b bool)`

 SetClaimNameNil sets the value for ClaimName to be an explicit nil

### UnsetClaimName
`func (o *WorkloadResourceClaim) UnsetClaimName()`

UnsetClaimName ensures that no value is present for ClaimName, not even an explicit nil
### GetTemplate

`func (o *WorkloadResourceClaim) GetTemplate() ResourceClaimTemplateRef`

GetTemplate returns the Template field if non-nil, zero value otherwise.

### GetTemplateOk

`func (o *WorkloadResourceClaim) GetTemplateOk() (*ResourceClaimTemplateRef, bool)`

GetTemplateOk returns a tuple with the Template field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplate

`func (o *WorkloadResourceClaim) SetTemplate(v ResourceClaimTemplateRef)`

SetTemplate sets Template field to given value.

### HasTemplate

`func (o *WorkloadResourceClaim) HasTemplate() bool`

HasTemplate returns a boolean if a field has been set.

### SetTemplateNil

`func (o *WorkloadResourceClaim) SetTemplateNil(b bool)`

 SetTemplateNil sets the value for Template to be an explicit nil

### UnsetTemplate
`func (o *WorkloadResourceClaim) UnsetTemplate()`

UnsetTemplate ensures that no value is present for Template, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


