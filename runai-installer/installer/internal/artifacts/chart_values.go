package artifacts

import "helm.sh/helm/v3/pkg/chart/loader"

func ChartValues(workDir, name string) (map[string]any, error) {
	path, err := LockedChartPath(workDir, name)
	if err != nil {
		return nil, err
	}
	chart, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return chart.Values, nil
}
