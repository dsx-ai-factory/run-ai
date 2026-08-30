package statusmsg

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func GenerateStatus(err error) v1.Status {
	return v1.Status{
		Status:  "Failure",
		Message: err.Error(),
		Reason:  v1.StatusReasonUnknown,
	}
}
