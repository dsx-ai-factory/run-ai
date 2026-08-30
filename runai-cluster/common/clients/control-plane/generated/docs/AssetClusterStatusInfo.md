# AssetClusterStatusInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to [**AssetSyncStatusEnum**](AssetSyncStatusEnum.md) |  | [optional] 
**Issues** | Pointer to [**[]AssetClusterStatusIssue**](AssetClusterStatusIssue.md) |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Url** | Pointer to **string** |  | [optional] 

## Methods

### NewAssetClusterStatusInfo

`func NewAssetClusterStatusInfo() *AssetClusterStatusInfo`

NewAssetClusterStatusInfo instantiates a new AssetClusterStatusInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetClusterStatusInfoWithDefaults

`func NewAssetClusterStatusInfoWithDefaults() *AssetClusterStatusInfo`

NewAssetClusterStatusInfoWithDefaults instantiates a new AssetClusterStatusInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *AssetClusterStatusInfo) GetStatus() AssetSyncStatusEnum`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *AssetClusterStatusInfo) GetStatusOk() (*AssetSyncStatusEnum, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *AssetClusterStatusInfo) SetStatus(v AssetSyncStatusEnum)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *AssetClusterStatusInfo) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetIssues

`func (o *AssetClusterStatusInfo) GetIssues() []AssetClusterStatusIssue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *AssetClusterStatusInfo) GetIssuesOk() (*[]AssetClusterStatusIssue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *AssetClusterStatusInfo) SetIssues(v []AssetClusterStatusIssue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *AssetClusterStatusInfo) HasIssues() bool`

HasIssues returns a boolean if a field has been set.

### GetMessage

`func (o *AssetClusterStatusInfo) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AssetClusterStatusInfo) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AssetClusterStatusInfo) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *AssetClusterStatusInfo) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetUrl

`func (o *AssetClusterStatusInfo) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *AssetClusterStatusInfo) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *AssetClusterStatusInfo) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *AssetClusterStatusInfo) HasUrl() bool`

HasUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


