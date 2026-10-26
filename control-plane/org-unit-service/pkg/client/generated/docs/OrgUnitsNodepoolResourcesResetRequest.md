# OrgUnitsNodepoolResourcesResetRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**NodePoolId** | **string** | The node pool id to which the request applies | 
**Action** | [**NodepoolResourcesResetAction**](NodepoolResourcesResetAction.md) |  | 

## Methods

### NewOrgUnitsNodepoolResourcesResetRequest

`func NewOrgUnitsNodepoolResourcesResetRequest(clusterId string, nodePoolId string, action NodepoolResourcesResetAction, ) *OrgUnitsNodepoolResourcesResetRequest`

NewOrgUnitsNodepoolResourcesResetRequest instantiates a new OrgUnitsNodepoolResourcesResetRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgUnitsNodepoolResourcesResetRequestWithDefaults

`func NewOrgUnitsNodepoolResourcesResetRequestWithDefaults() *OrgUnitsNodepoolResourcesResetRequest`

NewOrgUnitsNodepoolResourcesResetRequestWithDefaults instantiates a new OrgUnitsNodepoolResourcesResetRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *OrgUnitsNodepoolResourcesResetRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *OrgUnitsNodepoolResourcesResetRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *OrgUnitsNodepoolResourcesResetRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetNodePoolId

`func (o *OrgUnitsNodepoolResourcesResetRequest) GetNodePoolId() string`

GetNodePoolId returns the NodePoolId field if non-nil, zero value otherwise.

### GetNodePoolIdOk

`func (o *OrgUnitsNodepoolResourcesResetRequest) GetNodePoolIdOk() (*string, bool)`

GetNodePoolIdOk returns a tuple with the NodePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolId

`func (o *OrgUnitsNodepoolResourcesResetRequest) SetNodePoolId(v string)`

SetNodePoolId sets NodePoolId field to given value.


### GetAction

`func (o *OrgUnitsNodepoolResourcesResetRequest) GetAction() NodepoolResourcesResetAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *OrgUnitsNodepoolResourcesResetRequest) GetActionOk() (*NodepoolResourcesResetAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *OrgUnitsNodepoolResourcesResetRequest) SetAction(v NodepoolResourcesResetAction)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


