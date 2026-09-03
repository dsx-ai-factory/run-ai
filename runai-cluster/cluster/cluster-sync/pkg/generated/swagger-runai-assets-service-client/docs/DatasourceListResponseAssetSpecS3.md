# DatasourceListResponseAssetSpecS3

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bucket** | **NullableString** | The name of the bucket Bucket name is mandatory for creating a workspace. | 
**Path** | **NullableString** | Local path within the workspace to which the S3 bucket will be mapped. Path is mandatory for creating a workspace. | 
**AccessKeyAssetId** | Pointer to **NullableString** | ID of credentials asset of type access-key, for private S3 buckets. | [optional] 
**Url** | Pointer to **NullableString** | The url of the S3 service provider. The default is the URL of the Amazon AWS S3 service. | [optional] 

## Methods

### NewDatasourceListResponseAssetSpecS3

`func NewDatasourceListResponseAssetSpecS3(bucket NullableString, path NullableString, ) *DatasourceListResponseAssetSpecS3`

NewDatasourceListResponseAssetSpecS3 instantiates a new DatasourceListResponseAssetSpecS3 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecS3WithDefaults

`func NewDatasourceListResponseAssetSpecS3WithDefaults() *DatasourceListResponseAssetSpecS3`

NewDatasourceListResponseAssetSpecS3WithDefaults instantiates a new DatasourceListResponseAssetSpecS3 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBucket

`func (o *DatasourceListResponseAssetSpecS3) GetBucket() string`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *DatasourceListResponseAssetSpecS3) GetBucketOk() (*string, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *DatasourceListResponseAssetSpecS3) SetBucket(v string)`

SetBucket sets Bucket field to given value.


### SetBucketNil

`func (o *DatasourceListResponseAssetSpecS3) SetBucketNil(b bool)`

 SetBucketNil sets the value for Bucket to be an explicit nil

### UnsetBucket
`func (o *DatasourceListResponseAssetSpecS3) UnsetBucket()`

UnsetBucket ensures that no value is present for Bucket, not even an explicit nil
### GetPath

`func (o *DatasourceListResponseAssetSpecS3) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *DatasourceListResponseAssetSpecS3) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *DatasourceListResponseAssetSpecS3) SetPath(v string)`

SetPath sets Path field to given value.


### SetPathNil

`func (o *DatasourceListResponseAssetSpecS3) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *DatasourceListResponseAssetSpecS3) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetAccessKeyAssetId

`func (o *DatasourceListResponseAssetSpecS3) GetAccessKeyAssetId() string`

GetAccessKeyAssetId returns the AccessKeyAssetId field if non-nil, zero value otherwise.

### GetAccessKeyAssetIdOk

`func (o *DatasourceListResponseAssetSpecS3) GetAccessKeyAssetIdOk() (*string, bool)`

GetAccessKeyAssetIdOk returns a tuple with the AccessKeyAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyAssetId

`func (o *DatasourceListResponseAssetSpecS3) SetAccessKeyAssetId(v string)`

SetAccessKeyAssetId sets AccessKeyAssetId field to given value.

### HasAccessKeyAssetId

`func (o *DatasourceListResponseAssetSpecS3) HasAccessKeyAssetId() bool`

HasAccessKeyAssetId returns a boolean if a field has been set.

### SetAccessKeyAssetIdNil

`func (o *DatasourceListResponseAssetSpecS3) SetAccessKeyAssetIdNil(b bool)`

 SetAccessKeyAssetIdNil sets the value for AccessKeyAssetId to be an explicit nil

### UnsetAccessKeyAssetId
`func (o *DatasourceListResponseAssetSpecS3) UnsetAccessKeyAssetId()`

UnsetAccessKeyAssetId ensures that no value is present for AccessKeyAssetId, not even an explicit nil
### GetUrl

`func (o *DatasourceListResponseAssetSpecS3) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *DatasourceListResponseAssetSpecS3) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *DatasourceListResponseAssetSpecS3) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *DatasourceListResponseAssetSpecS3) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### SetUrlNil

`func (o *DatasourceListResponseAssetSpecS3) SetUrlNil(b bool)`

 SetUrlNil sets the value for Url to be an explicit nil

### UnsetUrl
`func (o *DatasourceListResponseAssetSpecS3) UnsetUrl()`

UnsetUrl ensures that no value is present for Url, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


