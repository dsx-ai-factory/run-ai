# Agent Skills

Agent Skills are pre-built, evaluated workflows on top of the NVIDIA Run:ai MCP server that teach an AI agent how a Run:ai expert performs an operational task. They follow the [agentskills.io](https://agentskills.io/specification) specification.

The MCP server gives the agent access to Run:ai tools and data, while Agent Skills provide the instructions and workflow for using those capabilities to accomplish a specific task. You interact with a skill using natural-language prompts, and the agent follows the skill's workflow to determine which MCP tools to use, interpret the results, and return an actionable response.

## Prerequisites

* Deploy and connect the NVIDIA Run:ai MCP server. See [Deploy the MCP Server](mcp-server/deploy.md) and [Connect Your AI Tool](mcp-server/connect.md).
* An agent client that supports the agentskills.io specification, such as Claude Code, Cursor, VS Code with GitHub Copilot, or Codex.

## Use a Skill

Install or make the Run:ai skills directory available to your AI client according to the client's skill configuration.

After a skill is available, ask the agent to perform the task in natural language. For example, "Why did my workload fail?"

The agent identifies the appropriate skill, follows its workflow, and uses the Run:ai MCP server to retrieve the information required to complete the task.

You can also name a skill explicitly when you want the agent to use a specific workflow, for example, "Use `runai-diagnose-failed-workload` to diagnose my failed workload."

## How Skills Work

Each skill defines a repeatable workflow for a specific Run:ai task. The skill tells the agent:

* What information to collect.
* Which Run:ai MCP tools to use.
* How to interpret the returned data.
* How to present findings and recommended actions.

Skills do not replace the MCP server. They use the capabilities exposed by the MCP server to perform higher-level Run:ai workflows.

## Available Skills

| Skill | Type | Description | Example Prompt |
|---|---|---|---|
| [`runai-diagnose-failed-workload`](https://github.com/run-ai/runai/tree/master/skills/runai-diagnose-failed-workload) | Read | Diagnose a failed, crashed, or CrashLoopBackOff workload — classifies the root cause, cites evidence, and recommends a fix. | "Why did my workload fail?" |
| [`runai-audit-assigned-quota`](https://github.com/run-ai/runai/tree/master/skills/runai-audit-assigned-quota) | Read | Audit GPU quota fairness across a cluster — identifies starved and over-borrowing projects per node pool and quantifies reclaimable GPUs. | "Who is using over-quota in node-pool-a?" |
