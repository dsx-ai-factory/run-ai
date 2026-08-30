package cluster_seeder

import (
	"context"
	"fmt"
	"log"
	"time"

	nodepools "github.com/run-ai/runai/scale-tests/unified_tests/internal/node-pools"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions/config_scale"
	prometheus_monitor "github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions/prometheus-monitor"
	"github.com/samber/lo"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/output"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

func generateRunIdentifier() string {
	return fmt.Sprintf("scaleTestRun_%s", time.Now().Format("20060102-150405"))
}

func LoadCluster(runner SamplerRunner, clusterConfig *scale.ClusterConfig) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user-agent", "runai-cli/0.10.0 sdk")

	fmt.Println("Starting LoadCluster, NodepoolCount:", clusterConfig.NodepoolCount, "NodesPerPool:", clusterConfig.NodesPerPool)
	runIdentifier := generateRunIdentifier()
	fmt.Println("Generated run identifier:", runIdentifier)

	kubeConfig := sampler_actions.InitKubeConfig(clusterConfig)
	fmt.Println("Initialized kube config")

	fmt.Println("Initializing clients...")
	clients, err := sampler_actions.InitClients(ctx, clusterConfig, kubeConfig)
	if err != nil {
		log.Fatalf("Failed to init clients: %v", err)
	}
	fmt.Println("Successfully initialized clients")

	fmt.Println("Setting cluster ID...")
	err = sampler_actions.SetClusterId(ctx, clients, clusterConfig)
	if err != nil {
		log.Fatalf("Failed to set cluster id: %v", err)
	}
	fmt.Printf("Successfully set cluster ID: %s\n", clusterConfig.ClusterID)

	resourceScaleConfig := config_scale.ConfigureResourceForScaleRun(clients.KubeClient, clusterConfig)
	if err = config_scale.Configure(ctx, clients.KubeClient, clusterConfig, resourceScaleConfig); err != nil {
		log.Fatalf("Failed to adjust scheduling service memory: %v", err)
	}
	fmt.Println("Successfully adjusted test resources")

	// Wait for cluster services to stabilize after resource configuration changes
	// This prevents nodepool readiness timeouts when services are restarting
	// fmt.Println("Waiting for cluster services to stabilize after resource changes...")
	// time.Sleep(2 * time.Minute)
	// fmt.Println("Cluster services should be stable now")

	fmt.Println("Gathering environment information...")
	envInfo, err := sampler_actions.GatherEnvironmentInfo(ctx, clients.KubeClient, kubeConfig, clusterConfig)
	if err != nil {
		log.Printf("Warning: Failed to gather environment info: %v", err)
		envInfo = make(map[string]string)
	}
	fmt.Println("Successfully gathered environment information")

	fmt.Println("Setting up Prometheus monitor...")
	branchVersions := calculateBranchVersions(clusterConfig)
	metricLabels := lo.Assign(runner.GetMetricLabels(), branchVersions, map[string]string{"id": runIdentifier})

	prometheusMonitor := prometheus_monitor.NewPrometheusMonitor(clients.KubeClient, kubeConfig, clusterConfig, metricLabels)
	err = prometheusMonitor.Setup(ctx)
	if err != nil {
		log.Fatalf("Failed to setup prometheus monitor: %v", err)
	}
	fmt.Println("Successfully set up Prometheus monitor")
	defer prometheusMonitor.Teardown(ctx)

	fmt.Println("Creating node pools...")
	nodePools := nodepools.NewClusterNodePools(nodepools.ToNodePoolConfig(clusterConfig), clients)
	err = nodePools.Create(ctx)
	if err != nil {
		log.Fatalf("Failed to create node pools: %v", err)
	}
	fmt.Printf("KwokEnabled: %v\n", clusterConfig.KwokEnabled)
	fmt.Println("Waiting for node pools to be ready...")
	err = nodePools.WaitForNodepoolsReadiness()
	if err != nil {
		log.Fatalf("Failed to create node pools: %v", err)
	} else {
		fmt.Println("Node pools are ready")
	}

	fmt.Println("Initializing sampled entities...")
	entities, err := runner.InitSampledEntities(clients)
	if err != nil {
		log.Fatalf("Failed to init test entities: %v", err)
	}
	entitiesMonitorStartedAt := time.Now()
	fmt.Printf("Creating and monitoring %d test entities\n", len(entities))
	for _, entity := range entities {
		fmt.Printf("Creating entity: %s\n", entity.GetName())
		err := entity.Create(ctx)
		if err != nil {
			log.Fatalf("Failed to create entity %s: %v", entity.GetName(), err)
		}
		if entity.NumberOfSuccessfulCreatedObjects() == 0 {
			log.Fatalf("Failed to create any objects for entity -  %s", entity.GetName())
		}
		fmt.Printf("Starting monitoring for entity: %s, created: %d objects\n", entity.GetName(), entity.NumberOfSuccessfulCreatedObjects())
		err = entity.Monitor()
		if err != nil {
			log.Printf("Monitor entity: %s completed with an error: %v", entity.GetName(), err)
		}
	}

	fmt.Println("Sampling Prometheus metrics...")
	metricSamples, err := prometheusMonitor.Sample(ctx, entitiesMonitorStartedAt)
	if err != nil {
		log.Printf("Failed to sample metrics: %v", err)
	}
	fmt.Println("Successfully sampled metrics")

	fmt.Println("Generating sample reports")
	for _, entity := range entities {
		reportGenerator, ok := entity.(scale.SampledEntity)
		if !ok {
			fmt.Printf("Skipping report generation for non-sampled entity: %s\n", entity.GetName())
			continue
		}
		fmt.Printf("Generating report for entity: %s\n", entity.GetName())
		report, err := reportGenerator.GenerateSamplesReport()
		if err != nil {
			log.Fatalf("Failed to generate raw data report: %v", err)
		}
		report.RunIdentifier = runIdentifier
		report.OperatorVersion = clusterConfig.OperatorVersion
		report.CpVersion = clusterConfig.CpVersion
		report.UpdateMetaData(*clusterConfig)
		report.UpdateMetadataMap(branchVersions)
		report.UpdateMetadataMap(envInfo)
		report.Metrics = metricSamples
		fmt.Printf("Writing report for entity: %s\n", entity.GetName())
		err = output.WriteReport(report, reportGenerator.GetSamplesReportFileName())
		if err != nil {
			log.Fatalf("Failed to write report: %v", err)
		}
	}

	if clusterConfig.ClusterCleanup {
		fmt.Println("Starting cluster cleanup...")
		clusterCleanup(entities, false)
		fmt.Println("Completed cluster cleanup")
		err := nodePools.Delete(ctx)
		if err != nil {
			fmt.Printf("Failed to delete node pools: %v\n", err)
		}
		fmt.Println("Completed node pool cleanup")
	} else {
		fmt.Println("Skipping cluster cleanup as per configuration")
	}
	fmt.Println("Completed run")
}

func DeleteCluster(runner SamplerRunner, clusterConfig *scale.ClusterConfig) {
	var err error
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user-agent", "runai-cli/0.10.0 sdk")

	clients, err := sampler_actions.InitClients(ctx, clusterConfig, sampler_actions.InitKubeConfig(clusterConfig))
	if err != nil {
		log.Fatalf("Failed to init clients: %v", err)
	}

	err = sampler_actions.SetClusterId(ctx, clients, clusterConfig)
	if err != nil {
		log.Fatalf("Failed to set cluster id: %v", err)
	}

	testEntities, err := runner.InitSampledEntities(clients)
	if err != nil {
		log.Fatalf("Failed to init test entities: %v", err)
	}
	clusterCleanup(testEntities, true)
}

func clusterCleanup(testEntities []scale.Entity, forceDelete bool) {
	for i := len(testEntities) - 1; i >= 0; i-- {
		entity := testEntities[i]
		entity.Delete(forceDelete)
	}
}

func calculateBranchVersions(clusterConfig *scale.ClusterConfig) map[string]string {
	opBranchVersion, err := extractBranchVersion(clusterConfig.OperatorVersion)
	if err != nil {
		log.Printf("Failed to extract branch version from operator version: %v", err)
		opBranchVersion = "unknown"
	}
	cpBranchVersion, err := extractBranchVersion(clusterConfig.CpVersion)
	if err != nil {
		log.Printf("Failed to extract branch version from cp version: %v", err)
		cpBranchVersion = "unknown"
	}

	return map[string]string{
		"operator_branch_version": opBranchVersion,
		"cp_branch_version":       cpBranchVersion,
	}
}
