# S3ListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Entries** | [**[]S3Asset**](S3Asset.md) |  | 

## Methods

### NewS3ListResponse

`func NewS3ListResponse(entries []S3Asset, ) *S3ListResponse`

NewS3ListResponse instantiates a new S3ListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3ListResponseWithDefaults

`func NewS3ListResponseWithDefaults() *S3ListResponse`

NewS3ListResponseWithDefaults instantiates a new S3ListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntries

`func (o *S3ListResponse) GetEntries() []S3Asset`

GetEntries returns the Entries field if non-nil, zero value otherwise.

### GetEntriesOk

`func (o *S3ListResponse) GetEntriesOk() (*[]S3Asset, bool)`

GetEntriesOk returns a tuple with the Entries field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntries

`func (o *S3ListResponse) SetEntries(v []S3Asset)`

SetEntries sets Entries field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


