package migration

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/webhook"
	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WebhookHandler struct {
	cachedWorkloadControllerUID string
}

func NewWebhookHandler(k8sClient client.Client, workloadUid string) (*WebhookHandler, error) {
	if workloadUid == "" {
		var err error
		workloadUid, err = webhook.GetWorkloadControllerUID(context.Background(), k8sClient)
		if err != nil {
			return nil, err
		}
	}
	return &WebhookHandler{cachedWorkloadControllerUID: workloadUid}, nil
}

func (w *WebhookHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.Operation != admissionv1.Update {
		return admission.Denied("migration in progress")
	}
	workloadControllerRequest, err := w.isWorkloadControllerReq(req)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !workloadControllerRequest {
		return admission.Denied("migration in progress")
	}
	return admission.Allowed("workload controller request")
}

func (w *WebhookHandler) isWorkloadControllerReq(req admission.Request) (bool, error) {
	workloadControllerUID := w.cachedWorkloadControllerUID
	requestUserUID := webhook.GetRequesterUserId(req.UserInfo)
	return requestUserUID == workloadControllerUID, nil
}
