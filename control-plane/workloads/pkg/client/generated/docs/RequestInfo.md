# RequestInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AfterUpdate** | **int64** | the value of an update counter of a data in the system. | 

## Methods

### NewRequestInfo

`func NewRequestInfo(afterUpdate int64, ) *RequestInfo`

NewRequestInfo instantiates a new RequestInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRequestInfoWithDefaults

`func NewRequestInfoWithDefaults() *RequestInfo`

NewRequestInfoWithDefaults instantiates a new RequestInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAfterUpdate

`func (o *RequestInfo) GetAfterUpdate() int64`

GetAfterUpdate returns the AfterUpdate field if non-nil, zero value otherwise.

### GetAfterUpdateOk

`func (o *RequestInfo) GetAfterUpdateOk() (*int64, bool)`

GetAfterUpdateOk returns a tuple with the AfterUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAfterUpdate

`func (o *RequestInfo) SetAfterUpdate(v int64)`

SetAfterUpdate sets AfterUpdate field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


