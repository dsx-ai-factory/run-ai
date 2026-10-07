# ResourceClaim

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | ResourceClaim name. Must follow Kubernetes DNS-1123 subdomain rules (lowercase alphanumeric, &#x60;-&#x60;, and &#x60;.&#x60;, starting and ending with an alphanumeric character), up to 253 characters.  | 
**Namespace** | **string** | the namespace the ResourceClaim lives in | 
**ResourceClaimTemplateName** | Pointer to **NullableString** | Name of the ResourceClaimTemplate this claim was instantiated from, as reported by the agent. May be omitted when the agent could not resolve the template (e.g. the owning Pod was already gone).  | [optional] 
**Spec** | Pointer to **map[string]interface{}** | Raw ResourceClaim spec as sent by the agent. Shape follows the Kubernetes &#x60;resource.k8s.io/v1&#x60; &#x60;ResourceClaimSpec&#x60;. Stored verbatim; the tracker does not interpret its contents.  | [optional] 
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 

## Methods

### NewResourceClaim

`func NewResourceClaim(name string, namespace string, ) *ResourceClaim`

NewResourceClaim instantiates a new ResourceClaim object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceClaimWithDefaults

`func NewResourceClaimWithDefaults() *ResourceClaim`

NewResourceClaimWithDefaults instantiates a new ResourceClaim object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ResourceClaim) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceClaim) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceClaim) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *ResourceClaim) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ResourceClaim) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ResourceClaim) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetResourceClaimTemplateName

`func (o *ResourceClaim) GetResourceClaimTemplateName() string`

GetResourceClaimTemplateName returns the ResourceClaimTemplateName field if non-nil, zero value otherwise.

### GetResourceClaimTemplateNameOk

`func (o *ResourceClaim) GetResourceClaimTemplateNameOk() (*string, bool)`

GetResourceClaimTemplateNameOk returns a tuple with the ResourceClaimTemplateName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaimTemplateName

`func (o *ResourceClaim) SetResourceClaimTemplateName(v string)`

SetResourceClaimTemplateName sets ResourceClaimTemplateName field to given value.

### HasResourceClaimTemplateName

`func (o *ResourceClaim) HasResourceClaimTemplateName() bool`

HasResourceClaimTemplateName returns a boolean if a field has been set.

### SetResourceClaimTemplateNameNil

`func (o *ResourceClaim) SetResourceClaimTemplateNameNil(b bool)`

 SetResourceClaimTemplateNameNil sets the value for ResourceClaimTemplateName to be an explicit nil

### UnsetResourceClaimTemplateName
`func (o *ResourceClaim) UnsetResourceClaimTemplateName()`

UnsetResourceClaimTemplateName ensures that no value is present for ResourceClaimTemplateName, not even an explicit nil
### GetSpec

`func (o *ResourceClaim) GetSpec() map[string]interface{}`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ResourceClaim) GetSpecOk() (*map[string]interface{}, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ResourceClaim) SetSpec(v map[string]interface{})`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ResourceClaim) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetIssues

`func (o *ResourceClaim) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *ResourceClaim) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *ResourceClaim) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *ResourceClaim) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


