package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	validity_gate_monitor "github.com/run-ai/runai/scale-tests/unified_tests/internal/validity-gate-monitor"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Level: zapcore.InfoLevel,
	})))

	clusterConfig := scale.ConstructConfig()
	monitorConfig := validity_gate_monitor.ConstructConfig()

	kubeConfig := ctrl.GetConfigOrDie()
	kubeConfig.QPS = float32(clusterConfig.KubeQPS)
	kubeConfig.Burst = 200

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	if monitorConfig.RunDuration != "" {
		duration, err := time.ParseDuration(monitorConfig.RunDuration)
		if err != nil {
			log.Fatalf("invalid run_duration %q: %v", monitorConfig.RunDuration, err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}

	startTime := time.Now()
	monitor := validity_gate_monitor.NewMonitor(monitorConfig, kubeConfig)

	report, err := monitor.Run(ctx, startTime)
	if err != nil {
		log.Fatalf("validity gate monitor failed: %v", err)
	}

	if err := validity_gate_monitor.WriteReport(report, monitorConfig.OutputDir); err != nil {
		log.Fatalf("failed to write validity gate report: %v", err)
	}

	reportPath := monitorConfig.OutputDir + "/validity-gate-report.json"
	fmt.Printf("\nValidity gate report written to %s\n", reportPath)
	fmt.Printf("Overall status: %s\n", report.OverallStatus)

	if report.OverallStatus != validity_gate_monitor.StatusPass {
		os.Exit(1)
	}
}
