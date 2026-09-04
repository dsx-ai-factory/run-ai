import shutil
import os
import fileinput
import sys
from datetime import datetime
from jinja2 import Environment, FileSystemLoader

# SERVICE_DIRECTORY is the directory of the new service
SERVICE_DIRECTORY = os.path.realpath(os.path.curdir)
# REPO_ROOT_DIRECTORY is the repo root directory, should be the parent of the SERVICE_DIRECTORY
REPO_ROOT_DIRECTORY = os.path.dirname(SERVICE_DIRECTORY)
# TMP_DIRECTORY is a path to template directory
TMP_DIRECTORY = "/tmp"

# clear_prefixes is a list of prefixes (of lines in files) that annotates if the line needs to removed
clear_prefixes = {
    "DB": {{ cookiecutter.use_postgresql }},
    "CACHE": {{ cookiecutter.use_redis }},
    "PUBSUB": {{ cookiecutter.use_nats }}
}

# ci_workflow_files is a list of the ci workflows, with path relative to the repo root
ci_workflow_files = [".github/workflows/ci-commit-master.yml",
    ".github/workflows/ci-pr-to-master.yml",
    ".github/workflows/ci-pr-to-v.yml",
    ".github/workflows/ci-tag.yml"]

# github_labeler_file is the path to the github labeler file, with path relative to the repo root
github_labeler_file = ".github/labeler.yml"

# chart_templates_templates is list of path with the templates of the chart templates, with path relative to the repo root
chart_templates_templates = ["infra/service-cookiecutter/hooks/control-plane-chart-templates/deployment.yaml",
    "infra/service-cookiecutter/hooks/control-plane-chart-templates/service.yaml",
    "infra/service-cookiecutter/hooks/control-plane-chart-templates/hpa.yaml",
    "infra/service-cookiecutter/hooks/control-plane-chart-templates/rbac.yaml",
    "infra/service-cookiecutter/hooks/control-plane-chart-templates/serviceaccount.yaml",
    "infra/service-cookiecutter/hooks/control-plane-chart-templates/configmap.yaml",
    "infra/service-cookiecutter/hooks/control-plane-chart-templates/secret.yaml",
]

# chart_values is path to the helm chart values of the control plane
chart_values = "control-plane-chart/values.yaml"
# chart_values_template is a path to the template of the helm values value, with path relative to the repo root
chart_values_template = "infra/service-cookiecutter/hooks/control-plane-chart-templates/values-service-section.yaml"

# cc_variables are 2nd mapping of the cookiecutter.json, used for accessing it and using it to render templates
# outside the service directory (since there is not simple way to reuse it
# (ref: https://github.com/cookiecutter/cookiecutter/issues/851#issuecomment-269794953)
cc_variables = {
    'cookiecutter_app_name': '{{ cookiecutter.app_name }}',
    'cookiecutter_app_name_camel_case_upper': '{{ cookiecutter.__app_name_camel_case_upper }}',
    'cookiecutter_app_name_camel_case_lower': '{{ cookiecutter.__app_name_camel_case_lower }}',
    'cookiecutter_app_name_underscore_lower': '{{ cookiecutter.__app_name_underscore_lower }}',
}


    #
    # utils section
    #################

def make_tmp_directory(base_tmp_directory):
    """
    Creates a unique tmp directory (unique as time can be unique)
    """
    current_date = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
    tmp_directory_path = f'{base_tmp_directory}/control_plane_cc_{current_date}'
    os.makedirs(tmp_directory_path)
    return tmp_directory_path

def remove_db_folders():
    """
    Removes db migration folders
    """
    shutil.rmtree(os.path.join(
        SERVICE_DIRECTORY, "pkg/database"
    ))

def clean_file_from_prefix(filename, prefixes):
    """
    Removes prefix from lines, or removing the line it self from the file

    Args:
        filename (str): The name of the file to edit.
        prefixs (dict): dict of prefix and a bool to keep/skip the line
    """

    with open(filename, 'r') as file:
        lines = file.readlines()

    with open(filename, 'w') as file:
        for line in lines:

            content = line.split("-CC-PREFIX", 1)
            # If the prefix of the line is define in the prefixes dict, keep or skip
            if content[0] in prefixes:
                # if true keep
                if prefixes[content[0]]:
                    file.write(content[1])
                # o.w. skip the line
            else:
                # o.w. write the line as is
                file.write(line)
    file.close()

def delete_gitkeep_files(folder_path, gitkeep):
  """
  This function scans a folder recursively and deletes git keep files

  Args:
      folder_path: The path to the folder to scan.
      gitkeep: name of the gitkeep file to delete
  """
  if not {{ cookiecutter.remove_gitkeep}}:
    return

  for filename in os.listdir(folder_path):
    if filename == gitkeep:
        os.unlink(os.path.join(folder_path, filename))
    else:
        full_path = os.path.join(folder_path, filename)
        if os.path.isdir(full_path):
            delete_gitkeep_files(full_path, gitkeep)


    #
    # github workflow section
    #################

def generate_service_to_ci_workflow(workflow_in, workflow_out, service_name):
    """
    This function generates a copy of the input workflow with the new service added to the env/service-list in the workflow file

    Args:
      workflow_in: The path of the workflow file
      workflow_out: The path of the workflow output file
      service_name: the service to add to the service list
    """
    # Although the ci workflow are kinda yaml (they are really jinja yaml), the function try to avoid huge changes to the workflow
    # Using yaml pkgs to update the workflow leads to changes in the structure and order of the sections
    # Therefore using a simple state-mechain to in-place add the service
    try:
        with open(workflow_in, 'r') as file:
            lines = file.readlines()

        state = "pre-env"
        with open(workflow_out, 'w') as file:
            for line in lines:
                file.write(line)
                if state == "pre-env" and line.startswith("env:"):
                   state = "in-env-pre-service-list"
                if state == "in-env-pre-service-list" and "SERVICE_LIST:" in line:
                    file.write(f'    {service_name}\n')
                    state = "service-added"

    except FileNotFoundError:
        print(f"File '{workflow_in}' not found.")
    except Exception as e:
        print(f"An error occurred: {e}")

def generate_service_to_ci_workflows(workflows, service_name, tmp_directory):
  """
  This function generates ci workflows files with the new service_name

  Args:
      workflows: list of ci workflow files
      service_name: The name of the service to add
      tmp_directory: a directory to output the files
  """
  if not {{ cookiecutter.add_service_to_ci }}:
    return
  workflow_out_directory = f'{tmp_directory}/workflows'
  os.makedirs(workflow_out_directory)
  for workflow in workflows:
    generate_service_to_ci_workflow(f'{REPO_ROOT_DIRECTORY}/{workflow}',  f'{workflow_out_directory}/{os.path.basename(workflow)}' , service_name)

def copy_ci_workflow(ci_workflow_files, tmp_directory):
    """
    This function copies the generated ci workflows files back to the repo

    Args:
      ci_workflow_files: list of ci workflow files
      tmp_directory: a directory to output the files
    """
    if not {{ cookiecutter.add_service_to_ci }}:
        return

    workflow_out_directory = f'{tmp_directory}/workflows'
    for workflow in ci_workflow_files:
        shutil.copyfile(f'{workflow_out_directory}/{os.path.basename(workflow)}', f'{REPO_ROOT_DIRECTORY}/{workflow}')

    #
    # helm chart section
    #################

def render_chart_template(template_in, template_out, service_name):
    """
    This function renders a chart template base on the input template

    Args:
      template_in: the template to render
      template_out: the template output
    """
    #  Both the chart template and the template of the chart template (don't laugh), are jinja base
    #  To distinguish between placeholders. chart template uses { } vs template of the chart template [ ]
    env = Environment(loader=FileSystemLoader('/'), variable_start_string='[[', variable_end_string=']]')
    template = env.get_template(template_in)
    with open(template_out, "w") as file:
        file.write(template.render(**cc_variables))
    # clean the file
    clean_file_from_prefix(template_out, clear_prefixes)


def inject_service_into_helm_chart_values(values_template, repo_values, service_name, charts_template_out_directory):
    """
    This function inject a values section for the new service in the values.yaml of the control plane

    Args:
      values_template: path to the template of the value
      service_name: The name of the service to add
      tmp_directory: a directory to output the files
    """
    #  Generate a values section for the new service

    #  Both the chart template and the template of the chart template (don't laugh), are jinja base
    #  To distinguish between placeholders. chart template uses { } vs template of the chart template [ ]
    env = Environment(loader=FileSystemLoader(REPO_ROOT_DIRECTORY), variable_start_string='[[', variable_end_string=']]')
    template = env.get_template(values_template)
    service_values_section = template.render(**cc_variables)

    # Copy the helm chart values into the tmp directory
    tmp_values = f'{charts_template_out_directory}/{os.path.basename(repo_values)}'
    shutil.copyfile(repo_values, tmp_values)

    # Append the values with the service section
    with open(tmp_values, "a") as file:
      file.write(service_values_section)

    # clean the file
    clean_file_from_prefix(tmp_values, clear_prefixes)



def generate_service_helm_chart(templates, service_name, tmp_directory):
  """
  This function creates helm template for the service

  Args:
      templates: are list of helm chart templates files needed for the service
      service_name: The name of the service to add
      tmp_directory: a directory to output the files
  """
  if not {{ cookiecutter.add_to_helm_chart }}:
    return

  # Generating service helm chart is done in 2 steps
  #  1. Generate chart templates for the new service (deployment.yaml, configmap.yaml service.yaml etc)
  #  2. Add a section for the new service in values.yaml

  # The helm template files are jinja, which lead to issues when generate them with the cookiecutter (which relies on jinja it self)
  # Therefore the script generates the helm template by itself, from a template of templates
  charts_template_out_directory = f'{tmp_directory}/charts-template'
  os.makedirs(charts_template_out_directory)
  for template in templates:
    render_chart_template(f'{REPO_ROOT_DIRECTORY}/{template}', f'{charts_template_out_directory}/{os.path.basename(template)}', service_name)

  inject_service_into_helm_chart_values(chart_values_template,
   f'{REPO_ROOT_DIRECTORY}/{chart_values}',
   service_name,
   charts_template_out_directory)


def add_service_to_labelers(labeler_file, service_name):
    """
    This function adds the service to the labeler file

    Args:
      labeler_file: The path of the github labeler file
      service_name: the service to add to the service list
    """
    with open(f'{REPO_ROOT_DIRECTORY}/{labeler_file}', "a") as file:
        file.write(f'{service_name}: {service_name}/**/*\n')

def create_service_chart_templates_directory(service_name):
    """
    This function creates a directory in the repository to hold the service helm chart templates

    Args:
      service_name: The name of the service to add
    """
    if not {{ cookiecutter.add_to_helm_chart }}:
        return

    repo_template_directory = f'{REPO_ROOT_DIRECTORY}/control-plane-chart/templates/{service_name}'
    os.makedirs(repo_template_directory, exist_ok=True)

def copy_helm_chart_to_repo(templates, service_name, tmp_directory):
    """
    This function copies the generate helm chart into the repository

    Args:
      templates: are list of helm chart templates files needed for the service
      service_name: The name of the service to add
      tmp_directory: a directory to output the files
    """
    if not {{ cookiecutter.add_to_helm_chart }}:
        return

    repo_template_directory = f'{REPO_ROOT_DIRECTORY}/control-plane-chart/templates/{service_name}'
    charts_template_out_directory = f'{tmp_directory}/charts-template'
    for template in templates:
        shutil.copyfile(f'{charts_template_out_directory}/{os.path.basename(template)}', f'{repo_template_directory}/{os.path.basename(template)}')

    tmp_values = f'{charts_template_out_directory}/{os.path.basename(chart_values)}'
    shutil.copyfile(tmp_values, f'{REPO_ROOT_DIRECTORY}/{chart_values}')

    # If there is not use for db or redis, there is not need for secret
    if not {{ cookiecutter.use_postgresql }} and not {{ cookiecutter.use_redis }} and not {{ cookiecutter.use_nats }}:
        os.remove(f'{repo_template_directory}/secret.yaml')


    #
    # main section
    #################

def main():
    service_name = "{{  cookiecutter.app_name }}"
    # Handle db migration
    if not {{  cookiecutter.use_postgresql }}:
        remove_db_folders()

    # Clean prefixes from files
    clean_file_from_prefix(os.path.join(SERVICE_DIRECTORY, "pkg/config/config.go"), clear_prefixes)
    clean_file_from_prefix(os.path.join(SERVICE_DIRECTORY, "cmd/main.go"), clear_prefixes)
    delete_gitkeep_files(SERVICE_DIRECTORY, ".gitkeep")

    # In order to have minimum impact on the repo, creating/editing of repo files is done on tmp directory
    # once everything is successfully, copied back into the repository
    tmp_directory = make_tmp_directory(TMP_DIRECTORY)

    generate_service_to_ci_workflows(ci_workflow_files,
        service_name,
        tmp_directory)

    generate_service_helm_chart(chart_templates_templates,
        service_name,
        tmp_directory)



    # Finish generating, now copy
    create_service_chart_templates_directory(service_name)
    copy_helm_chart_to_repo(chart_templates_templates, service_name, tmp_directory)
    copy_ci_workflow(ci_workflow_files, tmp_directory)
    add_service_to_labelers(github_labeler_file, service_name)

    # clean up temp
    shutil.rmtree(tmp_directory)

if __name__ == "__main__":
    main()
