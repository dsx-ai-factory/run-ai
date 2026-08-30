# ImageTags

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the image | 
**Tags** | **[]string** | The tags of the image | 

## Methods

### NewImageTags

`func NewImageTags(name string, tags []string, ) *ImageTags`

NewImageTags instantiates a new ImageTags object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImageTagsWithDefaults

`func NewImageTagsWithDefaults() *ImageTags`

NewImageTagsWithDefaults instantiates a new ImageTags object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ImageTags) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ImageTags) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ImageTags) SetName(v string)`

SetName sets Name field to given value.


### GetTags

`func (o *ImageTags) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *ImageTags) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *ImageTags) SetTags(v []string)`

SetTags sets Tags field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


