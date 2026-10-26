# ResourceClaimTemplateRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Name of the ResourceClaimTemplate in the pod&#39;s namespace. | 
**Managed** | **bool** | Whether the template is managed by NVIDIA Run:ai (created through the resource claim templates API) rather than created directly in the cluster. | 
**Id** | Pointer to **NullableString** | The id of the managed resource claim template. Null for an unmanaged template. | [optional] 

## Methods

### NewResourceClaimTemplateRef

`func NewResourceClaimTemplateRef(name string, managed bool, ) *ResourceClaimTemplateRef`

NewResourceClaimTemplateRef instantiates a new ResourceClaimTemplateRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceClaimTemplateRefWithDefaults

`func NewResourceClaimTemplateRefWithDefaults() *ResourceClaimTemplateRef`

NewResourceClaimTemplateRefWithDefaults instantiates a new ResourceClaimTemplateRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ResourceClaimTemplateRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceClaimTemplateRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceClaimTemplateRef) SetName(v string)`

SetName sets Name field to given value.


### GetManaged

`func (o *ResourceClaimTemplateRef) GetManaged() bool`

GetManaged returns the Managed field if non-nil, zero value otherwise.

### GetManagedOk

`func (o *ResourceClaimTemplateRef) GetManagedOk() (*bool, bool)`

GetManagedOk returns a tuple with the Managed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetManaged

`func (o *ResourceClaimTemplateRef) SetManaged(v bool)`

SetManaged sets Managed field to given value.


### GetId

`func (o *ResourceClaimTemplateRef) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ResourceClaimTemplateRef) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ResourceClaimTemplateRef) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ResourceClaimTemplateRef) HasId() bool`

HasId returns a boolean if a field has been set.

### SetIdNil

`func (o *ResourceClaimTemplateRef) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *ResourceClaimTemplateRef) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


