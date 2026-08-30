# S3AccessKeyAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessKeyAssetId** | Pointer to **NullableString** | ID of credentials asset of type access-key, for private S3 buckets. | [optional] 

## Methods

### NewS3AccessKeyAsset

`func NewS3AccessKeyAsset() *S3AccessKeyAsset`

NewS3AccessKeyAsset instantiates a new S3AccessKeyAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3AccessKeyAssetWithDefaults

`func NewS3AccessKeyAssetWithDefaults() *S3AccessKeyAsset`

NewS3AccessKeyAssetWithDefaults instantiates a new S3AccessKeyAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessKeyAssetId

`func (o *S3AccessKeyAsset) GetAccessKeyAssetId() string`

GetAccessKeyAssetId returns the AccessKeyAssetId field if non-nil, zero value otherwise.

### GetAccessKeyAssetIdOk

`func (o *S3AccessKeyAsset) GetAccessKeyAssetIdOk() (*string, bool)`

GetAccessKeyAssetIdOk returns a tuple with the AccessKeyAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyAssetId

`func (o *S3AccessKeyAsset) SetAccessKeyAssetId(v string)`

SetAccessKeyAssetId sets AccessKeyAssetId field to given value.

### HasAccessKeyAssetId

`func (o *S3AccessKeyAsset) HasAccessKeyAssetId() bool`

HasAccessKeyAssetId returns a boolean if a field has been set.

### SetAccessKeyAssetIdNil

`func (o *S3AccessKeyAsset) SetAccessKeyAssetIdNil(b bool)`

 SetAccessKeyAssetIdNil sets the value for AccessKeyAssetId to be an explicit nil

### UnsetAccessKeyAssetId
`func (o *S3AccessKeyAsset) UnsetAccessKeyAssetId()`

UnsetAccessKeyAssetId ensures that no value is present for AccessKeyAssetId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


