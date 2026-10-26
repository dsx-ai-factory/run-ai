# ResourceRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **string** | The Kubernetes UID of the referenced resource. | [optional] 
**Name** | **string** | The name of the referenced resource. | 
**Kind** | [**ResourceRefKind**](ResourceRefKind.md) |  | 

## Methods

### NewResourceRef

`func NewResourceRef(name string, kind ResourceRefKind, ) *ResourceRef`

NewResourceRef instantiates a new ResourceRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceRefWithDefaults

`func NewResourceRefWithDefaults() *ResourceRef`

NewResourceRefWithDefaults instantiates a new ResourceRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *ResourceRef) GetUid() string`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *ResourceRef) GetUidOk() (*string, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *ResourceRef) SetUid(v string)`

SetUid sets Uid field to given value.

### HasUid

`func (o *ResourceRef) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetName

`func (o *ResourceRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceRef) SetName(v string)`

SetName sets Name field to given value.


### GetKind

`func (o *ResourceRef) GetKind() ResourceRefKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *ResourceRef) GetKindOk() (*ResourceRefKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *ResourceRef) SetKind(v ResourceRefKind)`

SetKind sets Kind field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


