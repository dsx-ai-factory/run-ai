package readiness

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type restartKey struct {
	namespace string
	pod       string
	container string
}

func (c *Checker) sampleRestartCounts(ctx context.Context) (map[restartKey]int32, error) {
	counts := map[restartKey]int32{}
	for _, ns := range c.namespaces {
		pods, err := c.listPods(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("listing pods in %s: %w", ns, err)
		}
		for _, p := range pods {
			for _, cs := range p.Status.ContainerStatuses {
				counts[restartKey{ns, p.Name, cs.Name}] = cs.RestartCount
			}
		}
	}
	return counts, nil
}

// checkRestartsStable asserts that nothing restarted during a quiet window, rather than
// that the lifetime restart count is zero.
//
// This distinction is the whole reason this package exists. remove-limits restarts pods
// by design, and enabling the Prometheus admin API rolls its StatefulSet — so a nonzero
// lifetime restart count is expected on a healthy environment and says nothing about
// whether it has finished settling. What does say something is whether anything restarted
// while we were watching.
//
// Containers appearing or disappearing during the window count as instability too: either
// means a pod was still being created or replaced.
func (c *Checker) checkRestartsStable(
	ctx context.Context,
	first map[restartKey]int32,
	firstErr error,
	sampleStart time.Time,
) CheckResult {
	res := CheckResult{Check: "restarts_stable", Status: StatusPass}
	if firstErr != nil {
		return failf(res, "first restart sample failed: %v", firstErr)
	}

	quiet := c.cfg.quietPeriod()
	remaining := quiet - time.Since(sampleStart)
	if remaining > 0 {
		fmt.Printf("waiting %v for the restart quiet period to elapse...\n", remaining.Round(time.Second))
		select {
		case <-time.After(remaining):
		case <-ctx.Done():
			return failf(res, "context cancelled during quiet period: %v", ctx.Err())
		}
	}

	second, err := c.sampleRestartCounts(ctx)
	if err != nil {
		return failf(res, "second restart sample failed: %v", err)
	}

	for k, after := range second {
		before, seen := first[k]
		if !seen {
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf(
				"%s/%s/%s appeared during the quiet period (restarts=%d)", k.namespace, k.pod, k.container, after))
			continue
		}
		if after > before {
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf(
				"%s/%s/%s restarted %d time(s) during the quiet period", k.namespace, k.pod, k.container, after-before))
		}
	}
	for k := range first {
		if _, seen := second[k]; !seen {
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf(
				"%s/%s/%s disappeared during the quiet period", k.namespace, k.pod, k.container))
		}
	}

	sort.Strings(res.Detail)
	res.Summary = fmt.Sprintf("%d containers sampled over %v, %d unstable", len(second), quiet, len(res.Detail))
	return res
}
