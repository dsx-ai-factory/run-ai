# ImagePullPolicyOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Options** | Pointer to [**[]ImagePullPolicyOptionsOptionsInner**](ImagePullPolicyOptionsOptionsInner.md) | Limit the permitted values for imagePullPolicy to a given set of values. | [optional] 

## Methods

### NewImagePullPolicyOptions

`func NewImagePullPolicyOptions() *ImagePullPolicyOptions`

NewImagePullPolicyOptions instantiates a new ImagePullPolicyOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImagePullPolicyOptionsWithDefaults

`func NewImagePullPolicyOptionsWithDefaults() *ImagePullPolicyOptions`

NewImagePullPolicyOptionsWithDefaults instantiates a new ImagePullPolicyOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOptions

`func (o *ImagePullPolicyOptions) GetOptions() []ImagePullPolicyOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *ImagePullPolicyOptions) GetOptionsOk() (*[]ImagePullPolicyOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *ImagePullPolicyOptions) SetOptions(v []ImagePullPolicyOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *ImagePullPolicyOptions) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *ImagePullPolicyOptions) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *ImagePullPolicyOptions) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


