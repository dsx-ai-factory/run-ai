package tests

import (
	"context"
	"flag"

	"github.com/rs/zerolog"
	"go.uber.org/zap/zapcore"

	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
//
// The suite runs against a fake (in-memory) controller-runtime client rather
// than envtest. There is no controller manager and no cache firing events, so
// reconciliation is driven manually: reconcileAllAccessRules() stands in for
// the watch fan-out (AccessRule/RoleBinding/ClusterRoleBinding/Project events
// all map back to AccessRule reconcile requests). The event-filter and mapper
// logic that the manager wires up is covered directly by unit tests in
// pkg/controllers/accessrule. This makes the suite a plain `go test` package.
const (
	suite = "AccessRuleController Tests"

	// reconcilePasses is the number of times reconcileAllAccessRules sweeps all
	// rules per call. Two passes reach a fixed point for cross-rule cases where
	// one rule's reconcile changes another rule's inputs (e.g. a shared
	// RoleBinding), which the manager would converge to via repeated events.
	reconcilePasses = 2
)

var (
	// k8sClient, cachedClient and basicK8sClient all point at the same fake
	// client; the three names are kept so the existing helpers read naturally.
	k8sClient      client.Client
	cachedClient   client.Client
	basicK8sClient client.Client
	apiCtx         context.Context
	arController   *accessrule.AccessRuleController
	scheme         = scheme_init.Scheme()
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	opts := zap.Options{
		Development:     true,
		StacktraceLevel: zapcore.LevelEnabler(zapcore.FatalLevel),
		DestWriter:      GinkgoWriter,
		TimeEncoder:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func testSuiteSetup() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	initLogging()

	apiCtx = context.Background()

	// register the kai.resources types on the test scheme here
	// unconditionally so the suite (and its fixtures that create both legacy and kai org units)
	// passes under both compile-time values of the flag.
	Expect(kaiorg.AddToScheme(scheme)).To(Succeed())

	// Status subresources are declared so the controller's status writes and the
	// tests' Project status updates persist without clobbering spec. The field
	// indexes mirror those installed by AccessRuleController.indexFields for both
	// the legacy and kai worlds, so the reconcile path's field-selector lists
	// resolve as in production regardless of the flag value.
	fakeClient := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.AccessRule{}, &v2.Project{}, &kaiorg.Project{}).
		WithIndex(&kaiorg.Project{}, common.ProjectsParentNameField, accessrule.ProjectsParentIndexer).
		WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleProjectNameField, accessrule.AccessRulesProjectIndexer).
		WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleDepartmentNameField, accessrule.AccessRulesDepartmentIndexer).
		WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleClusterField, accessrule.AccessRulesClusterIndexer).
		Build()

	k8sClient = fakeClient
	cachedClient = fakeClient
	basicK8sClient = fakeClient
}

// controllerSetup (re)creates the controller wired to the fake client.
func controllerSetup() {
	arController = accessrule.NewAccessRuleController(k8sClient)
}

// reconcile drives a single manual reconciliation of one access rule. Errors
// are swallowed: in production Reconcile requeues on error (and sets a NotReady
// status before returning), and tests assert object state, not the return value.
func reconcile(accessRuleName string) {
	_, _ = arController.Reconcile(apiCtx, ctrl.Request{NamespacedName: types.NamespacedName{Name: accessRuleName}})
}

// reconcileAllAccessRules reconciles every access rule currently in the cluster,
// a deterministic stand-in for the manager's watch fan-out. It sweeps all rules
// reconcilePasses times so cross-rule dependencies converge. Rules pending
// deletion are reconciled too, so their finalizer is removed and they are
// actually deleted (the fake client keeps finalized objects until then).
func reconcileAllAccessRules() {
	for pass := 0; pass < reconcilePasses; pass++ {
		accessRules := listAccessRules(k8sClient)
		if accessRules == nil {
			return
		}
		for i := range accessRules.Items {
			reconcile(accessRules.Items[i].Name)
		}
	}
}

var _ = BeforeSuite(testSuiteSetup)
