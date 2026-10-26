# TelemetryResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | specifies what data returned | 
**Timestamp** | **time.Time** |  | 
**Values** | [**[]TelemetryResponseValuesInner**](TelemetryResponseValuesInner.md) |  | 

## Methods

### NewTelemetryResponse

`func NewTelemetryResponse(type_ string, timestamp time.Time, values []TelemetryResponseValuesInner, ) *TelemetryResponse`

NewTelemetryResponse instantiates a new TelemetryResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTelemetryResponseWithDefaults

`func NewTelemetryResponseWithDefaults() *TelemetryResponse`

NewTelemetryResponseWithDefaults instantiates a new TelemetryResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *TelemetryResponse) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *TelemetryResponse) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *TelemetryResponse) SetType(v string)`

SetType sets Type field to given value.


### GetTimestamp

`func (o *TelemetryResponse) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *TelemetryResponse) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *TelemetryResponse) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.


### GetValues

`func (o *TelemetryResponse) GetValues() []TelemetryResponseValuesInner`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *TelemetryResponse) GetValuesOk() (*[]TelemetryResponseValuesInner, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *TelemetryResponse) SetValues(v []TelemetryResponseValuesInner)`

SetValues sets Values field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


