# ResourceClaimTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | ResourceClaimTemplate name. Must follow Kubernetes DNS-1123 subdomain rules (lowercase alphanumeric, &#x60;-&#x60;, and &#x60;.&#x60;, starting and ending with an alphanumeric character), up to 253 characters.  | 
**Namespace** | **string** | the namespace the ResourceClaimTemplate lives in | 
**Origin** | Pointer to [**NullableOrigin**](Origin.md) |  | [optional] 
**AssetId** | Pointer to **NullableString** | Control-plane asset id this template was created from. Present only when origin is \&quot;assets\&quot;. Allows the control plane to join tracker rows back to the originating asset record.  | [optional] 
**Spec** | Pointer to **map[string]interface{}** | Raw ResourceClaimTemplate spec as sent by the agent. Stored as-is so that the control plane has enough fidelity to describe what will be submitted on claim creation, rather than relying on the template name alone. Shape follows the Kubernetes &#x60;resource.k8s.io/v1&#x60; &#x60;ResourceClaimTemplateSpec&#x60;.  | [optional] 
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 

## Methods

### NewResourceClaimTemplate

`func NewResourceClaimTemplate(name string, namespace string, ) *ResourceClaimTemplate`

NewResourceClaimTemplate instantiates a new ResourceClaimTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceClaimTemplateWithDefaults

`func NewResourceClaimTemplateWithDefaults() *ResourceClaimTemplate`

NewResourceClaimTemplateWithDefaults instantiates a new ResourceClaimTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ResourceClaimTemplate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceClaimTemplate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceClaimTemplate) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *ResourceClaimTemplate) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ResourceClaimTemplate) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ResourceClaimTemplate) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetOrigin

`func (o *ResourceClaimTemplate) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *ResourceClaimTemplate) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *ResourceClaimTemplate) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.

### HasOrigin

`func (o *ResourceClaimTemplate) HasOrigin() bool`

HasOrigin returns a boolean if a field has been set.

### SetOriginNil

`func (o *ResourceClaimTemplate) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *ResourceClaimTemplate) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetAssetId

`func (o *ResourceClaimTemplate) GetAssetId() string`

GetAssetId returns the AssetId field if non-nil, zero value otherwise.

### GetAssetIdOk

`func (o *ResourceClaimTemplate) GetAssetIdOk() (*string, bool)`

GetAssetIdOk returns a tuple with the AssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssetId

`func (o *ResourceClaimTemplate) SetAssetId(v string)`

SetAssetId sets AssetId field to given value.

### HasAssetId

`func (o *ResourceClaimTemplate) HasAssetId() bool`

HasAssetId returns a boolean if a field has been set.

### SetAssetIdNil

`func (o *ResourceClaimTemplate) SetAssetIdNil(b bool)`

 SetAssetIdNil sets the value for AssetId to be an explicit nil

### UnsetAssetId
`func (o *ResourceClaimTemplate) UnsetAssetId()`

UnsetAssetId ensures that no value is present for AssetId, not even an explicit nil
### GetSpec

`func (o *ResourceClaimTemplate) GetSpec() map[string]interface{}`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ResourceClaimTemplate) GetSpecOk() (*map[string]interface{}, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ResourceClaimTemplate) SetSpec(v map[string]interface{})`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ResourceClaimTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetIssues

`func (o *ResourceClaimTemplate) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *ResourceClaimTemplate) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *ResourceClaimTemplate) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *ResourceClaimTemplate) HasIssues() bool`

HasIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


