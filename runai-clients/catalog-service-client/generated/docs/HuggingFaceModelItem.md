# HuggingFaceModelItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | the huggingface model id | [optional] 
**Gated** | Pointer to **bool** |  | [optional] 
**Downloads** | Pointer to **int32** |  | [optional] 
**Trending** | Pointer to **bool** | indicates if the model is trending | [optional] 

## Methods

### NewHuggingFaceModelItem

`func NewHuggingFaceModelItem() *HuggingFaceModelItem`

NewHuggingFaceModelItem instantiates a new HuggingFaceModelItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHuggingFaceModelItemWithDefaults

`func NewHuggingFaceModelItemWithDefaults() *HuggingFaceModelItem`

NewHuggingFaceModelItemWithDefaults instantiates a new HuggingFaceModelItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *HuggingFaceModelItem) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *HuggingFaceModelItem) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *HuggingFaceModelItem) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *HuggingFaceModelItem) HasId() bool`

HasId returns a boolean if a field has been set.

### GetGated

`func (o *HuggingFaceModelItem) GetGated() bool`

GetGated returns the Gated field if non-nil, zero value otherwise.

### GetGatedOk

`func (o *HuggingFaceModelItem) GetGatedOk() (*bool, bool)`

GetGatedOk returns a tuple with the Gated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGated

`func (o *HuggingFaceModelItem) SetGated(v bool)`

SetGated sets Gated field to given value.

### HasGated

`func (o *HuggingFaceModelItem) HasGated() bool`

HasGated returns a boolean if a field has been set.

### GetDownloads

`func (o *HuggingFaceModelItem) GetDownloads() int32`

GetDownloads returns the Downloads field if non-nil, zero value otherwise.

### GetDownloadsOk

`func (o *HuggingFaceModelItem) GetDownloadsOk() (*int32, bool)`

GetDownloadsOk returns a tuple with the Downloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDownloads

`func (o *HuggingFaceModelItem) SetDownloads(v int32)`

SetDownloads sets Downloads field to given value.

### HasDownloads

`func (o *HuggingFaceModelItem) HasDownloads() bool`

HasDownloads returns a boolean if a field has been set.

### GetTrending

`func (o *HuggingFaceModelItem) GetTrending() bool`

GetTrending returns the Trending field if non-nil, zero value otherwise.

### GetTrendingOk

`func (o *HuggingFaceModelItem) GetTrendingOk() (*bool, bool)`

GetTrendingOk returns a tuple with the Trending field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrending

`func (o *HuggingFaceModelItem) SetTrending(v bool)`

SetTrending sets Trending field to given value.

### HasTrending

`func (o *HuggingFaceModelItem) HasTrending() bool`

HasTrending returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


