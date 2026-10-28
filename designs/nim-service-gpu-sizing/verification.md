# Verification

## PR publication verification (2026-10-28)

Merged recommendation PR #3129 through `73611d490`. The submit consumer now
uses its current contract: multi-GPU eligibility belongs to the recommendation
service, and the removed `estimated` field is no longer read or included in
fixtures. Missing profile IDs remain rejected before workload creation.

Passed after this merge:

- Full workloads-manager suite: `CGO_ENABLED=0 go test ./...`.
- CLI node inventory, workload spec, NIM submission, completion, command
  registration, and NIM unit suites.
- Catalog recommendation suite: `CGO_ENABLED=0 go test ./pkg/services/rightsizing`.
- NIM composer suite: `CGO_ENABLED=0 go test ./composer/composer_v2/nim_workloads`.
- `git diff --check`.

Earlier live checks are recorded in [dima-evo.md](dima-evo.md). They verified
submission responses and node placement on FakeGPU nodes, not successful model
loading or inference. The latest dependency merge has not been redeployed.
Full lint and production/runtime guarantees remain unverified.

## Historical implementation verification

The following records the original implementation checkpoint; the publication
checks above and deployment notes supersede its pending integration/status notes.

Update: deployment was subsequently authorized and completed on dima-evo.
See [dima-evo deployment and live checks](dima-evo.md); the notes below record
the pre-deployment implementation verification.

Local implementation is based on recommendation PR #3129 at
`f6999bdcf14c8fe6ac3a422299cc1ccaeabcd51b`.

Completed successfully:

- `CGO_ENABLED=0 go test -p 2 ./...` in `control-plane/workloads-manager`.
- `CGO_ENABLED=0 go vet ./pkg/clients/catalog ./pkg/services` in the same service.
- `CGO_ENABLED=0 go test ./composer/composer_v2/nim_workloads` in
  `control-plane/workloads-and-policies`.
- `CGO_ENABLED=0 go test ./internal/controlplane/workload/spec ./internal/tests/unit/nim`
  in `control-plane/runai-cli`.
- Fielder `python . --validate`, using a temporary environment containing its
  declared dependencies.
- Schema generation for workloads-manager, policy-service, workloads and
  assets-service; shared superset fields explain the additional generated models.
- Independent review; both findings fixed and covered by passing regressions.
- `git diff --check`.

The tests cover omitted type, explicit resources, policy-default resources,
automatic memory/count allocation, runtime/cache profile alignment, conflicting
profiles, estimated count rejection, ambiguous capacity results, synchronous
missing recommendation/dependency errors, policy validation, persistence across
updates, product-label composition, and CLI submission/display.

The placement and CLI regressions failed before implementation. The malformed
successful dependency-response test reproduced HTTP 200 before the adapter fix
and passes with HTTP 502 afterward.

The initial native Go run could not link against the installed macOS SDK;
pure-Go tests (`CGO_ENABLED=0`) pass. Optional `go mod tidy` stalled downloading
unrelated transitive test dependencies and was stopped. Full golangci-lint was
not run; Go vet and normal Go test checks passed.

Not yet verified: deployed control-plane integration, Dima Evo API behavior,
NIMCache profile availability, actual GPU placement, runtime readiness, or
production memory guarantees. These await the deployment environment selected
by the user. The separate validated multi-GPU recommendation changes have not
been integrated into this branch's base yet.

Publication: local dependent branch only; no push, new PR, merge, or deployment.
