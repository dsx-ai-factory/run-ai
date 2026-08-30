package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/steps"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewLWS(t *testing.T) {
	k8sClient := &mocks.MockK8sClients{}
	lws := NewLWS(k8sClient)
	assert.NotNil(t, lws)
	assert.Equal(t, ids.PrerequisiteIdentifierLWS, lws.GetName())
	assert.Equal(t, "0.7.0", lws.GetVersion())
}

func TestGetMandatoryInputs(t *testing.T) {
	lws := NewLWS(&mocks.MockK8sClients{})
	inputs := lws.GetMandatoryInputs(&model.RunAIInstallData{})
	assert.Empty(t, inputs)
}

func TestGetOptionalInputs(t *testing.T) {
	lws := NewLWS(&mocks.MockK8sClients{})
	inputs := lws.GetOptionalInputs(&model.RunAIInstallData{})
	assert.Empty(t, inputs)
}

func TestGetInstallSteps(t *testing.T) {
	k8sClient := &mocks.MockK8sClients{}
	lws := NewLWS(k8sClient)
	stepsList := lws.GetInstallSteps(k8sClient, nil, nil, &model.RunAIInstallData{})
	assert.Len(t, stepsList, 1)
	assert.IsType(t, &steps.LWSInstallStep{}, stepsList[0])
}

func TestInstalledOnCluster(t *testing.T) {
	fakeClient := fake.NewClientBuilder().WithObjects().Build()
	k8sClient := &mocks.MockK8sClients{}
	k8sClient.On("GetClient").Return(fakeClient).Maybe()
	k8sClient.On("GetConfig").Return(&rest.Config{}).Maybe()

	lws := NewLWS(k8sClient)
	phase := lws.InstalledOnCluster(&model.RunAIInstallData{})
	assert.IsType(t, model.InstallPhase(0), phase)
}
