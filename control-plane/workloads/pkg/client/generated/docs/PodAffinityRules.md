# PodAffinityRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to [**NullablePodAffinityTypeRules**](PodAffinityTypeRules.md) |  | [optional] 
**Key** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewPodAffinityRules

`func NewPodAffinityRules() *PodAffinityRules`

NewPodAffinityRules instantiates a new PodAffinityRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodAffinityRulesWithDefaults

`func NewPodAffinityRulesWithDefaults() *PodAffinityRules`

NewPodAffinityRulesWithDefaults instantiates a new PodAffinityRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *PodAffinityRules) GetType() PodAffinityTypeRules`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *PodAffinityRules) GetTypeOk() (*PodAffinityTypeRules, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *PodAffinityRules) SetType(v PodAffinityTypeRules)`

SetType sets Type field to given value.

### HasType

`func (o *PodAffinityRules) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *PodAffinityRules) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *PodAffinityRules) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetKey

`func (o *PodAffinityRules) GetKey() StringRules`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *PodAffinityRules) GetKeyOk() (*StringRules, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *PodAffinityRules) SetKey(v StringRules)`

SetKey sets Key field to given value.

### HasKey

`func (o *PodAffinityRules) HasKey() bool`

HasKey returns a boolean if a field has been set.

### SetKeyNil

`func (o *PodAffinityRules) SetKeyNil(b bool)`

 SetKeyNil sets the value for Key to be an explicit nil

### UnsetKey
`func (o *PodAffinityRules) UnsetKey()`

UnsetKey ensures that no value is present for Key, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


