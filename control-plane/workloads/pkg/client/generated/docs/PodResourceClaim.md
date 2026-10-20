# PodResourceClaim

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Pod-local name of the claim entry (DNS_LABEL). Mirrors &#x60;pod.spec.resourceClaims[].name&#x60; and the key used by &#x60;pod.status.resourceClaimStatuses[].name&#x60; to publish the bound claim. | 
**ResourceClaimTemplateName** | Pointer to **NullableString** | Name of the referenced &#x60;ResourceClaimTemplate&#x60; in the pod&#39;s namespace, when the claim entry was created from a template. Mutually exclusive with &#x60;resourceClaimName&#x60;. Mirrors &#x60;pod.spec.resourceClaims[].resourceClaimTemplateName&#x60;. | [optional] 
**ResourceClaimName** | Pointer to **NullableString** | Name of a standalone &#x60;ResourceClaim&#x60; that the pod references directly, when the claim entry is not created from a template. Mutually exclusive with &#x60;resourceClaimTemplateName&#x60;. Mirrors &#x60;pod.spec.resourceClaims[].resourceClaimName&#x60;. | [optional] 
**AllocatedResourceClaimName** | Pointer to **NullableString** | Name of the actual &#x60;ResourceClaim&#x60; bound to the pod for this entry. For a template-generated entry this is populated from &#x60;pod.status.resourceClaimStatuses[].resourceClaimName&#x60; once Kubernetes publishes it, and is null until then. For a standalone entry it equals &#x60;resourceClaimName&#x60; and is populated immediately. This is the field the forward-view API surfaces as the \&quot;actual\&quot; bound claim. | [optional] 

## Methods

### NewPodResourceClaim

`func NewPodResourceClaim(name string, ) *PodResourceClaim`

NewPodResourceClaim instantiates a new PodResourceClaim object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodResourceClaimWithDefaults

`func NewPodResourceClaimWithDefaults() *PodResourceClaim`

NewPodResourceClaimWithDefaults instantiates a new PodResourceClaim object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PodResourceClaim) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PodResourceClaim) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PodResourceClaim) SetName(v string)`

SetName sets Name field to given value.


### GetResourceClaimTemplateName

`func (o *PodResourceClaim) GetResourceClaimTemplateName() string`

GetResourceClaimTemplateName returns the ResourceClaimTemplateName field if non-nil, zero value otherwise.

### GetResourceClaimTemplateNameOk

`func (o *PodResourceClaim) GetResourceClaimTemplateNameOk() (*string, bool)`

GetResourceClaimTemplateNameOk returns a tuple with the ResourceClaimTemplateName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaimTemplateName

`func (o *PodResourceClaim) SetResourceClaimTemplateName(v string)`

SetResourceClaimTemplateName sets ResourceClaimTemplateName field to given value.

### HasResourceClaimTemplateName

`func (o *PodResourceClaim) HasResourceClaimTemplateName() bool`

HasResourceClaimTemplateName returns a boolean if a field has been set.

### SetResourceClaimTemplateNameNil

`func (o *PodResourceClaim) SetResourceClaimTemplateNameNil(b bool)`

 SetResourceClaimTemplateNameNil sets the value for ResourceClaimTemplateName to be an explicit nil

### UnsetResourceClaimTemplateName
`func (o *PodResourceClaim) UnsetResourceClaimTemplateName()`

UnsetResourceClaimTemplateName ensures that no value is present for ResourceClaimTemplateName, not even an explicit nil
### GetResourceClaimName

`func (o *PodResourceClaim) GetResourceClaimName() string`

GetResourceClaimName returns the ResourceClaimName field if non-nil, zero value otherwise.

### GetResourceClaimNameOk

`func (o *PodResourceClaim) GetResourceClaimNameOk() (*string, bool)`

GetResourceClaimNameOk returns a tuple with the ResourceClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaimName

`func (o *PodResourceClaim) SetResourceClaimName(v string)`

SetResourceClaimName sets ResourceClaimName field to given value.

### HasResourceClaimName

`func (o *PodResourceClaim) HasResourceClaimName() bool`

HasResourceClaimName returns a boolean if a field has been set.

### SetResourceClaimNameNil

`func (o *PodResourceClaim) SetResourceClaimNameNil(b bool)`

 SetResourceClaimNameNil sets the value for ResourceClaimName to be an explicit nil

### UnsetResourceClaimName
`func (o *PodResourceClaim) UnsetResourceClaimName()`

UnsetResourceClaimName ensures that no value is present for ResourceClaimName, not even an explicit nil
### GetAllocatedResourceClaimName

`func (o *PodResourceClaim) GetAllocatedResourceClaimName() string`

GetAllocatedResourceClaimName returns the AllocatedResourceClaimName field if non-nil, zero value otherwise.

### GetAllocatedResourceClaimNameOk

`func (o *PodResourceClaim) GetAllocatedResourceClaimNameOk() (*string, bool)`

GetAllocatedResourceClaimNameOk returns a tuple with the AllocatedResourceClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResourceClaimName

`func (o *PodResourceClaim) SetAllocatedResourceClaimName(v string)`

SetAllocatedResourceClaimName sets AllocatedResourceClaimName field to given value.

### HasAllocatedResourceClaimName

`func (o *PodResourceClaim) HasAllocatedResourceClaimName() bool`

HasAllocatedResourceClaimName returns a boolean if a field has been set.

### SetAllocatedResourceClaimNameNil

`func (o *PodResourceClaim) SetAllocatedResourceClaimNameNil(b bool)`

 SetAllocatedResourceClaimNameNil sets the value for AllocatedResourceClaimName to be an explicit nil

### UnsetAllocatedResourceClaimName
`func (o *PodResourceClaim) UnsetAllocatedResourceClaimName()`

UnsetAllocatedResourceClaimName ensures that no value is present for AllocatedResourceClaimName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


