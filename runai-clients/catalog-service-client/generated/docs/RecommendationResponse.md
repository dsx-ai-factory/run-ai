# RecommendationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Image** | **string** | Untagged NIM image reference supplied in the request. | 
**Tag** | **string** | Exact NIM image tag supplied in the request. | 
**Results** | [**[]RecommendationResult**](RecommendationResult.md) | One result per evaluated GPU type and per-device memory capacity. Empty when the cluster has no reported GPUs and no GPU type was requested. | 

## Methods

### NewRecommendationResponse

`func NewRecommendationResponse(image string, tag string, results []RecommendationResult, ) *RecommendationResponse`

NewRecommendationResponse instantiates a new RecommendationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRecommendationResponseWithDefaults

`func NewRecommendationResponseWithDefaults() *RecommendationResponse`

NewRecommendationResponseWithDefaults instantiates a new RecommendationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImage

`func (o *RecommendationResponse) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *RecommendationResponse) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *RecommendationResponse) SetImage(v string)`

SetImage sets Image field to given value.


### GetTag

`func (o *RecommendationResponse) GetTag() string`

GetTag returns the Tag field if non-nil, zero value otherwise.

### GetTagOk

`func (o *RecommendationResponse) GetTagOk() (*string, bool)`

GetTagOk returns a tuple with the Tag field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTag

`func (o *RecommendationResponse) SetTag(v string)`

SetTag sets Tag field to given value.


### GetResults

`func (o *RecommendationResponse) GetResults() []RecommendationResult`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *RecommendationResponse) GetResultsOk() (*[]RecommendationResult, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *RecommendationResponse) SetResults(v []RecommendationResult)`

SetResults sets Results field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


