package actions

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
)

// CheckPackSources verifies every source the bundle is built from before the
// download begins: each registry it pulls images from and each repository it
// pulls charts from must be reachable and authorized, and each override must
// resolve to a real image. Charts are checked here rather than at their own
// step because they download last, after the images: an unauthorized chart
// repository would otherwise surface only after the whole image set has been
// pulled.
type CheckPackSources struct {
	Lock           *artifactlock.ArtifactLock
	Credentials    artifacts.Credentials
	ImageOverrides []spec.SourceOverride
}

func (a CheckPackSources) Kind() string { return "preflight.pack_sources" }

func (a CheckPackSources) Args() map[string]any { return map[string]any{} }

func (a CheckPackSources) Describe() string {
	return "verify every registry and chart repository the bundle pulls from is reachable and authorized"
}

func (a CheckPackSources) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	var problems, registries, chartRepositories, cliSources []string

	if deps.RegistryProber != nil {
		for _, target := range a.targets() {
			if problem := probeRegistry(ctx, deps, target, a.Credentials.Match(target.repository)); problem != "" {
				problems = append(problems, problem)
				continue
			}
			registries = append(registries, target.host)
		}
	}

	if deps.ChartFetcher != nil {
		for _, chart := range a.Lock.Spec.Charts {
			host := chartHost(chart.Source)
			if err := deps.ChartFetcher.Probe(ctx, chart); err != nil {
				problems = append(problems, fmt.Sprintf("chart %s: %v", chartLabel(chart.Name, host), err))
				continue
			}
			// An OCI chart has no host of its own to report: its registry is
			// probed with the image registries.
			if host != "" && !slices.Contains(chartRepositories, host) {
				chartRepositories = append(chartRepositories, host)
			}
		}
	}

	if deps.ImagePuller != nil {
		for _, o := range a.ImageOverrides {
			if _, _, err := deps.ImagePuller.ResolveDigests(ctx, o.Source, a.Lock.Spec.Platform); err != nil {
				problems = append(problems, fmt.Sprintf("image override %q: %v", o.Name, err))
			}
		}
	}
	if deps.CLIFetcher != nil {
		for _, cli := range a.Lock.Spec.CLIs {
			if err := deps.CLIFetcher.Probe(ctx, cli); err != nil {
				problems = append(problems, fmt.Sprintf("CLI %s: %v", cli.Name, err))
				continue
			}
			cliSources = append(cliSources, cli.Source)
		}
	}

	data := map[string]any{"registries": registries, "chartRepositories": chartRepositories, "cliSources": cliSources}
	if len(problems) > 0 {
		data["problems"] = problems
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("%d source(s) not accessible: %s", len(problems), strings.Join(problems, "; ")),
			Remediation: "add credentials for private registries under registryAuth",
			Data:        data,
		}, nil
	}

	// The sources are named rather than counted: which ones the bundle pulls
	// from is derived from the lock, so seeing the list is how a user confirms
	// it is the set they expected.
	var summary []string
	if len(registries) > 0 {
		summary = append(summary, "registries "+strings.Join(registries, ", "))
	}
	if len(chartRepositories) > 0 {
		summary = append(summary, "chart repositories "+strings.Join(chartRepositories, ", "))
	}
	if len(summary) == 0 {
		return orchestrator2.ActionResult{}, nil
	}
	return orchestrator2.ActionResult{
		Reason: "verified " + strings.Join(summary, "; "),
		Data:   data,
	}, nil
}

const ociScheme = "oci://"

// chartHost returns the host a chart is served from over HTTP, or "" for an OCI
// chart, whose registry is probed with the image registries.
func chartHost(source string) string {
	for _, scheme := range []string{"https://", "http://"} {
		rest, found := strings.CutPrefix(source, scheme)
		if !found {
			continue
		}
		if i := strings.Index(rest, "/"); i >= 0 {
			return rest[:i]
		}
		return rest
	}
	return ""
}

func chartLabel(name, host string) string {
	if host == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, host)
}

// packTarget is one registry the pack pulls from, with a repository on it to
// probe pull access against.
type packTarget struct {
	host       string
	repository string
}

// targets returns one target per distinct registry the effective image set is
// pulled from, so each registry is probed once.
func (a CheckPackSources) targets() []packTarget {
	overridden := make(map[string]bool, len(a.ImageOverrides))
	for _, o := range a.ImageOverrides {
		for _, entry := range a.Lock.ImageEntries() {
			if entry.Image.Name == o.Name && (o.Group == "" || entry.Group == o.Group) {
				overridden[entry.Group+"/"+entry.Image.Name] = true
			}
		}
	}

	byHost := map[string]string{}
	add := func(ref string) {
		host, repository, err := artifacts.ParseImageTarget(ref)
		if err != nil {
			return
		}
		if _, ok := byHost[host]; !ok {
			byHost[host] = repository
		}
	}
	for _, entry := range a.Lock.ImageEntries() {
		if !overridden[entry.Group+"/"+entry.Image.Name] {
			add(entry.Image.Image)
		}
	}
	for _, o := range a.ImageOverrides {
		add(o.Source)
	}
	// An OCI chart is served by a registry, so it is probed as one.
	for _, chart := range a.Lock.Spec.Charts {
		if strings.HasPrefix(chart.Source, ociScheme) {
			add(strings.TrimPrefix(chart.Source, ociScheme))
		}
	}

	hosts := make([]string, 0, len(byHost))
	for host := range byHost {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	targets := make([]packTarget, 0, len(hosts))
	for _, host := range hosts {
		targets = append(targets, packTarget{host: host, repository: byHost[host]})
	}
	return targets
}

func probeRegistry(ctx context.Context, deps orchestrator2.Deps, target packTarget, cred *artifacts.Credential) string {
	req := orchestrator2.RegistryCheckRequest{Host: target.host, Repo: repositoryPath(target.repository)}
	if cred != nil {
		req.Username = cred.Username
		req.Password = cred.Password
		req.CAPath = cred.CAChainPath
		req.Insecure = cred.Insecure
		req.InsecureSkipVerify = cred.InsecureSkipVerify
	}

	res, err := deps.RegistryProber.Check(ctx, req)
	switch {
	case err != nil:
		return fmt.Sprintf("%s: %v", target.host, err)
	case !res.Reachable:
		return fmt.Sprintf("%s: not reachable%s", target.host, detailSuffix(res.Detail))
	case !res.Authenticated:
		return fmt.Sprintf("%s: credentials rejected%s", target.host, detailSuffix(res.Detail))
	case !res.CanPull:
		return fmt.Sprintf("%s: no pull access%s", target.host, detailSuffix(res.Detail))
	}
	return ""
}

// repositoryPath drops the host from a host-qualified repository.
func repositoryPath(repository string) string {
	if slash := strings.Index(repository, "/"); slash >= 0 {
		return repository[slash+1:]
	}
	return repository
}
