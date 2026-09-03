# GVK

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Group** | Pointer to **NullableString** | The API group of the Kubernetes resource. Empty string for core resources like Pod. | [optional] 
**Version** | **string** | The API version of the resource within the specified group. | 
**Kind** | **string** | The type of Kubernetes resource being referenced. | 

## Methods

### NewGVK

`func NewGVK(version string, kind string, ) *GVK`

NewGVK instantiates a new GVK object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGVKWithDefaults

`func NewGVKWithDefaults() *GVK`

NewGVKWithDefaults instantiates a new GVK object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGroup

`func (o *GVK) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *GVK) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *GVK) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *GVK) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### SetGroupNil

`func (o *GVK) SetGroupNil(b bool)`

 SetGroupNil sets the value for Group to be an explicit nil

### UnsetGroup
`func (o *GVK) UnsetGroup()`

UnsetGroup ensures that no value is present for Group, not even an explicit nil
### GetVersion

`func (o *GVK) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *GVK) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *GVK) SetVersion(v string)`

SetVersion sets Version field to given value.


### GetKind

`func (o *GVK) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *GVK) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *GVK) SetKind(v string)`

SetKind sets Kind field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


