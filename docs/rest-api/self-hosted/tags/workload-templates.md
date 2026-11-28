This set of endpoints manages workload templates used to define reusable workload configurations across various workload types in the NVIDIA Run:ai platform.
Templates help standardize workload definitions for consistent submission and resource usage across projects and users.

Use these endpoints to:

* **List and inspect templates** - List all templates, count them, and retrieve a template of any type by its ID.
* **Check names** - Check whether a template name is available under a given scope.
* **Manage templates by type** - Create, list, and retrieve templates for workspaces, trainings, distributed trainings, and inferences, and delete them. For workspace, training, and distributed training templates, you can also update them with patch or full update requests.
