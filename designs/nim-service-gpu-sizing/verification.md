# Verification

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
