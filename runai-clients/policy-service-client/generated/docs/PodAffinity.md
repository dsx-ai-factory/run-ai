# PodAffinity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to [**NullablePodAffinityType**](PodAffinityType.md) |  | [optional] 
**Key** | Pointer to **NullableString** | The label key to use. (mandatory) | [optional] 

## Methods

### NewPodAffinity

`func NewPodAffinity() *PodAffinity`

NewPodAffinity instantiates a new PodAffinity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodAffinityWithDefaults

`func NewPodAffinityWithDefaults() *PodAffinity`

NewPodAffinityWithDefaults instantiates a new PodAffinity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *PodAffinity) GetType() PodAffinityType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *PodAffinity) GetTypeOk() (*PodAffinityType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *PodAffinity) SetType(v PodAffinityType)`

SetType sets Type field to given value.

### HasType

`func (o *PodAffinity) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *PodAffinity) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *PodAffinity) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetKey

`func (o *PodAffinity) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *PodAffinity) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *PodAffinity) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *PodAffinity) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *PodAffinity) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *PodAffinity) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


