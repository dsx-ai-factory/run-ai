# Workload Types Policy Mapping

NVIDIA Run:ai provides policy mappings for many supported workload types. For those types, you can create workload policies directly. See [Supported workload types policies](supported-workload-type-policies.md).

This document describes how to register mapping resources for a workload type that is not yet mapped, or for a new workload type you are adding to the platform using the [Workload Types API](api:main-api:POST/api/v1/workload-types).

<Note>

**Note**

* Workload type policy mapping is configured via API only.
* All mapping APIs are currently experimental.
* To register mapping resources, follow the instructions in this document.
* To check which workload types are mapped, see [Manage mappings](#managing-mappings).

</Note>

## How It Works

<Note>

**Note**

For workload types that use a standard Kubernetes pod template, NVIDIA Run:ai provides system-provided formatters and field groups (`podSpec`, `podSpecContainer`, `podSpecAndContainer`, `metadata`) that cover standard Kubernetes fields out of the box. In this case you only need to create the spec mapping and can skip the formatter and field group steps.

</Note>

Three building blocks work together to enable policies for a new workload type. The dependency order is: formatters are referenced by field groups, which are referenced by spec mappings.

| Resource         | Purpose                                                                                                                                              | Endpoint                        |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| **Formatter**    | Defines how to read and write a specific field value in the CRD spec using JQ expressions                                                            | `/api/v1/policies/formatters`   |
| **Field group**  | Bundles policy-controllable fields together; each field references a formatter (or a direct spec path) and lists the item selectors that apply to it | `/api/v1/policies/field-groups` |
| **Spec mapping** | Connects your workload type (GVK) to spec selectors (JQ queries that locate subtrees of the CRD spec), each linked to one or more field groups       | `/api/v1/policies/mappings`     |

## Before You Start

* Your new workload type must be registered using the [Workload Types API](api:main-api:POST/api/v1/workload-types). See [Extending workload support with Karta](../../workloads-in-nvidia-run-ai/workload-types/extending-workload-support/README.md).
* Identify the **group**, **version**, and **kind** (GVK) of your workload type as defined in its CRD.
* You have the `policies:create` permission at the **tenant** scope. Department or project administrators with policy creation permissions scoped to a department, cluster, or project cannot create mappings.

## Creating Formatters

A formatter defines how to read and write a single policy-controllable field using JQ expressions. **In most cases, you do not need to create a custom formatter.** NVIDIA Run:ai provides two built-in formatters that cover the most common field structures:

* **`KeyValue`** - for simple key-value fields, such as `image` or `privileged`.
* **`KeyValueMap`** - for key-value map fields, such as `labels` or `options`.

To reference these in a field group, see [Creating Field Groups](workload-type-policy-mapping.md#creating-field-groups).

Create a custom formatter only when the built-in formatters do not cover your field's structure, for example when reading or writing a field requires custom JQ logic.

Send a `POST` request to `/api/v1/policies/formatters`.

**Request body:**

```json
{
  "name": "<formatter-name>",
  "description": "<optional>",
  "fields": {
    "<field-name>": {
      "getExpr": "<jq-expression>",
      "setExpr": "<jq-expression>",
      "argsInfo": {
        "<arg-name>": "<type>"
      },
      "argsEnum": {
        "<arg-name>": ["<value>"]
      },
      "arrayKey": {
        "argName": "<arg-name>"
      }
    }
  }
}
```

| Field                          | Required | Description                                                                                                                                                                                                                        |
| ------------------------------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`                         | Yes      | Unique formatter name (alphanumeric with optional internal hyphens)                                                                                                                                                                |
| `fields.<field-name>.getExpr`  | No       | JQ expression that reads the field value from the CRD spec. Required if you want the field to support `set` or `default` enforcements.                                                                                             |
| `fields.<field-name>.setExpr`  | No       | JQ expression that writes the field value into the CRD spec. Required if you want the field to support `set` or `default` enforcements.                                                                                            |
| `fields.<field-name>.argsInfo` | No       | Argument names and their types (`string`, `integer`, `boolean`, `number`, `quantity`, `ipaddress`). Argument names must match the names referenced in `getExpr` and `setExpr` using standard JQ notation: `$args.named.<argName>`. |
| `fields.<field-name>.argsEnum` | No       | Allowed values for string-typed arguments                                                                                                                                                                                          |
| `fields.<field-name>.arrayKey` | No       | The unique key argument per item, for itemized (array) fields                                                                                                                                                                      |

**Example: formatter for a ConfigMap volume field**

This field requires a custom formatter because writing it involves constructing a nested Kubernetes volume object from multiple arguments. A `KeyValue` or `specRef` approach cannot handle this structure.

```json
{
  "name": "podSpecVolumes",
  "description": "Formatter for pod spec volume fields",
  "fields": {
    "configMapVolumes": {
      "setExpr": "{ \"volumes\": [{ \"name\": $ARGS.named.name, \"configMap\": { \"name\": $ARGS.named.configMapName, \"defaultMode\": $ARGS.named.defaultMode, \"items\": $ARGS.named.items, \"optional\": $ARGS.named.optional } }] }",
      "arrayKey": {
        "argName": "name"
      },
      "argsInfo": {
        "name": "string",
        "configMapName": "string",
        "defaultMode": "integer",
        "items": "object",
        "optional": "boolean"
      }
    }
  }
}
```

A successful response returns the created formatter with its `id` and `scope: tenant`.

To retrieve available system-provided formatters for reference: `GET /api/v1/policies/formatters`

<Note>

**Note**

System-provided formatters cannot be modified. You can only update formatters that you have created.

</Note>

## Creating Field Groups

A field group bundles policy-controllable fields. Each field either references a formatter by name or provides a direct path (`specRef`) into the spec for simple scalar fields. Field groups also define item selectors (JQ queries used to filter specific items within array fields).

Send a `POST` request to `/api/v1/policies/field-groups`.

**Request body:**

```json
{
  "name": "<field-group-name>",
  "description": "<optional>",
  "fields": {
    "<field-name>": {
      "formatter": "<formatter-name>",
      "description": "<optional>",
      "type": "<field-type>",
      "itemSelectors": ["<item-selector-name>"]
    }
  },
  "itemSelectors": {
    "<item-selector-name>": {
      "description": "<optional>",
      "query": "<jq-expression>",
      "argsInfo": {
        "<arg-name>": "<type>"
      }
    }
  }
}
```

| Field                               | Required | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ----------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`                              | Yes      | Unique field group name                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `fields.<field-name>.formatter`     | No\*     | Name of the formatter that reads and writes this field. Use this or `specRef`. Must be either a built-in formatter (`KeyValue` or `KeyValueMap`) or the name of a formatter you created using `POST /api/v1/policies/formatters`.                                                                                                                                                                                                                                               |
| `fields.<field-name>.specRef`       | No\*     | Dot-separated path to the field relative to the spec element targeted by the spec selector. By default, the system assumes the field is found directly under that spec element and shares the same name. Use `specRef` when the field is nested deeper or has a different name in the spec. For example, if the field is named `privileged` in the policy but sits under `securityContext` in the spec, set `specRef` to `securityContext.privileged`. Use this or `formatter`. |
| `fields.<field-name>.type`          | No       | Field data type. One of: `string`, `boolean`, `integer`, `number`, `quantity`, `ipaddress`, `strings`, `integers`, `numbers`, `namesArray`, `object`                                                                                                                                                                                                                                                                                                                            |
| `fields.<field-name>.itemSelectors` | No       | Item selectors that apply to this field. Only needed when the field is stored in the spec as an array of items.                                                                                                                                                                                                                                                                                                                                                                 |
| `itemSelectors.<name>.query`        | No       | JQ expression that filters specific items within an array field                                                                                                                                                                                                                                                                                                                                                                                                                 |

**Example: field group for a custom job workload type**

```json
{
  "name": "acmeJobSpec",
  "description": "Policy-controllable fields for AcmeJob",
  "fields": {
    "parallelism": {
      "formatter": "KeyValue",
      "description": "Maximum number of pods running in parallel",
      "type": "integer"
    },
    "privileged": {
      "specRef": "securityContext.privileged",
      "description": "Whether the container runs in privileged mode",
      "type": "boolean"
    }
  }
}
```

A successful response returns the created field group with its `id` and `scope: tenant`.

To retrieve available system-provided field groups for reference: `GET /api/v1/policies/field-groups`

<Note>

**Note**

Unlike formatters, you can modify system-provided field groups by adding fields or item selectors to them. You cannot modify system-provided formatters; you can only create new formatters at the tenant scope.

</Note>

## Creating the Spec Mapping

A spec mapping connects your workload type (GVK) to spec selectors. Each spec selector is a JQ query that locates a subtree of the CRD spec and links it to one or more field groups, exposing those fields for policy rules at that location.

Send a `POST` request to `/api/v1/policies/mappings`.

**Request body:**

```json
{
  "workloadType": {
    "group": "<crd-group>",
    "version": "<crd-version>",
    "kind": "<crd-kind>"
  },
  "description": "<optional>",
  "specSelectors": {
    "<selector-name>": {
      "description": "<optional>",
      "query": "<jq-query>",
      "fieldGroups": ["<field-group-name>"],
      "dependency": {
        "specSelector": "<parent-spec-selector-name>",
        "path": "<dot-separated-path>"
      }
    }
  }
}
```

| Field                                          | Required | Description                                                                                                                                                                                                                                                  |
| ---------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `workloadType.group`                           | No       | The API group of the CRD, for example `acme.io`. Omit for core Kubernetes types that have no group, such as `Pod` (`v1`).                                                                                                                                    |
| `workloadType.version`                         | Yes      | The API version, for example `v1`                                                                                                                                                                                                                            |
| `workloadType.kind`                            | Yes      | The resource kind, for example `AcmeJob`                                                                                                                                                                                                                     |
| `specSelectors.<name>.query`                   | Yes      | JQ query that selects a subtree of the CRD spec                                                                                                                                                                                                              |
| `specSelectors.<name>.fieldGroups`             | Yes      | Field groups whose fields are exposed at this spec location                                                                                                                                                                                                  |
| `specSelectors.<name>.dependency.specSelector` | No       | The name of a parent spec selector that this selector depends on. Together with `path`, this forms a dependency chain the engine traces back to ensure all intermediate spec nodes exist before applying `set` or `default` rules.                           |
| `specSelectors.<name>.dependency.path`         | No       | Dot-separated path that must exist under the parent spec selector's target before this selector's `set` or `default` rules are applied. For example, if the parent selector targets `spec.template`, a `path` of `spec` ensures `spec.template.spec` exists. |

**Example: spec mapping with system and custom field groups**

```json
{
  "workloadType": {
    "group": "acme.io",
    "version": "v1",
    "kind": "AcmeJob"
  },
  "description": "AcmeJob workload type policy mapping",
  "specSelectors": {
    "acmeJobSpec": {
      "description": "the job (top) spec",
      "query": ".spec",
      "fieldGroups": ["acmeJobSpec"],
      "dependency": {
        "path": "spec"
      }
    },
    "acmePodSpec": {
      "description": "the pod spec",
      "query": ".spec.template.spec",
      "fieldGroups": ["podSpec", "podSpecAndContainer"],
      "dependency": {
        "specSelector": "acmeJobSpec",
        "path": "template.spec"
      }
    },
    "allContainers": {
      "description": "all app and init containers",
      "query": "(\n  .spec.template.spec.containers[]?,\n  .spec.template.spec.initContainers[]?\n)\n| select(.)",
      "fieldGroups": ["podSpecContainer", "podSpecAndContainer"]
    }
  }
}
```

A successful response returns the created mapping with its `id` and `scope: tenant`.

<Note>

**Note**

As a best practice, include an `allContainers` spec selector in any workload mapping. This ensures container-level fields (such as image, resource limits, and environment variables) are exposed for policy rules across all app and init containers.

</Note>

## Managing Formatters

| Operation           | Method   | Endpoint                                      |
| ------------------- | -------- | --------------------------------------------- |
| List all formatters | `GET`    | `/api/v1/policies/formatters`                 |
| Get a formatter     | `GET`    | `/api/v1/policies/formatters/{FormatterName}` |
| Update a formatter  | `PATCH`  | `/api/v1/policies/formatters/{FormatterName}` |
| Delete a formatter  | `DELETE` | `/api/v1/policies/formatters/{FormatterName}` |

## Managing Field Groups

| Operation             | Method   | Endpoint                                         |
| --------------------- | -------- | ------------------------------------------------ |
| List all field groups | `GET`    | `/api/v1/policies/field-groups`                  |
| Get a field group     | `GET`    | `/api/v1/policies/field-groups/{FieldGroupName}` |
| Update a field group  | `PATCH`  | `/api/v1/policies/field-groups/{FieldGroupName}` |
| Delete a field group  | `DELETE` | `/api/v1/policies/field-groups/{FieldGroupName}` |

## Managing Mappings

| Operation               | Method   | Endpoint                                                                            |
| ----------------------- | -------- | ----------------------------------------------------------------------------------- |
| List all mappings       | `GET`    | `/api/v1/policies/mappings`                                                         |
| Filter by workload type | `GET`    | `/api/v1/policies/mappings?filterBy=group==<group>,kind==<kind>,version==<version>` |
| Update a mapping        | `PATCH`  | `/api/v1/policies/mappings`                                                         |
| Delete a mapping        | `DELETE` | `/api/v1/policies/mappings`                                                         |

To delete a mapping, provide the GVK in the request body:

```json
{
  "workloadType": {
    "group": "acme.io",
    "version": "v1",
    "kind": "AcmeJob"
  }
}
```

## Next Steps

Once the spec mapping is registered, create a policy for your new workload type the same way as any other supported workload type. See [Supported workload types policies](supported-workload-type-policies.md#adding-a-policy).
