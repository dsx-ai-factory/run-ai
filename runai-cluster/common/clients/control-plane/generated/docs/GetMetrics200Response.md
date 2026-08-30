# GetMetrics200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | **string** |  | 
**Query** | **string** |  | 
**IsRange** | **bool** |  | 
**Data** | [**[]Result**](Result.md) |  | 

## Methods

### NewGetMetrics200Response

`func NewGetMetrics200Response(status string, query string, isRange bool, data []Result, ) *GetMetrics200Response`

NewGetMetrics200Response instantiates a new GetMetrics200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetMetrics200ResponseWithDefaults

`func NewGetMetrics200ResponseWithDefaults() *GetMetrics200Response`

NewGetMetrics200ResponseWithDefaults instantiates a new GetMetrics200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *GetMetrics200Response) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GetMetrics200Response) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GetMetrics200Response) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetQuery

`func (o *GetMetrics200Response) GetQuery() string`

GetQuery returns the Query field if non-nil, zero value otherwise.

### GetQueryOk

`func (o *GetMetrics200Response) GetQueryOk() (*string, bool)`

GetQueryOk returns a tuple with the Query field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuery

`func (o *GetMetrics200Response) SetQuery(v string)`

SetQuery sets Query field to given value.


### GetIsRange

`func (o *GetMetrics200Response) GetIsRange() bool`

GetIsRange returns the IsRange field if non-nil, zero value otherwise.

### GetIsRangeOk

`func (o *GetMetrics200Response) GetIsRangeOk() (*bool, bool)`

GetIsRangeOk returns a tuple with the IsRange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRange

`func (o *GetMetrics200Response) SetIsRange(v bool)`

SetIsRange sets IsRange field to given value.


### GetData

`func (o *GetMetrics200Response) GetData() []Result`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *GetMetrics200Response) GetDataOk() (*[]Result, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *GetMetrics200Response) SetData(v []Result)`

SetData sets Data field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


