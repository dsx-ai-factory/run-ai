# ImagesListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Images** | **[]string** | The list of images. | 
**Prefix** | Pointer to **NullableString** | The prefix of the image names. | [optional] 

## Methods

### NewImagesListResponse

`func NewImagesListResponse(images []string, ) *ImagesListResponse`

NewImagesListResponse instantiates a new ImagesListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImagesListResponseWithDefaults

`func NewImagesListResponseWithDefaults() *ImagesListResponse`

NewImagesListResponseWithDefaults instantiates a new ImagesListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImages

`func (o *ImagesListResponse) GetImages() []string`

GetImages returns the Images field if non-nil, zero value otherwise.

### GetImagesOk

`func (o *ImagesListResponse) GetImagesOk() (*[]string, bool)`

GetImagesOk returns a tuple with the Images field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImages

`func (o *ImagesListResponse) SetImages(v []string)`

SetImages sets Images field to given value.


### GetPrefix

`func (o *ImagesListResponse) GetPrefix() string`

GetPrefix returns the Prefix field if non-nil, zero value otherwise.

### GetPrefixOk

`func (o *ImagesListResponse) GetPrefixOk() (*string, bool)`

GetPrefixOk returns a tuple with the Prefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrefix

`func (o *ImagesListResponse) SetPrefix(v string)`

SetPrefix sets Prefix field to given value.

### HasPrefix

`func (o *ImagesListResponse) HasPrefix() bool`

HasPrefix returns a boolean if a field has been set.

### SetPrefixNil

`func (o *ImagesListResponse) SetPrefixNil(b bool)`

 SetPrefixNil sets the value for Prefix to be an explicit nil

### UnsetPrefix
`func (o *ImagesListResponse) UnsetPrefix()`

UnsetPrefix ensures that no value is present for Prefix, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


