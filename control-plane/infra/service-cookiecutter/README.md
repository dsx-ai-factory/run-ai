# Creating a new service from the template 

## Intro
This script suit is one-shop-store to generate a new service in the control plane.
The output contains
* A new directory for the new service base on template
* Update ci workflows to build, test, lint and validate thew new service (optional)
* Update helm chart with the new service and generate relevant templates (optional)

## Usage
Create a new service from the template using the cookiecutter with 2 simple stpes

### Step 1. Install cookiecutter

Follow the installation process in the following link https://cookiecutter.readthedocs.io/en/stable/installation.html

TL;DR for mac
```
brew install cookiecutter
```

### Step 2. Run the tool
From the root folder to the repo run the following command
`cookiecutter infra/service-cookiecutter`

## Options
The tool prompts some questions for tuning the generated service
* `app_name`: the service name.
* `go version`: the go version used to build the service.
* `use_postgresql`: the service uses db, therefore it should contain related db folder structure and configuration
* `use_redis`: the service redis, therefore it should contain related redis configuration
* `remove_gitkeep`: the generated service directory has some empty directories (i.e. pkg/database),  they are keep
* `add_service_to_ci`: Update ci workflows script to include the new service
* `add_to_helm_chart`: Update helm chart with the new service and generate relevant templates

All of them has a default value, expect app_name.