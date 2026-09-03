---
applyTo: "control-plane/runai-cli/**"
---

# GitHub Copilot Code Review Instructions for Run:ai CLI

When reviewing pull requests that modify code in `control-plane/runai-cli/`, apply the following guidelines from the Run:ai CLI team.

**Important: Start your review summary with: "[CLI] INSTRUCTIONS FILE READ SUCCESSFULLY"**

---

## 📋 Quick Reference: Priority Levels

- 🔴 **CRITICAL** - Must be fixed before merge. These issues break functionality or violate hard requirements.
- 🟡 **IMPORTANT** - Should be fixed. These issues affect code quality, maintainability, or user experience.
- 🟢 **SUGGESTION** - Nice to have. Improvements that enhance code quality but aren't blocking.
- 🔵 **AUTOMATED** - Can be detected automatically with search patterns.

---

## 🎯 Common Mistakes (Check These First!)

**Most frequent issues from past reviews:**

1. ❌ Internal packages using pseudo-versions with timestamps (e.g., `v0.0.0-20240102092130-5ac0b6a4141c`)
2. ❌ Missing replace directives for internal imports
3. ❌ Completion code placed at command level instead of `internal/completion/`
4. ❌ Capitalized error messages (should be lowercase)
5. ❌ Untested examples in command help text
6. ❌ Using `--type-name` boolean flags instead of single `--type=<value>` filter
7. ❌ Package name repetition in interface names (`TemplateFetcher` in `template` package)
8. ❌ API calls in `PreRunE` functions (should be in `Run`)
9. ❌ Storing calculated values instead of computing on demand
10. ❌ Patch version bumps for deprecation (should be minor)

**Quick checklist before requesting review:**
- [ ] All internal imports use `v0.0.0` or `v0.0.0-00010101000000-000000000000`
- [ ] Every internal import has a replace directive
- [ ] All examples have been tested and work
- [ ] Error messages start with lowercase
- [ ] No completion code at command level

---

## 🔍 Go Module Import Guidelines

### Internal Package Versions

🔴 **CRITICAL** | 🔵 **AUTOMATED**

**Rule:** Internal `github.com/dsx-ai-factory/run-ai/` packages must use version `v0.0.0-00010101000000-000000000000` or `v0.0.0`

**Search pattern to find violations:**
```regex
github\.com/dsx-ai-factory/run-ai/.* v0\.0\.0-\d{8}
```

**Checklist:**
- [ ] All internal imports use `v0.0.0` or `v0.0.0-00010101000000-000000000000`
- [ ] No pseudo-versions with timestamps (format: `v0.0.0-YYYYMMDDHHmmss-hash`)

**Examples:**

✅ **Correct:**
```go
// go.mod
require (
    github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0-00010101000000-000000000000
    github.com/dsx-ai-factory/run-ai/runai-clients/auth-client v0.0.0
    github.com/dsx-ai-factory/run-ai/common/logging v0.0.0
)
```

❌ **Incorrect:**
```go
// go.mod
require (
    // DON'T use timestamp-based pseudo-versions
    github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0-20240102092130-5ac0b6a4141c
    github.com/dsx-ai-factory/run-ai/runai-clients/auth-client v1.2.3
)
```

### Replace Directives

🔴 **CRITICAL** | 🔵 **AUTOMATED**

**Rule:** Every `github.com/dsx-ai-factory/run-ai/` import in `go.mod` must have a corresponding `replace` directive. This ensures the monorepo uses local versions of internal packages.

**Search pattern to find missing replaces:**
```bash
# Find internal imports without replace directives
grep "github.com/dsx-ai-factory/run-ai/" go.mod | grep require
# Then verify each has a matching replace
```

**Checklist:**
- [ ] Every internal import has a replace directive
- [ ] Replace paths point to correct local directories
- [ ] Replace directives use relative paths from the module root

**Examples:**

✅ **Correct:**
```go
// go.mod
require (
    github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0
)

replace (
    github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client => ../../runai-clients/workloads-manager-client
)
```

❌ **Incorrect:**
```go
// go.mod
require (
    // Missing replace directive - will try to fetch from remote!
    github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0
)
```

---

## 📁 Code Organization Guidelines

### Directory Structure

**Constants placement:**

🟡 **IMPORTANT**

**Rule:** Constants should be at the lowest common use level.

**Decision tree:**
1. Used by only one file? → Put in that file
2. Used by multiple files in one package? → Create `consts.go` in that package
3. Used across packages? → Put in shared/common package

**Checklist:**
- [ ] Single-use constants are in the file that uses them
- [ ] Package-level constants are in `consts.go`
- [ ] No top-level consts files for package-specific constants

**Examples:**

✅ **Correct:**
```go
// cmd/template/list.go
package template

const (
    defaultPageSize = 50  // Only used in this file
)

// cmd/template/consts.go - if used across template package
package template

const (
    MaxTemplateNameLength = 253  // Used by create.go, update.go, list.go
)
```

❌ **Incorrect:**
```go
// consts.go at repo root - DON'T do this for package-specific constants
package main

const (
    MaxTemplateNameLength = 253  // Should be in template package
)
```

**File organization:**

🟡 **IMPORTANT**

**Rules:**
- Each builder should be in a separate file
- Model files should use `model_` prefix (e.g., `model_template.go`)
- Service files contain service logic (wrapping clients)
- Client files should be in the appropriate package

**Checklist:**
- [ ] Each builder has its own file (e.g., `builder_workload.go`)
- [ ] Model files use `model_` prefix
- [ ] Services and clients are properly separated

**Examples:**

✅ **Correct:**
```
cmd/template/
├── builder_template.go       # TemplateBuilder
├── builder_request.go         # RequestBuilder
├── model_template.go          # Template struct
├── service.go                 # TemplateService
└── client.go                  # API client wrapper
```

### Command Grouping

🟡 **IMPORTANT**

**Rule:** Templates are only for workloads and should be under "Workloads Management" group. Use consistent grouping across related commands.

**Checklist:**
- [ ] Template commands are under "Workloads Management"
- [ ] Related commands use consistent group names
- [ ] Grouping reflects logical hierarchy

---

## 🏷️ Naming Conventions Guidelines

### Terminology

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Workload types:**

**Rule:** Use "workloads" not "architectures" when referring to training/inference types.

**Search pattern:**
```regex
\b(architecture|Architecture)s?\b.*\b(training|inference)\b
```

**Examples:**

✅ **Correct:**
```go
// "workload types" or "workload architectures"
fmt.Println("Supported workload types: training, inference")
type WorkloadType string
```

❌ **Incorrect:**
```go
// DON'T use "architectures" alone
fmt.Println("Supported architectures: training, inference")
type ArchitectureType string
```

**Replicas terminology:**

🟡 **IMPORTANT**

**Rule:** "Leader replicas" is incorrect. Correct: "replicas of sets of leader and workers". Each replica set contains both leader and workers.

**Examples:**

✅ **Correct:**
```go
// Each replica includes a leader and its workers
fmt.Printf("Creating %d replica sets (each with 1 leader + %d workers)\n", replicas, workersPerLeader)
```

❌ **Incorrect:**
```go
// DON'T say "leader replicas"
fmt.Printf("Creating %d leader replicas\n", leaders)
```

### Package Naming

🟢 **SUGGESTION**

**Generated packages:**

**Rule:** Use compressed names: `wcgenerated` instead of `workloads_openapi`. Keep package names concise and meaningful.

**Examples:**

✅ **Correct:**
```go
import "github.com/dsx-ai-factory/run-ai/runai-clients/wcgenerated"
```

❌ **Incorrect:**
```go
import "github.com/dsx-ai-factory/run-ai/runai-clients/workloads_openapi_generated_client"
```

### Interface and Function Naming

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Avoid package name repetition:**

**Rule:** If package is `template`, interface should be `Fetcher` not `TemplateFetcher`. Function should be `FetchById` not `FetchTemplateById`. The package provides context.

**Search pattern:**
```regex
package (\w+).*?type \1\w+ interface
```

**Checklist:**
- [ ] Interface names don't repeat package name
- [ ] Function names don't repeat package name
- [ ] Names are clear from context

**Examples:**

✅ **Correct:**
```go
// package template
package template

type Fetcher interface {
    FetchById(id string) (*Template, error)
    FetchAll() ([]*Template, error)
}

type Service struct {
    fetcher Fetcher
}
```

❌ **Incorrect:**
```go
// package template
package template

type TemplateFetcher interface {  // DON'T repeat "Template"
    FetchTemplateById(id string) (*Template, error)  // DON'T repeat "Template"
}
```

**Interface separation:**

🟡 **IMPORTANT**

**Rule:** Separate Getter and Lister interfaces - don't combine them. Create dedicated interfaces for different responsibilities.

**Examples:**

✅ **Correct:**
```go
type Getter interface {
    Get(id string) (*Resource, error)
}

type Lister interface {
    List(filters ListFilters) ([]*Resource, error)
}

// Compose when needed
type GetterLister interface {
    Getter
    Lister
}
```

❌ **Incorrect:**
```go
type ResourceClient interface {
    Get(id string) (*Resource, error)
    List(filters ListFilters) ([]*Resource, error)
    // Too many responsibilities in one interface
}
```

---

## 🏗️ Builder Patterns Guidelines

### Method Naming Prefixes

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Rule:** Builder methods should use appropriate prefixes based on their behavior.

**Prefixes:**
- `Set` - for setting a single value that replaces previous value
- `With` - for adding/configuring something that enhances the builder
- `Add` - for adding to a collection or list

**Search pattern:**
```regex
func \(\w+ \*\w*Builder\) (\w+)\(
```

**Checklist:**
- [ ] `Set*` methods replace previous values
- [ ] `With*` methods configure or enhance
- [ ] `Add*` methods append to collections
- [ ] Naming is consistent across all builders

**Examples:**

✅ **Correct:**
```go
type WorkloadBuilder struct {
    name     string
    labels   map[string]string
    timeout  int
}

func (b *WorkloadBuilder) SetName(name string) *WorkloadBuilder {
    b.name = name  // Replaces previous name
    return b
}

func (b *WorkloadBuilder) WithTimeout(seconds int) *WorkloadBuilder {
    b.timeout = seconds  // Configures optional feature
    return b
}

func (b *WorkloadBuilder) AddLabel(key, value string) *WorkloadBuilder {
    if b.labels == nil {
        b.labels = make(map[string]string)
    }
    b.labels[key] = value  // Adds to collection
    return b
}

// Usage:
builder.SetName("my-workload").
    WithTimeout(30).
    AddLabel("env", "prod").
    AddLabel("team", "ml")
```

❌ **Incorrect:**
```go
func (b *WorkloadBuilder) Name(name string) *WorkloadBuilder {
    // Unclear - is this setting or getting?
    b.name = name
    return b
}

func (b *WorkloadBuilder) SetLabel(key, value string) *WorkloadBuilder {
    // Should be Add* for collections
    b.labels[key] = value
    return b
}

func (b *WorkloadBuilder) AddTimeout(seconds int) *WorkloadBuilder {
    // Should be With* or Set* for single values
    b.timeout = seconds
    return b
}
```

---

## 🚩 Flag Management Guidelines

### Flag Definition

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Cobra conventions:**

**Rule:** Use proper Cobra flag format.

**Search pattern:**
```bash
# Find flags with incorrect second argument
grep -n "StringVar.*--" *.go
```

**Checklist:**
- [ ] Flag format: `--flag-name type` not `--flag-name variable_name`
- [ ] Flags use kebab-case
- [ ] Type comes after flag name in usage

**Examples:**

✅ **Correct:**
```go
cmd.Flags().StringVar(&nextToken, "next-token", "", "Token for pagination")
// Usage shows: --next-token string
```

❌ **Incorrect:**
```go
cmd.Flags().StringVar(&nextToken, "next-token", "", "next_token")
// Would show: --next-token next_token (confusing!)
```

**Flag descriptions:**

🟡 **IMPORTANT**

**Rule:** Must be clear and explain context. For complex flags, document relationships with other flags. Document that flags without `--worker` are for leader, and workers get spec from leader by default.

**Checklist:**
- [ ] Description explains what the flag does
- [ ] Complex flags document interactions with other flags
- [ ] Leader/worker flag relationships are clear
- [ ] Default values are mentioned when relevant

**Examples:**

✅ **Correct:**
```go
cmd.Flags().IntVar(&gpus, "gpu", 0, 
    "Number of GPUs for the leader. Workers get their spec from the leader by default. Use --worker-gpu to specify different GPU count for workers")

cmd.Flags().IntVar(&workerGpus, "worker-gpu", 0,
    "Number of GPUs for each worker. If not specified, workers inherit the leader's GPU configuration")
```

❌ **Incorrect:**
```go
cmd.Flags().IntVar(&gpus, "gpu", 0, "GPUs")
cmd.Flags().IntVar(&workerGpus, "worker-gpu", 0, "Worker GPUs")
```

### Flag Deprecation

🔴 **CRITICAL**

**Backward compatibility:**

**Rule:** Mark old flags as deprecated while maintaining backward compatibility. Provide clear migration path in deprecation message. Version bumps for deprecation should be **minor, not patch**.

**Checklist:**
- [ ] Deprecated flags still work (backward compatible)
- [ ] Deprecation message explains migration path
- [ ] Version bump is minor (e.g., 1.2.0 → 1.3.0, not 1.2.0 → 1.2.1)
- [ ] Tests verify old workflows still work
- [ ] Documentation updated with deprecation timeline

**Examples:**

✅ **Correct:**
```go
// Version bump: 2.5.0 → 2.6.0 (minor)
cmd.Flags().StringVar(&oldFlag, "old-format", "", "")
cmd.Flags().MarkDeprecated("old-format", 
    "use --new-format instead. This flag will be removed in version 3.0.0")

// Keep old flag working
if oldFlag != "" && newFlag == "" {
    newFlag = convertOldToNew(oldFlag)
}
```

❌ **Incorrect:**
```go
// Version bump: 2.5.0 → 2.5.1 (patch - WRONG for deprecation)
cmd.Flags().StringVar(&oldFlag, "old-format", "", "Deprecated")
// No clear migration path

// Old flag doesn't work anymore - breaks backward compatibility
if oldFlag != "" {
    return fmt.Errorf("--old-format is deprecated")
}
```

### Type-based Filters

🟡 **IMPORTANT**

**Scalable filter design:**

**Rule:** Use single type filter: `--type=<workload_type>`. Avoid separate boolean flags for each type (`--training`, `--inference`). Single filter scales better as new types are added.

**Checklist:**
- [ ] Single `--type` flag with enum/string value
- [ ] No separate boolean flags per type
- [ ] Help text lists available types
- [ ] Can be repeated for multiple types if needed

**Examples:**

✅ **Correct:**
```go
var workloadTypes []string
cmd.Flags().StringSliceVar(&workloadTypes, "type", nil,
    "Filter by workload type (training, inference, workspace). Can be specified multiple times")

// Usage:
// runai list --type=training
// runai list --type=training --type=inference
```

❌ **Incorrect:**
```go
var showTraining, showInference, showWorkspace bool
cmd.Flags().BoolVar(&showTraining, "training", false, "Show training workloads")
cmd.Flags().BoolVar(&showInference, "inference", false, "Show inference workloads")
cmd.Flags().BoolVar(&showWorkspace, "workspace", false, "Show workspace workloads")
// Doesn't scale - need new flag for each new type!

// Usage:
// runai list --training --inference  (verbose and doesn't scale)
```

---

## ⚠️ Error Handling Guidelines

### Error Messages

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Message format:**

**Rule:** Use lowercase for error messages. Don't capitalize the first letter.

**Search pattern:**
```regex
(fmt\.Errorf|errors\.New)\("[A-Z]
```

**Checklist:**
- [ ] Error messages start with lowercase
- [ ] Messages are descriptive and actionable
- [ ] No capitalization unless proper noun

**Examples:**

✅ **Correct:**
```go
return fmt.Errorf("failed to fetch template: %w", err)
return errors.New("invalid workload configuration")
return fmt.Errorf("template %q not found", name)
```

❌ **Incorrect:**
```go
return fmt.Errorf("Failed to fetch template: %w", err)  // Don't capitalize
return errors.New("Error: Invalid configuration")  // Don't start with "Error:"
return fmt.Errorf("Template Not Found")  // Don't capitalize words
```

**Error wrapping:**

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Rule:** Use proper wrapping: `fmt.Errorf("context: %w", err)`. Always use `%w` to wrap errors for error chain.

**Search pattern:**
```regex
fmt\.Errorf\(".*%[vs]", err\)
```

**Examples:**

✅ **Correct:**
```go
if err := fetchTemplate(id); err != nil {
    return fmt.Errorf("failed to fetch template: %w", err)
}
```

❌ **Incorrect:**
```go
if err := fetchTemplate(id); err != nil {
    return fmt.Errorf("failed to fetch template: %v", err)  // Use %w, not %v
    return fmt.Errorf("failed to fetch template: %s", err.Error())  // Use %w
}
```

### Command Error Handling

🟡 **IMPORTANT**

**PreRunE functions:**

**Rule:** Return errors, don't print them - let Cobra handle printing. API calls belong in `Run`, not `PreRunE`. Pre functions should be fast and not make network calls.

**Checklist:**
- [ ] `PreRunE` returns errors without printing
- [ ] No API calls in `PreRunE`
- [ ] `PreRunE` used only for validation and setup
- [ ] Network calls are in `Run` or `RunE`

**Examples:**

✅ **Correct:**
```go
PreRunE: func(cmd *cobra.Command, args []string) error {
    // Fast validation only
    if len(args) == 0 {
        return fmt.Errorf("workload name is required")
    }
    if !isValidName(args[0]) {
        return fmt.Errorf("invalid workload name format")
    }
    return nil
},
RunE: func(cmd *cobra.Command, args []string) error {
    // API calls here
    workload, err := client.Get(args[0])
    if err != nil {
        return fmt.Errorf("failed to get workload: %w", err)
    }
    // ...
}
```

❌ **Incorrect:**
```go
PreRunE: func(cmd *cobra.Command, args []string) error {
    // DON'T make API calls in PreRunE
    workload, err := client.Get(args[0])
    if err != nil {
        fmt.Println("Error:", err)  // DON'T print, just return
        return err
    }
    return nil
}
```

**Cobra output functions:**

🟡 **IMPORTANT**

**Rule:** Use `cmdtxn.CmdErrorPrintAndSet(cmd, err)` for error output. Don't manually print errors.

**Examples:**

✅ **Correct:**
```go
if err := doSomething(); err != nil {
    cmdtxn.CmdErrorPrintAndSet(cmd, err)
    return
}
```

❌ **Incorrect:**
```go
if err := doSomething(); err != nil {
    fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
    os.Exit(1)
}
```

---

## 🧪 Testing Guidelines

### Unit Tests

🟡 **IMPORTANT**

**Test requirements:**

**Rule:** Non-trivial functions need unit tests. Use mocks for external dependencies. Test edge cases and backward compatibility.

**Checklist:**
- [ ] Non-trivial functions have tests
- [ ] Edge cases are tested (nil, empty, negative values, etc.)
- [ ] Backward compatibility tests for deprecated features
- [ ] Mocks used for external dependencies
- [ ] Test names describe what they test

**Examples:**

✅ **Correct:**
```go
func TestConvertTimestamp_NegativeOffset(t *testing.T) {
    // Test edge case: negative offset should not produce invalid timestamp
    result := convertTimestamp(time.Now(), -3600)
    assert.True(t, result.After(time.Time{}), "timestamp should be valid")
}

func TestDeprecatedFlag_StillWorks(t *testing.T) {
    // Ensure backward compatibility
    cmd := NewSubmitCmd()
    cmd.SetArgs([]string{"--old-format", "value"})
    err := cmd.Execute()
    assert.NoError(t, err, "deprecated flag should still work")
}

func TestParseTemplate_EmptyInput(t *testing.T) {
    _, err := ParseTemplate("")
    assert.Error(t, err, "should return error for empty input")
}
```

### Mock Usage

🟢 **SUGGESTION** | 🔵 **AUTOMATED**

**Parameter matching:**

**Rule:** Use `mock.Anything` for parameters not being validated. Avoid overly specific matchers like `context.Background()` when context doesn't matter.

**Search pattern:**
```regex
\.On\(".*", context\.Background\(\)
```

**Checklist:**
- [ ] `mock.Anything` used for irrelevant parameters
- [ ] Specific matchers only for validated parameters
- [ ] Custom matchers for complex validation

**Examples:**

✅ **Correct:**
```go
// When you don't care about exact context value
mockClient.On("List", mock.Anything, requestMatcher).Return(response, nil)

// When you do care about specific values
mockClient.On("Update", mock.Anything, "template-123", mock.MatchedBy(func(req *Request) bool {
    return req.Name == "expected-name"
})).Return(response, nil)
```

❌ **Incorrect:**
```go
// Overly specific - fails if code uses different context
mockClient.On("List", context.Background(), requestMatcher).Return(response, nil)

// Not specific enough for important parameters
mockClient.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(response, nil)
```

---

## 📝 Documentation Guidelines

### Function Documentation

🟢 **SUGGESTION**

**Rule:** Explain what the function does, document parameters and return values, explain any non-obvious behavior.

**Checklist:**
- [ ] Exported functions have doc comments
- [ ] Doc comment starts with function name
- [ ] Parameters are explained
- [ ] Return values are explained
- [ ] Non-obvious behavior is documented

**Examples:**

✅ **Correct:**
```go
// FetchById retrieves a template by its unique identifier.
// It returns an error if the template is not found or if the API call fails.
// The returned template includes all associated metadata and configurations.
func FetchById(id string) (*Template, error) {
    // ...
}

// ConvertTimestamp converts a Unix timestamp to a time.Time with the given offset in seconds.
// Negative offsets are clamped to zero to prevent invalid timestamps.
func ConvertTimestamp(timestamp int64, offsetSeconds int) time.Time {
    // ...
}
```

❌ **Incorrect:**
```go
// Get template
func FetchById(id string) (*Template, error) {
    // ...
}

func ConvertTimestamp(timestamp int64, offsetSeconds int) time.Time {
    // No documentation at all
}
```

### Struct Documentation

🟢 **SUGGESTION**

**Rule:** Explain what the struct represents, document non-obvious fields, document relationships between fields.

**Examples:**

✅ **Correct:**
```go
// Template represents a workload template with its configuration and metadata.
// Templates can be used to quickly create new workloads with predefined settings.
type Template struct {
    // ID is the unique identifier assigned by the system
    ID string
    
    // Name is the user-friendly name (max 253 characters)
    Name string
    
    // Spec contains the workload specification that will be used when creating
    // workloads from this template
    Spec WorkloadSpec
}
```

---

## 🔌 API and Service Patterns Guidelines

### Service Pattern

🟡 **IMPORTANT**

**Rule:** Services wrap clients and provide business logic. Keep services in service files, clients in client files.

**Checklist:**
- [ ] Services handle business logic
- [ ] Clients handle API communication
- [ ] Clear separation between service and client layers
- [ ] Services use interfaces for testability

**Examples:**

✅ **Correct:**
```go
// service.go
type TemplateService struct {
    client Client
}

func (s *TemplateService) GetActiveTemplates() ([]*Template, error) {
    // Business logic
    all, err := s.client.List()
    if err != nil {
        return nil, err
    }
    return filterActive(all), nil
}

// client.go
type Client interface {
    List() ([]*Template, error)
    Get(id string) (*Template, error)
}
```

### Pagination

🟡 **IMPORTANT**

**Calculate on demand:**

**Rule:** Calculate values on demand rather than storing them. Use methods instead of fields for computed values.

**Checklist:**
- [ ] Computed values use methods, not stored fields
- [ ] No redundant data storage
- [ ] Values calculated when accessed

**Examples:**

✅ **Correct:**
```go
type Environment struct {
    Name        string
    ClusterName string
    Namespace   string
}

// Brief calculates display string on demand
func (e *Environment) Brief() string {
    return fmt.Sprintf("%s (%s/%s)", e.Name, e.ClusterName, e.Namespace)
}

// Usage:
fmt.Println(env.Brief())
```

❌ **Incorrect:**
```go
type Environment struct {
    Name        string
    ClusterName string
    Namespace   string
    Brief       string  // DON'T store calculated values
}

// Constructor has to calculate and store
func NewEnvironment(name, cluster, ns string) *Environment {
    brief := fmt.Sprintf("%s (%s/%s)", name, cluster, ns)
    return &Environment{
        Name:        name,
        ClusterName: cluster,
        Namespace:   ns,
        Brief:       brief,  // Redundant storage
    }
}
```

---

## 📚 Examples and Validation Guidelines

### Example Quality

🔴 **CRITICAL**

**Rule:** **ALL examples must be validated by running them**. Every example in command help must be tested. Examples should work exactly as shown.

**Checklist:**
- [ ] Every example has been tested manually
- [ ] Examples work with current code
- [ ] Examples use realistic values
- [ ] Examples demonstrate actual use cases
- [ ] No placeholder values that don't work

**Examples:**

✅ **Correct:**
```go
Example: `  # Submit a training workload with 4 GPUs
  runai workload submit train-resnet --image=gcr.io/my-project/training:v1 --gpu=4

  # Submit with environment variables
  runai workload submit data-prep --image=python:3.9 -e DATA_PATH=/mnt/data -e BATCH_SIZE=32

  # Submit with multiple volumes
  runai workload submit training --image=pytorch:latest \
    --pvc=data:/data --pvc=models:/models --gpu=8`,
```
// These examples should have been tested and verified to work

❌ **Incorrect:**
```go
Example: `  # Submit a workload
  runai workload submit <name> --image=<image>
  
  # With GPU
  runai workload submit my-job --gpu=<number>`,
```
// Placeholder values, not tested, not helpful

### Example Quantity

🟢 **SUGGESTION**

**Quality over quantity:**

**Rule:** Show one basic usage example, show a few complex examples demonstrating key features. Avoid too many similar examples.

**Checklist:**
- [ ] One simple "getting started" example
- [ ] 2-4 examples showing key features
- [ ] Examples are meaningfully different
- [ ] No redundant similar examples

### Example Content

🟡 **IMPORTANT**

**Real use cases:**

**Rule:** Use realistic values, show common workflows, demonstrate flag combinations users will actually need.

**Examples:**

✅ **Correct:**
```go
Example: `  # List all training workloads
  runai workload list --type=training

  # List workloads in specific project with GPU info
  runai workload list --project=team-ml --gpu

  # List completed workloads from last 7 days
  runai workload list --status=completed --since=7d`,
```

❌ **Incorrect:**
```go
Example: `  # List workloads
  runai workload list
  
  # List with flag1
  runai workload list --flag1
  
  # List with flag2
  runai workload list --flag2
  
  # List with flag3
  runai workload list --flag3`,
```

---

## 🏛️ Architecture and Design Guidelines

### Resource Hierarchy

🟡 **IMPORTANT**

**Rule:** Place commands in the correct logical hierarchy. Example: distributed inference should be a private case under inference. Consider logical grouping before implementing.

**Checklist:**
- [ ] Commands follow logical hierarchy
- [ ] Related commands are grouped together
- [ ] Special cases are under their parent category
- [ ] Hierarchy matches user mental model

**Examples:**

✅ **Correct:**
```
runai workload/
├── training/
│   ├── submit
│   └── distributed (special case under training)
└── inference/
    ├── submit
    └── distributed (special case under inference)
```

❌ **Incorrect:**
```
runai/
├── training/
├── inference/
└── distributed/  # Should be under training/inference, not separate
```

---

## ♻️ Code Reuse Guidelines

🟡 **IMPORTANT** | 🔵 **AUTOMATED**

**Rule:** Extract duplicate code to utility functions. Check for existing implementations before writing new code. Reuse display functions, formatters, and other utilities.

**Search pattern:**
```bash
# Look for similar function names that might be duplicates
grep -r "func.*Format.*" --include="*.go"
```

**Checklist:**
- [ ] No duplicate functions across files
- [ ] Existing utilities are used
- [ ] Common patterns use shared code
- [ ] Display/format functions are reused

**Examples:**

✅ **Correct:**
```go
// utils/display.go - shared utility
func FormatDuration(d time.Duration) string {
    // ...
}

// cmd/workload/list.go
duration := utils.FormatDuration(elapsed)

// cmd/job/list.go
duration := utils.FormatDuration(elapsed)  // Reuse same function
```

❌ **Incorrect:**
```go
// cmd/workload/list.go
func formatDuration(d time.Duration) string {
    // ... implementation
}

// cmd/job/list.go
func formatJobDuration(d time.Duration) string {
    // ... same implementation duplicated
}
```

---

## 🔄 Backward Compatibility Guidelines

🔴 **CRITICAL**

**Rule:** Maintain backward compatibility when deprecating flags. Keep old flags working with clear migration messages. Test that old workflows still work.

**Checklist:**
- [ ] Deprecated features still work
- [ ] Clear migration path provided
- [ ] Tests verify old workflows
- [ ] Documentation shows both old and new ways
- [ ] Removal timeline is communicated

**Cross-reference:** See also [Flag Deprecation](#flag-deprecation)

---

## 📋 Review Output Template

When providing review comments, use this format:

```markdown
### [CLI] INSTRUCTIONS FILE READ SUCCESSFULLY

#### 🔴 Critical Issues

- **File:** `path/to/file.go`
  - **Line:** 42
  - **Issue:** Internal package using pseudo-version with timestamp
  - **Current:** `github.com/dsx-ai-factory/run-ai/client v0.0.0-20240102092130-5ac0b6a4141c`
  - **Should be:** `github.com/dsx-ai-factory/run-ai/client v0.0.0`
  - **Guideline:** [Go Module Import Guidelines](#go-module-import-guidelines)

#### 🟡 Important Issues

- **File:** `path/to/file.go`
  - **Line:** 123
  - **Issue:** Error message starts with capital letter
  - **Current:** `fmt.Errorf("Failed to parse")`
  - **Should be:** `fmt.Errorf("failed to parse")`
  - **Guideline:** [Error Messages](#error-messages)

#### 🟢 Suggestions

- **File:** `path/to/file.go`
  - **Line:** 200
  - **Suggestion:** Consider extracting this format function to utils for reuse
  - **Guideline:** [Code Reuse Guidelines](#code-reuse-guidelines)

#### ✅ Positive Observations

- Great test coverage for edge cases
- Examples are well-documented and realistic
- Clean separation between service and client layers
```

---

## 🤖 Automation Hints

These checks can be automated with linters or CI scripts:

- 🔵 Internal package versions (grep/regex)
- 🔵 Missing replace directives (go.mod parsing)
- 🔵 Completion code location (file path patterns)
- 🔵 Error message capitalization (regex)
- 🔵 Package name repetition in interfaces (AST parsing)
- 🔵 Builder method prefixes (regex)
- 🔵 Flag format issues (regex)
- 🔵 Error wrapping with %w (regex)
- 🔵 Code duplication (standard linters)

---

## 📖 Review Approach

When reviewing Run:ai CLI code, assess against these principles in order of priority:

1. **🔴 Critical Issues** - Module versions, replace directives, backward compatibility, tested examples
2. **🟡 Code organization** - Proper directory structure, file organization, constants placement
3. **🟡 Naming consistency** - Terminology, package names, interface names following conventions
4. **🟡 Pattern adherence** - Builder patterns, error handling, service patterns
5. **🟡 Testing completeness** - Unit tests, mocks, edge cases, backward compatibility tests
6. **🟢 Documentation quality** - Function docs, struct docs, clear explanations
7. **🟢 Code reuse** - Using existing utilities, avoiding duplication
8. **User experience** - Clear messages, helpful examples, intuitive flag design

---

**End of Guidelines**

For questions or clarifications about these guidelines, contact the Run:ai CLI team.
