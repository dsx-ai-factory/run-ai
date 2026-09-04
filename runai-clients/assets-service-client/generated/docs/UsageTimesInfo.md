# UsageTimesInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastUsedByWorkload** | Pointer to **NullableTime** | The time at which the asset were last used by a workload of given type and project. This field requires complyToProject and complyToWorkload types flags of the API, to specify the workload type and project id. | [optional] 

## Methods

### NewUsageTimesInfo

`func NewUsageTimesInfo() *UsageTimesInfo`

NewUsageTimesInfo instantiates a new UsageTimesInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUsageTimesInfoWithDefaults

`func NewUsageTimesInfoWithDefaults() *UsageTimesInfo`

NewUsageTimesInfoWithDefaults instantiates a new UsageTimesInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastUsedByWorkload

`func (o *UsageTimesInfo) GetLastUsedByWorkload() time.Time`

GetLastUsedByWorkload returns the LastUsedByWorkload field if non-nil, zero value otherwise.

### GetLastUsedByWorkloadOk

`func (o *UsageTimesInfo) GetLastUsedByWorkloadOk() (*time.Time, bool)`

GetLastUsedByWorkloadOk returns a tuple with the LastUsedByWorkload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUsedByWorkload

`func (o *UsageTimesInfo) SetLastUsedByWorkload(v time.Time)`

SetLastUsedByWorkload sets LastUsedByWorkload field to given value.

### HasLastUsedByWorkload

`func (o *UsageTimesInfo) HasLastUsedByWorkload() bool`

HasLastUsedByWorkload returns a boolean if a field has been set.

### SetLastUsedByWorkloadNil

`func (o *UsageTimesInfo) SetLastUsedByWorkloadNil(b bool)`

 SetLastUsedByWorkloadNil sets the value for LastUsedByWorkload to be an explicit nil

### UnsetLastUsedByWorkload
`func (o *UsageTimesInfo) UnsetLastUsedByWorkload()`

UnsetLastUsedByWorkload ensures that no value is present for LastUsedByWorkload, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


