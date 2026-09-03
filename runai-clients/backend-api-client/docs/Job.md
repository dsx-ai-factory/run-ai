# Job

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | [**JobMetadata**](JobMetadata.md) |  | 
**Current** | [**JobCurrent**](JobCurrent.md) |  | 
**TimeRange** | Pointer to [**JobTimeRange**](JobTimeRange.md) |  | [optional] 

## Methods

### NewJob

`func NewJob(metadata JobMetadata, current JobCurrent, ) *Job`

NewJob instantiates a new Job object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobWithDefaults

`func NewJobWithDefaults() *Job`

NewJobWithDefaults instantiates a new Job object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *Job) GetMetadata() JobMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *Job) GetMetadataOk() (*JobMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *Job) SetMetadata(v JobMetadata)`

SetMetadata sets Metadata field to given value.


### GetCurrent

`func (o *Job) GetCurrent() JobCurrent`

GetCurrent returns the Current field if non-nil, zero value otherwise.

### GetCurrentOk

`func (o *Job) GetCurrentOk() (*JobCurrent, bool)`

GetCurrentOk returns a tuple with the Current field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrent

`func (o *Job) SetCurrent(v JobCurrent)`

SetCurrent sets Current field to given value.


### GetTimeRange

`func (o *Job) GetTimeRange() JobTimeRange`

GetTimeRange returns the TimeRange field if non-nil, zero value otherwise.

### GetTimeRangeOk

`func (o *Job) GetTimeRangeOk() (*JobTimeRange, bool)`

GetTimeRangeOk returns a tuple with the TimeRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeRange

`func (o *Job) SetTimeRange(v JobTimeRange)`

SetTimeRange sets TimeRange field to given value.

### HasTimeRange

`func (o *Job) HasTimeRange() bool`

HasTimeRange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


