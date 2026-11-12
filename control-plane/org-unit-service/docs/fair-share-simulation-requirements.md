# Fair share simulation: requirements

Draft, 2026-11-12; open questions 1–5 answered by Barak the same day; design review applied (focus department, one demand rule, parent visibility, effective demand). Gathered from the Figma (V3 - Node Pools, Approved Version), the GitHub issues (#2949, #4107, #4112, #4114, #4250, #4350), Slack (Chen's 2026-10-29 meeting summary, Shiri's 2026-11-04 permission answer, Shiri's 2026-11-12 decisions and questions) and KAI's scheduler code.

Legend: **D** decided · **O** open, needs product · **T** technical fact we verified.

## 1. Goal

Before saving a node pool, show the admin how the scheduler would split the pool's GPUs between its departments under the scheduling mode and the values entered in the form, including unsaved changes. (#2949 success criterion.)

## 2. Where the simulation appears

| Form | Level shown | Status |
|---|---|---|
| Node pool, step 2 "Scheduling mode" (create and edit) | Departments of the node pool | **D**, Figma approved |
| Department form | Projects of the department | **D**: will show it; no Figma yet (epic #2950). The API must serve it. |
| Project form | The project among its siblings | **D**: same |

The node pool already exists in step 2 (created at step 1), so it always has an id.

## 3. What the user sees (node pool form)

From the Figma frames "Simulate - proportional", "Ranked", "Guaranteed quota", the scheduling-mode tooltip and the design notes.

- Header: "Total node pool capacity: N GPUs" and the three mode cards (Proportional, Ranked, Guaranteed quota).
- A grid with one row per department, grouped by rank in Ranked and Guaranteed quota. Only populated ranks show; lazy-loads rows on scroll (#4107).
- Columns depend on the mode: Weight / Rank / Weight within rank / Guaranteed GPUs / Max GPU allocation, in the order the scheduler considers them (#1..#4).
- **Weight within rank** is shown as a **percentage** of the rank's total weight (Proportional: of all departments). The popover takes the raw weight (whole number, default 2).
- **Result** column with a "Simulate workload demand" toggle. When on, each row shows:
  - a bar: blue = expected GPU allocation (guaranteed, then over quota), yellow = unmet demand;
  - a demand handle, starting at **50% of the node pool's capacity** (**D**; the Figma's "maximum in the node pool" note is superseded).
  - tooltip: Guaranteed GPU allocation, Weight within rank %, Max GPU allocation, Expected over-quota GPU allocation, Unmet GPU demand, Simulated real-time GPU demand.
- Every edit (quota, rank, weight, limit, demand, mode, advanced settings) updates all bars live.
- Warnings on the same screen: "The total assigned quota exceeds the node pool's total by N GPUs. Departments may not receive their full guaranteed quota." (banner) and per-cell popover warnings.

## 4. Permissions and visibility (D)

Shiri 2026-11-04, Figma frames 18523:1090082 / 18523:1090764, #4114:

- No **read** on departments: the grid isn't shown ("You're missing permissions to view departments"); only the node pool totals and the mode cards (disabled, "View only").
- Read but no **edit** on departments: the step is read-only; the user sees the mode and the split, including the simulation, but can't change anything.
- Switches that reset department values (CPU quota, Max GPU allocation) are disabled for users who can't edit all departments and projects in the cluster (Shiri 2026-11-12).
- A user may see only some of the departments or projects (RBAC). **Hidden org units still compete** and must count in the calculation; their individual values must not be returned. Aggregate leakage is accepted: `weightShareInRank` reveals a rank's total weight, and a unit's own fair share already depends on hidden units.
- The parent department's own result is returned only when the caller can read that department.

## 5. The calculation (T)

Must match the scheduler (KAI proportion plugin), per level:

1. Quota phase for every rank: each org unit gets min(deserved, min(request, limit)). Quota is never cut, even when Σ quota > capacity (fair shares then exceed capacity; the pool is overbooked).
2. Over-quota phase: the remainder by rank, highest first; within a rank by over-quota weight, water-filling capped by each unit's request and limit; leftover fractions handed out one GPU at a time (largest fraction, then oldest, then id).
3. Two levels: the pool between departments, then each department's fair share between its projects. A department's request is the **raw sum** of its projects' requests (not capped by their limits).
4. Capabilities change the inputs: quota off → deserved 0; ranks off → one rank; automatic weights → weight = deserved (0.01 when 0); limits off → no limits.
5. Fair share is the scheduler's target, not an allocation: whole-GPU granularity, gang scheduling and reclaim timing can make reality differ.

The code is KAI's `pkg/fairshare` (merged, kai-scheduler/KAI-Scheduler#2572; not yet in a release tag). The simulation must use it, not a port (#4250).

## 6. Demand assumptions (the modelling decisions)

| # | Assumption | Status |
|---|---|---|
| 6.1 | Every org unit the user didn't drag asks for a default demand. Dragged values are exact. | **D** |
| 6.2 | Default demand for an untouched org unit: **50% of the node pool's capacity**, capped by its limit. Applies to departments and projects alike, including hidden ones. The client sends it as a GPU count (it owns the rounding). | **D** |
| 6.3 | A department's demand is the sum of its projects' demands **only when its projects are the simulated level** (department and project forms). In the node pool form, a department's demand is its slider, else the default. So the node pool form needs no project data. | **D** (review) |
| 6.4 | In the department and project forms, the other departments are **idle: demand 0**, with a footnote. The focus department is named explicitly in the request; it still competes under the pool's rules, so it gets at most its limit, else the capacity. | **D** |
| 6.5 | The assumption must not depend on who's looking or on which rows were requested: two requests with the same scenario give the same numbers. Assumptions and the focus department are explicit request parameters, never inferred from permissions or from the rows. | **D** (review) |
| 6.7 | A department with no projects asks for 0 in the department forms, so it gets 0 even with quota (scheduler behaviour). The row stays, with `demand: 0`. | **T** |
| 6.6 | Possible later addition: a second number per unit, "up to, when the others use only their quota". | Deferred |

## 7. Scale and efficiency

- Departments per pool: tens to a few hundred. Projects per department: up to ~500 typical, 1K+ exists (barak-poc: 1,211). Projects per pool: thousands.
- The calculation needs every competing org unit, so it runs on the backend (Shiri 2026-11-12; #2949 POC conclusion). It cannot run in the FE.
- Per-rank aggregates of hidden units are **not** sufficient (water-filling depends on each weight/limit pair), so the backend holds the full set and returns only visible rows (T).
- Live updates: one call per edit, debounced; every call carries the whole scenario (stateless). Response size must not depend on the pool's size beyond the visible rows of the requested level. No paging (a scenario changes every row).
- Node pool form: no project rows and no project counts (6.3). Department and project forms: one department's projects (rank, weight, limit), bounded by the department.
- The client sends the rows it wants back (the mounted rows, up to 500), with its unsaved edits; the backend returns exactly those.

## 8. Validation and errors shown before save (D: "errors on change, not only on save")

- Local and mode-dependent errors: in the UI.
- Permission: permitted-scopes up front; disable the switches.
- Data-dependent warnings ("N departments/projects have limits that will be removed"): `/count` with a `gpuLimit` filter (#5110).
- Quota over capacity: the banner is computed in the UI from the grid. Hidden departments are out of scope for it.

## 9. Related decisions that bound the scope (D, Shiri 2026-11-12)

- "Limit projects from exceeding department quota" stays a global setting (default Off); not per node pool.
- In edit, moving down from Guaranteed quota is blocked. Ranked → Proportional is a separate decision, unrelated to the simulation.
- Max GPU allocation off: apply and reset with a warning (implemented in #4369: cluster-service resets `LIMITS_TO_DEFAULT` after a permission check).
- Mode switch before save keeps the entered values (UI state).

## 10. Non-goals

- Live workload usage as demand (no input for it; it would leak hidden units' activity).
- Time-based fair share (kValue > 0): not modelled; note it in the UI if the pool has it on.
- Historical or trend simulation.

## 11. Contract rules (from the design review)

- Required in every request: `nodePool.capacity`, `nodePool.schedulingCapabilities`, `assumptions`, `orgUnitType`, at least one row. No server-side fallback to saved values.
- Row edits merge over stored values field by field (`rank`, `gpu.deserved`, `gpu.limit`, `gpu.overQuotaWeight`, `demand`), each optional.
- 400 for: `limits` on with `quota` off; a `limit` on a row while `limits` is off; a `rank` while `ranks` is off; a duplicate id; an id not in the pool or not readable (same message); a row outside the focus department; `idle` without a focus; `capacity < 0`; `demand < 0`. A demand above the limit is capped, not rejected. `capacity = 0` is valid (a pool with no nodes yet): the quota phase still runs, so each org unit gets its quota (capped by demand and limit) and nothing above it.
- Response per row: `fairShare`, `guaranteed`, `overQuota`, `unmet`, `demand` (effective), `weightShareInRank`. Values may be fractional; the client rounds.
- Cost per call: the pool's departments (hundreds) plus, for the department forms, one department's projects (up to ~1.2k), two permitted-scopes calls, and KAI's division.

## 12. Decisions taken on the open questions (2026-11-12)

1. Default demand: 50% of the node pool's capacity, for departments and projects.
2. Other departments in the department and project forms: idle (demand 0).
3. Department and project forms show the simulation; no Figma yet, the API handles both levels.
4. Ranked → Proportional in edit: not a simulation question.
5. Overbooked banner with hidden departments: out of scope.

Still open:
6. Confirm with UX that unmet = demand − allocation, capped by the limit (the Figma tooltip's numbers don't add up).
7. In the department and project forms, is the parent department's own share shown (header or footnote)? The API returns it either way.
