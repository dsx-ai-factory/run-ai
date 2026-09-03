# AssetClusterStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**Status** | [**NullableAssetStatus**](AssetStatus.md) |  | 
**ErrorMessage** | Pointer to **NullableString** | Error message for the case that the asset is not available in this cluster. | [optional] 

## Methods

### NewAssetClusterStatus

`func NewAssetClusterStatus(clusterId string, status NullableAssetStatus, ) *AssetClusterStatus`

NewAssetClusterStatus instantiates a new AssetClusterStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetClusterStatusWithDefaults

`func NewAssetClusterStatusWithDefaults() *AssetClusterStatus`

NewAssetClusterStatusWithDefaults instantiates a new AssetClusterStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *AssetClusterStatus) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AssetClusterStatus) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AssetClusterStatus) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetStatus

`func (o *AssetClusterStatus) GetStatus() AssetStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *AssetClusterStatus) GetStatusOk() (*AssetStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *AssetClusterStatus) SetStatus(v AssetStatus)`

SetStatus sets Status field to given value.


### SetStatusNil

`func (o *AssetClusterStatus) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *AssetClusterStatus) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetErrorMessage

`func (o *AssetClusterStatus) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *AssetClusterStatus) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *AssetClusterStatus) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *AssetClusterStatus) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.

### SetErrorMessageNil

`func (o *AssetClusterStatus) SetErrorMessageNil(b bool)`

 SetErrorMessageNil sets the value for ErrorMessage to be an explicit nil

### UnsetErrorMessage
`func (o *AssetClusterStatus) UnsetErrorMessage()`

UnsetErrorMessage ensures that no value is present for ErrorMessage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


