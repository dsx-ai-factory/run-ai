# CountWorkloadElements200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StoredElementsCount** | **int32** | The number of structure elements stored for this workload. When truncation occurs, this will be less than totalElementsCount. | 
**TotalElementsCount** | **int32** | The total number of structure elements that exist for this workload. When no truncation occurred, this equals storedElementsCount. | 

## Methods

### NewCountWorkloadElements200Response

`func NewCountWorkloadElements200Response(storedElementsCount int32, totalElementsCount int32, ) *CountWorkloadElements200Response`

NewCountWorkloadElements200Response instantiates a new CountWorkloadElements200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCountWorkloadElements200ResponseWithDefaults

`func NewCountWorkloadElements200ResponseWithDefaults() *CountWorkloadElements200Response`

NewCountWorkloadElements200ResponseWithDefaults instantiates a new CountWorkloadElements200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStoredElementsCount

`func (o *CountWorkloadElements200Response) GetStoredElementsCount() int32`

GetStoredElementsCount returns the StoredElementsCount field if non-nil, zero value otherwise.

### GetStoredElementsCountOk

`func (o *CountWorkloadElements200Response) GetStoredElementsCountOk() (*int32, bool)`

GetStoredElementsCountOk returns a tuple with the StoredElementsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStoredElementsCount

`func (o *CountWorkloadElements200Response) SetStoredElementsCount(v int32)`

SetStoredElementsCount sets StoredElementsCount field to given value.


### GetTotalElementsCount

`func (o *CountWorkloadElements200Response) GetTotalElementsCount() int32`

GetTotalElementsCount returns the TotalElementsCount field if non-nil, zero value otherwise.

### GetTotalElementsCountOk

`func (o *CountWorkloadElements200Response) GetTotalElementsCountOk() (*int32, bool)`

GetTotalElementsCountOk returns a tuple with the TotalElementsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalElementsCount

`func (o *CountWorkloadElements200Response) SetTotalElementsCount(v int32)`

SetTotalElementsCount sets TotalElementsCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


