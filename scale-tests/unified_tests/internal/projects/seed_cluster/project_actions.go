package seed_cluster

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	nodepools "github.com/run-ai/runai/scale-tests/unified_tests/internal/node-pools"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	orgunitclient "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"golang.org/x/sync/semaphore"
	"k8s.io/utils/ptr"
)

const (
	EntityName = "project"
)

type NameIdPair struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type ProjectResponse struct {
	Name              string `json:"name"`
	Id                string `json:"id"`
	SuccessAtCreation bool   `json:"success_at_creation"`
	CreateFailMessage string `json:"create_fail_message"`
}

type ProjectConfig struct {
	ProjectCount           int
	ControlPlaneQPS        int
	ClusterID              string
	ProjectCreationTimeout time.Duration
	MonitorWaitTime        time.Duration
	BatchSize              int
}

type ProjectSampler struct {
	config        *ProjectConfig
	clients       *sampler_actions.Clients
	projects      sync.Map
	sem           *semaphore.Weighted
	wg            *sync.WaitGroup
	projectPrefix string
}

func NewProjectSampler(clusterConfig *scale.ClusterConfig, projectConfig *ProjectConfig, clients *sampler_actions.Clients, projectPrefix string) scale.Entity {
	return &ProjectSampler{
		config:        projectConfig,
		clients:       clients,
		projects:      sync.Map{},
		sem:           semaphore.NewWeighted(int64(clusterConfig.ControlPlaneQPS)),
		wg:            &sync.WaitGroup{},
		projectPrefix: projectPrefix,
	}
}

func (p *ProjectSampler) Create(ctx context.Context) error {
	fmt.Println(fmt.Sprintf("Creating %d projects", p.config.ProjectCount))
	nodepoolsNameIdPairs, err := p.getNodepoolNameIdPairs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get nodepools: %w", err)
	}
	for i := 1; i <= p.config.ProjectCount; i++ {
		i := i
		p.wg.Add(1)
		go p.createProject(ctx, i, nodepoolsNameIdPairs)
	}

	p.wg.Wait()
	return nil
}

func (p *ProjectSampler) GetName() string {
	return EntityName
}

func (p *ProjectSampler) GetProjectsPrefix() string {
	return p.projectPrefix
}

func (p *ProjectSampler) Monitor() error {
	fmt.Printf("Monitoring %d Projects, project timeout:[%s]\n", p.config.ProjectCount, p.config.ProjectCreationTimeout)
	ticker := time.NewTicker(p.config.MonitorWaitTime)
	defer ticker.Stop()
	timeout := time.After(p.config.ProjectCreationTimeout)

	for {
		select {
		case <-ticker.C:
			projectsAreReady := p.allProjectsReady()
			fmt.Println(fmt.Sprintf("Projects are in ready state ? %v", projectsAreReady))
			if projectsAreReady {
				fmt.Println("All projects are ready")
				return nil
			}
		case <-timeout:
			return fmt.Errorf("project readiness timeout has reached not all projects are ready")
		}
	}
}

func (p *ProjectSampler) NumberOfSuccessfulCreatedObjects() int {
	return len(p.GetActiveProjects())
}

func (p *ProjectSampler) Delete(forceDelete bool) {
	projectPrefixFilterString := fmt.Sprintf("name=@%s", p.projectPrefix)
	clusterIdFilterString := fmt.Sprintf("clusterId==%s", p.config.ClusterID)
	filter := []string{clusterIdFilterString, projectPrefixFilterString}
	if forceDelete {
		filter = []string{}
	}
	countVal, err := p.Count(filter)
	if err != nil {
		fmt.Printf("Failed to count projects: %v\n", err)
		return
	}
	fmt.Printf("Going to Delete: %d Projects\n", countVal)
	for i := 0; i < int(countVal); i += p.config.BatchSize {
		request := p.clients.OrgUnitServiceClient.ProjectsAPI.GetProjects(context.Background())
		if !forceDelete {
			request = request.FilterBy([]string{clusterIdFilterString, projectPrefixFilterString})
		}
		request = request.Limit(int32(p.config.BatchSize))
		request = request.Offset(int32(i))
		projects, _, err := request.Execute()
		if err != nil {
			fmt.Printf("Failed to get projects: %v", err)
			return
		}
		namespaces := []string{}
		for _, project := range projects.Projects {
			resp, err := p.clients.OrgUnitServiceClient.ProjectsAPI.DeleteProject(context.Background(), project.Id).Execute()
			if err != nil || resp.StatusCode > 300 {
				fmt.Printf("failed to delete project-%s: %v\n", project.Name, err)
				continue
			}
			namespace := sampler_actions.ProjectToNs(project.Name)
			namespaces = append(namespaces, namespace)
		}
		sampler_actions.DeleteK8sNameSpaces(namespaces, p.clients.KubeClient)
	}
	p.projects.Clear()
}

func (p *ProjectSampler) Count(filter []string) (int64, error) {
	countProjectsRequest := p.clients.OrgUnitServiceClient.ProjectsAPI.CountProjects(context.Background())
	if len(filter) > 0 {
		countProjectsRequest = countProjectsRequest.FilterBy(filter)
	}
	count, _, err := countProjectsRequest.Execute()
	if err != nil {
		return 0, err
	}
	if count == nil {
		return 0, fmt.Errorf("failed to count projects: count is nil")
	}
	return count.Count, nil
}

func (p *ProjectSampler) GetSampleReportGenerator() scale.SampledEntity {
	return nil
}

func (p *ProjectSampler) getProjectName(counter int) string {
	return fmt.Sprintf("%s-%d", p.projectPrefix, counter)
}

func (p *ProjectSampler) createProject(ctx context.Context, i int, nodepoolsNameIdPairs []NameIdPair) {
	defer p.wg.Done()
	p.sem.Acquire(ctx, 1)
	defer p.sem.Release(1)
	projectName := p.getProjectName(i)
	project, resp, err := p.clients.OrgUnitServiceClient.ProjectsAPI.CreateProject(ctx).ProjectCreationRequest(orgunitclient.ProjectCreationRequest{
		Name:      projectName,
		ClusterId: p.config.ClusterID,
		Resources: func() []orgunitclient.ResourcesNullable {
			ret := []orgunitclient.ResourcesNullable{}
			for _, np := range nodepoolsNameIdPairs {
				ret = append(ret, orgunitclient.ResourcesNullable{
					NodePool: *orgunitclient.NewNullableResourcesNodePool(&orgunitclient.ResourcesNodePool{
						Id:   np.Id,
						Name: ptr.To(np.Name),
					}),
					Gpu: &orgunitclient.ResourcesNullableGpu{
						Deserved:        ptr.To(float64(0)),
						Limit:           *orgunitclient.NewNullableFloat64(ptr.To(float64(-1))),
						OverQuotaWeight: orgunitclient.NullableFloat64{},
					},
					Cpu:    orgunitclient.NullableResourcesCpu{},
					Memory: orgunitclient.NullableResourcesMemory{},
				})
			}

			return ret
		}(),
	}).Execute()

	response := ProjectResponse{
		Name: projectName,
	}

	var body string
	if resp != nil {
		body = sampler_actions.BestEffortRead(resp.Body)
	}
	if err != nil && (resp == nil || resp.StatusCode != http.StatusConflict) {
		response.SuccessAtCreation = false
		if resp == nil {
			response.CreateFailMessage = fmt.Sprintf("failed to create project-%d: %v", i, err)
		} else {
			response.CreateFailMessage = fmt.Sprintf("failed to create project-%d: %s: %v", i, body, err)
		}
		fmt.Printf("[ERROR] %s\n", response.CreateFailMessage)

	} else if resp != nil && resp.StatusCode == http.StatusConflict {
		// Project already exists — resolve its ID and treat as success.
		existing, err := p.resolveExistingProject(ctx, projectName)
		if err != nil {
			response.SuccessAtCreation = false
			response.CreateFailMessage = fmt.Sprintf("project-%d already exists but could not resolve ID: %v", i, err)
			fmt.Printf("[ERROR] %s\n", response.CreateFailMessage)
		} else {
			fmt.Printf("[INFO] project %s already exists (id=%s), reusing\n", projectName, existing)
			response.SuccessAtCreation = true
			response.Id = existing
		}
	} else if resp.StatusCode != http.StatusCreated {
		response.SuccessAtCreation = false
		response.CreateFailMessage = fmt.Sprintf("failed to create project-%d, status code: %d: %s", i, resp.StatusCode, body)
		fmt.Printf("[ERROR] %s\n", response.CreateFailMessage)
	} else {
		response.SuccessAtCreation = true
		response.Id = project.Id
	}
	p.projects.Store(projectName, &response)
}

func (p *ProjectSampler) getNodepoolNameIdPairs(ctx context.Context) ([]NameIdPair, error) {
	nps, err := nodepools.FetchNodepools(ctx, p.clients.ClusterServiceClient.NodePoolsAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodepools: %w", err)
	}
	pairs := make([]NameIdPair, len(nps))
	for i, np := range nps {
		pairs[i] = NameIdPair{Name: np.Name, Id: np.ID}
	}
	return pairs, nil
}

func (p *ProjectSampler) resolveExistingProject(ctx context.Context, name string) (string, error) {
	projects, _, err := p.clients.OrgUnitServiceClient.ProjectsAPI.GetProjects(ctx).
		FilterBy([]string{
			fmt.Sprintf("clusterId==%s", p.config.ClusterID),
			fmt.Sprintf("name==%s", name),
		}).Execute()
	if err != nil {
		return "", err
	}
	if projects == nil || len(projects.Projects) == 0 {
		return "", fmt.Errorf("project %q not found", name)
	}
	return projects.Projects[0].Id, nil
}

func (p *ProjectSampler) allProjectsReady() bool {
	projectPrefixFilterString := fmt.Sprintf("name=@%s", p.projectPrefix)
	clusterIdFilterString := fmt.Sprintf("clusterId==%s", p.config.ClusterID)
	for i := 0; i < p.config.ProjectCount; i += p.config.BatchSize {
		request := p.clients.OrgUnitServiceClient.ProjectsAPI.GetProjects(context.Background())
		request = request.FilterBy([]string{clusterIdFilterString, projectPrefixFilterString})
		request = request.Limit(int32(p.config.BatchSize))
		request = request.Offset(int32(i))
		projects, _, _ := request.Execute()
		if projects == nil {
			return false
		}
		for _, project := range projects.Projects {
			if project.Status.GetPhase() != orgunitclient.PROJECTPHASE_READY {
				return false
			}
		}
	}
	return true
}

func (p *ProjectSampler) GetActiveProjects() []*ProjectResponse {
	res := []*ProjectResponse{}
	p.projects.Range(func(k, v interface{}) bool {
		project := v.(*ProjectResponse)
		if project.SuccessAtCreation {
			res = append(res, project)
		}
		return true
	})
	return res
}
