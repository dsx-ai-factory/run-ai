package legacyprojectpermissions

import (
	"context"
	"fmt"

	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestProjectLegacyPermissionsHandler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ProjectLegacyPermissionsHandler Suite")
}

var _ = Describe("ProjectLegacyPermissionsHandler integration with access rules", Ordered, func() {
	var (
		handler         *ProjectLegacyPermissionsHandler
		k8sClient       client.Client
		projectNames    []string
		accessRuleNames []string
		userMap         map[string]map[string]bool // project -> user
		groupMap        map[string]map[string]bool // project -> group
		departments     []string
		projectToDept   map[string]string
	)

	BeforeAll(func() {
		k8sClient = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).
			WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleProjectNameField, accessrule.AccessRulesProjectIndexer).
			WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleDepartmentNameField, accessrule.AccessRulesDepartmentIndexer).
			WithIndex(&v1alpha1.AccessRule{}, common.AccessRuleClusterField, accessrule.AccessRulesClusterIndexer).
			Build()

		departments = []string{"dept-0", "dept-1", "dept-2", "dept-3"}
		projectNames = []string{}
		userMap = map[string]map[string]bool{}
		groupMap = map[string]map[string]bool{}
		projectToDept = map[string]string{}
		// Distribute 20 projects evenly among 4 departments
		for i := 0; i < 20; i++ {
			p := fmt.Sprintf("proj-%d", i)
			dept := departments[i%len(departments)]
			projectNames = append(projectNames, p)
			createProject(p, dept, k8sClient)
			userMap[p] = map[string]bool{}
			groupMap[p] = map[string]bool{}
			projectToDept[p] = dept
		}
		accessRuleNames = []string{}
		// Create 100 access rules, rotating subject type and scope
		for i := 0; i < 100; i++ {
			subjectType := v1alpha1.UserSubject
			subjectId := fmt.Sprintf("user-%d", i%10)
			isUser := true
			if i%4 == 1 {
				subjectType = v1alpha1.GroupSubject
				subjectId = fmt.Sprintf("group-%d", i%10)
				isUser = false
			} else if i%4 == 2 {
				subjectType = v1alpha1.AppSubject
				subjectId = fmt.Sprintf("app-%d", i%10)
				isUser = true // apps are treated as users
			} else if i%4 == 3 {
				subjectType = v1alpha1.ServiceAccountSubject
				subjectId = fmt.Sprintf("service-account-%d", i%10)
				isUser = true // service-accounts are treated as users
			}
			// Rotate scope
			scopeIdx := (i / 4) % 3
			proj := projectNames[i%len(projectNames)]
			dept := projectToDept[proj]
			arName := fmt.Sprintf("ar-%d", i)
			var scopeType v1alpha1.ScopeType
			var scopeName string
			var roleName string
			if scopeIdx == 0 {
				scopeType = v1alpha1.ProjectScope
				scopeName = proj
				roleName = "role-proj"
			} else if scopeIdx == 1 {
				scopeType = v1alpha1.DepartmentScope
				scopeName = dept
				roleName = "role-dept"
			} else {
				scopeType = v1alpha1.ClusterScope
				scopeName = ""
				roleName = "role-cluster"
			}
			accessRuleNames = append(accessRuleNames, arName)
			createAccessRule(arName, &v1alpha1.AccessRuleSpec{
				SubjectId:   subjectId,
				SubjectType: subjectType,
				ScopeName:   scopeName,
				ScopeType:   scopeType,
				RoleName:    roleName,
			}, k8sClient)
			if scopeType == v1alpha1.ProjectScope {
				if isUser {
					userMap[proj][subjectId] = true
				} else {
					groupMap[proj][subjectId] = true
				}
			} else if scopeType == v1alpha1.DepartmentScope {
				for _, p := range projectNames {
					if projectToDept[p] == dept {
						if isUser {
							userMap[p][subjectId] = true
						} else {
							groupMap[p][subjectId] = true
						}
					}
				}
			} else if scopeType == v1alpha1.ClusterScope {
				for _, p := range projectNames {
					if isUser {
						userMap[p][subjectId] = true
					} else {
						groupMap[p][subjectId] = true
					}
				}
			}
		}
	})

	AfterAll(func() {
		for _, p := range projectNames {
			deleteProject(p, k8sClient)
		}
		for _, ar := range accessRuleNames {
			deleteAccessRule(ar, k8sClient)
		}
	})

	It("should update project permissions based on access rules", func() {
		handler = NewProjectLegacyPermissionsHandler(k8sClient, 1) // 1 second interval for fast test
		stopCh := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		handler.Start(ctx, stopCh)
		DeferCleanup(func() { close(stopCh) })

		for _, projName := range projectNames {
			users := userMap[projName]
			groups := groupMap[projName]
			Eventually(func(g Gomega) {
				p := getProjectFromClient(projName, k8sClient)
				g.Expect(p).ToNot(BeNil())
				g.Expect(p.Spec.Permissions.Users).To(ConsistOf(keys(users)))
				g.Expect(p.Spec.Permissions.Groups).To(ConsistOf(keys(groups)))
			}, 20, 0.5).Should(Succeed(), fmt.Sprintf("Project %s permissions did not match expected values", projName))
		}
	})
})

func getProjectFromClient(projectName string, k8sClient client.Client) *v2.Project {
	projectKey := types.NamespacedName{Name: projectName}
	projectObj := &v2.Project{}
	err := k8sClient.Get(context.Background(), projectKey, projectObj)
	if err != nil {
		return nil
	}
	return projectObj
}

func createProject(projectName, departmentName string, k8sClient client.Client) {
	projectObj := &v2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			UID:  types.UID(projectName),
		},
		Spec: v2.ProjectSpec{
			Department: departmentName,
			Queues: []v2.QueueProjectSpec{
				{
					Name: projectName,
				},
			},
		},
	}

	utils.ExpectCreateResource(k8sClient, projectObj)
}

func deleteProject(project string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting project <%s>", project))

	projectObj := getProjectFromClient(project, k8sClient)
	if projectObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, projectObj)
	}
}

func createAccessRule(accessRuleName string, accessRuleSpec *v1alpha1.AccessRuleSpec, k8sClient client.Client) *v1alpha1.AccessRule {
	accessRuleObj := &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:   accessRuleName,
			UID:    types.UID(accessRuleName),
			Labels: map[string]string{},
		},
		Spec: *accessRuleSpec,
	}

	// Set correct label for field selectors used in GetAllAccessRulesForProject
	switch accessRuleSpec.ScopeType {
	case v1alpha1.ProjectScope:
		accessRuleObj.Labels["spec.projectType.scopeName"] = accessRuleSpec.ScopeName
	case v1alpha1.DepartmentScope:
		accessRuleObj.Labels["spec.departmentType.scopeName"] = accessRuleSpec.ScopeName
	case v1alpha1.ClusterScope:
		accessRuleObj.Labels["spec.clusterType"] = "cluster-name"
	}

	utils.ExpectCreateResource(k8sClient, accessRuleObj)
	return accessRuleObj
}

func deleteAccessRule(accessRule string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting accessRule <%s>", accessRule))

	accessRuleObj := getAccessRuleFromClient(accessRule, k8sClient)
	if accessRuleObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, accessRuleObj)
	}
}

func getAccessRuleFromClient(accessRuleName string, k8sClient client.Client) *v1alpha1.AccessRule {
	accessRuleKey := types.NamespacedName{Name: accessRuleName}
	accessRuleObj := &v1alpha1.AccessRule{}
	err := k8sClient.Get(context.Background(), accessRuleKey, accessRuleObj)
	if err != nil {
		return nil
	}
	return accessRuleObj
}

func keys(m map[string]bool) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}
