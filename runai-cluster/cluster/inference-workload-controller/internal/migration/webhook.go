package migration

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WebhookHandler struct {
	cachedWorkloadControllerUID string
}

func NewWebhookHandler(k8sClient client.Client, workloadUid string) (*WebhookHandler, error) {
	if workloadUid == "" {
		var err error
		workloadUid, err = GetInferenceWorkloadControllerUID(context.Background(), k8sClient)
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
	workloadControllerRequest, err := w.isRequesterInferenceWorkloadController(req)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !workloadControllerRequest {
		return admission.Denied("migration in progress")
	}
	return admission.Allowed("inference workload controller request")
}

func (w *WebhookHandler) isRequesterInferenceWorkloadController(req admission.Request) (bool, error) {
	workloadControllerUID := w.cachedWorkloadControllerUID
	requestUserUID := getRequesterUserId(req.UserInfo)
	return requestUserUID == workloadControllerUID, nil
}

func getServiceAccount(ctx context.Context, k8sClient client.Client, name, namespace string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{}
	serviceAccountKey := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount)
	if err != nil {
		log.Error(err, "failed getting ServiceAccount", "namespace", namespace, "name", name)
		return nil, err
	}

	return serviceAccount, nil
}

func GetInferenceWorkloadControllerUID(ctx context.Context, k8sClient client.Client) (string, error) {
	inferenceWorkloadControllerServiceAccount, err := getServiceAccount(
		ctx, k8sClient,
		constants.InferenceWorkloadControllerOperandName, constants.NamespaceRunAi)
	if err != nil {
		log.Error(err, "failed getting inference-workload-controller service account")
		return "", err
	}

	return string(inferenceWorkloadControllerServiceAccount.UID), nil
}

func getRequesterUserId(admissionRequestUserInfo authv1.UserInfo) string {
	if admissionRequestUserInfo.UID != "" {
		return admissionRequestUserInfo.UID
	}
	return admissionRequestUserInfo.Username
}
