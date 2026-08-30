package helm_deployer

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/utils/ptr"
)

var _ = Describe("AppStatus", func() {
	Describe("ApplyOnApplication", func() {
		// Regression: concurrent status updates could leave k8sPhase stale (e.g. pending-install)
		// while helmStatus.helmPhase was already deployed. ApplyOnApplication must sync k8sPhase
		// from the Helm release phase whenever we apply release info.
		Context("syncing k8sPhase with Helm release phase", func() {
			var deployedRelease *release.Release

			BeforeEach(func() {
				deployedRelease = &release.Release{
					Version: 1,
					Info: &release.Info{
						Status: release.StatusDeployed,
					},
				}
			})

			It("updates stale k8sPhase when helmStatus already shows deployed", func() {
				app := &v1alpha1.Application{
					Status: v1alpha1.ApplicationStatus{
						K8sPhase: string(release.StatusPendingInstall),
						HelmStatus: &v1alpha1.HelmStatus{
							Phase:          string(release.StatusDeployed),
							ReleaseVersion: ptr.To(1),
						},
					},
				}

				changed := NewAppStatus().WithReleaseInfo(deployedRelease).ApplyOnApplication(app)

				Expect(changed).To(BeTrue())
				Expect(app.Status.K8sPhase).To(Equal(string(release.StatusDeployed)))
			})

			It("updates both helmStatus and k8sPhase from pending-install to deployed", func() {
				app := &v1alpha1.Application{
					Status: v1alpha1.ApplicationStatus{
						K8sPhase: string(release.StatusPendingInstall),
						HelmStatus: &v1alpha1.HelmStatus{
							Phase: string(release.StatusPendingInstall),
						},
					},
				}

				changed := NewAppStatus().WithReleaseInfo(deployedRelease).ApplyOnApplication(app)

				Expect(changed).To(BeTrue())
				Expect(app.Status.HelmStatus.Phase).To(Equal(string(release.StatusDeployed)))
				Expect(app.Status.K8sPhase).To(Equal(string(release.StatusDeployed)))
			})

			It("reports no change when k8sPhase and helmStatus already match the release", func() {
				app := &v1alpha1.Application{
					Status: v1alpha1.ApplicationStatus{
						K8sPhase: string(release.StatusDeployed),
						HelmStatus: &v1alpha1.HelmStatus{
							Phase:          string(release.StatusDeployed),
							ReleaseVersion: ptr.To(1),
						},
					},
				}

				changed := NewAppStatus().WithReleaseInfo(deployedRelease).ApplyOnApplication(app)

				Expect(changed).To(BeFalse())
			})
		})
	})
})
