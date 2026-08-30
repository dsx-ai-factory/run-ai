# \InternalAuditLogsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SaveAuditLog**](InternalAuditLogsAPI.md#SaveAuditLog) | **Post** /internal/api/v1/audit/log | Save audit log



## SaveAuditLog

> SaveAuditLog(ctx).AuditLog(auditLog).Execute()

Save audit log



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
)

func main() {
	auditLog := *openapiclient.NewAuditLog("12345", "researcher@run.ai", "SubjectType_example", "Action_example", "Result_example") // AuditLog | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.InternalAuditLogsAPI.SaveAuditLog(context.Background()).AuditLog(auditLog).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAuditLogsAPI.SaveAuditLog``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSaveAuditLogRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **auditLog** | [**AuditLog**](AuditLog.md) |  | 

### Return type

 (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

