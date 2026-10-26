# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

import os

IS_ENGINE = {{cookiecutter.engine_instead_of_cluster}}
DEPLOY_TO_REMOTE_CLUSTER = {{cookiecutter.deploy_remote_instead_of_local_cluster}}

PROJECT_DIRECTORY = os.path.realpath(os.path.curdir)
ORIGINAL_PROJECT_DIRECTORY = PROJECT_DIRECTORY
RUNAI_API_DEST_DIR = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'sdk/apis/run/v1'))
OPERAND_DEST_DIR = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'go-operator/internal/run/operands'))
GO_OPERATOR_DIR = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'go-operator'))
OPERANDS_COMMON_PATH = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'go-operator/pkg/operands/common.go'))
MANAGED_OPERANDS_PATH = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'go-operator/controllers/run/managed_operands.go'))
GO_OPERATOR_MAKEFILE = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'go-operator/Makefile'))
CLUSTERROLES_DIR = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'runai-chart/templates/clusterroles'))
PROJECT_ROLE_BINDINGS_PATH = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', 'go-operator/internal/run/operands/krm-operator/resources/project_role_bindings.go'))
WORKFLOW_CI_PATH = os.path.realpath(os.path.join(PROJECT_DIRECTORY, '..', '..', '..', '.github/workflows'))

APP_STRUCT_FIELD_TO_ADD_IN_RUNAICONFIG_STRUCT = '''
	// {{cookiecutter.__app_name_camel_case_upper}} specifies configuration for the {{cookiecutter.app_name}}
	// +kubebuilder:validation:Optional
	{{cookiecutter.__app_name_camel_case_upper}} *{{cookiecutter.__app_name_underscore}}.{{cookiecutter.__app_name_camel_case_upper}} `json:"{{cookiecutter.app_name}},omitempty"`'''
APP_SET_POINTER_TO_ADD_IN_RUNAICONFIG_SET_DEFAULTS_WHERE_NEEDED = '''	if spec.{{cookiecutter.__app_name_camel_case_upper}} == nil {
		spec.{{cookiecutter.__app_name_camel_case_upper}} = &{{cookiecutter.__app_name_underscore}}.{{cookiecutter.__app_name_camel_case_upper}}{}
	}'''
APP_SET_DEFAULTS_WHERE_NEEDED_CALL_TO_ADD_IN_RUNAICONFIG_SET_DEFAULTS_WHERE_NEEDED = '	spec.{{cookiecutter.__app_name_camel_case_upper}}.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry)'
APP_IMPORT_IN_RUNAICONFIG = '	{{cookiecutter.__app_name_underscore}} "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/{{cookiecutter.app_name}}"'
APP_OPERAND_NAME_CONST_TO_ADD_IN_COMMON = '	Runai{{cookiecutter.__app_name_camel_case_upper}}Name = "{{cookiecutter.app_name}}"'
APP_ADD_TO_MANAGED_OPERANDS_LOCAL = '	mo.addOldLocalCluster({{cookiecutter.__app_name_underscore}}.Operand(), *spec.{{cookiecutter.__app_name_camel_case_upper}}.Enabled)'
APP_ADD_TO_MANAGED_OPERANDS_REMOTE = '	mo.addOldRemoteCluster({{cookiecutter.__app_name_underscore}}.Operand(), *spec.{{cookiecutter.__app_name_camel_case_upper}}.Enabled)'
APP_IMPORT_IN_MANAGED_OPERANDS = '	{{cookiecutter.__app_name_underscore}} "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/{{cookiecutter.app_name}}"'
APP_PATH_FOR_OPERATOR_WORKFLOW_CI = '    runai-cluster/cluster/{{cookiecutter.app_name}}'
APP_PROJECT_BASED_ROLEBINDING_TO_ADD_IN_PROJECT_ROLE_BINDINGS = '''		{
			Name:               "{{cookiecutter.app_name}}",
			ClusterRoleName:    "runai-{{cookiecutter.app_name}}-project",
			ServiceAccountName: "{{cookiecutter.app_name}}",
		},'''


def _move_directory(src, dst):
    os.rename(src, dst)

def _rename_file(src, dst):
    os.rename(src, dst)

def _remove_go_tmp_suffix(filepath):
    os.rename(filepath, filepath[:-4])

def _execute(cmd):
    print('Executing: {}'.format(cmd))
    os.system(cmd)

def _insert_lines_to_file_after_line_with_str(filepath, str_lines_to_insert, str_to_find, num_of_lines_after_str_found_to_insert=0):
    with open(filepath, 'r') as f:
        file_lines = f.readlines()

    index_to_insert_at = 0
    for i, line in enumerate(file_lines):
        if line.find(str_to_find) >= 0:
            index_to_insert_at = i + 1 + num_of_lines_after_str_found_to_insert
            break

    if index_to_insert_at == 0:
        return False

    lines_to_insert = [l+'\n' for l in str_lines_to_insert.split('\n')]

    for i, line in enumerate(lines_to_insert):
        file_lines.insert(index_to_insert_at + i, line)

    with open(filepath, 'w') as f:
        f.writelines(file_lines)
    return True

def move_to_engine_if_needed():
    if not IS_ENGINE:
        return

    global PROJECT_DIRECTORY
    global ORIGINAL_PROJECT_DIRECTORY

    print('Moving to engine directory')
    engine_dir = os.path.join(PROJECT_DIRECTORY, '..', '..', 'engine')
    os.chdir(engine_dir)
    src_dir = PROJECT_DIRECTORY
    dst_dir = os.path.join(engine_dir, '{{cookiecutter.app_name}}')
    _move_directory(src_dir, dst_dir)

    # just so this directory exists - will delete later
    ORIGINAL_PROJECT_DIRECTORY = PROJECT_DIRECTORY
    os.mkdir(ORIGINAL_PROJECT_DIRECTORY)
    PROJECT_DIRECTORY = dst_dir
    os.chdir(PROJECT_DIRECTORY)
    print('Project directory is now: {}'.format(PROJECT_DIRECTORY))

def go_mod_init():
    go_mod_file = os.path.join(PROJECT_DIRECTORY, 'go.mod.tmp')
    _remove_go_tmp_suffix(go_mod_file)

    _execute('go mod init')
    _execute('go get ./...')

def move_runai_api_dir():
    runai_api_src_dir = os.path.join(PROJECT_DIRECTORY, '{{cookiecutter.app_name}}-runaiconfig')
    runai_api_dst_dir = os.path.join(RUNAI_API_DEST_DIR, '{{cookiecutter.app_name}}')

    types_go_file = os.path.join(runai_api_src_dir, 'types.go.tmp')
    _remove_go_tmp_suffix(types_go_file)

    print('Moving runai api directory')
    _move_directory(runai_api_src_dir, runai_api_dst_dir)

def move_runai_operand_dir():
    runai_operand_src_dir = os.path.join(PROJECT_DIRECTORY, '{{cookiecutter.app_name}}-operand')
    runai_operand_dst_dir = os.path.join(OPERAND_DEST_DIR, '{{cookiecutter.app_name}}')

    operand_go_file = os.path.join(runai_operand_src_dir, 'operand.go.tmp')
    _remove_go_tmp_suffix(operand_go_file)

    print('Moving runai operand directory')
    _move_directory(runai_operand_src_dir, runai_operand_dst_dir)

def move_rbac_yaml():
    rbac_yaml_src = os.path.join(PROJECT_DIRECTORY, '{{cookiecutter.app_name}}.yaml')
    rbac_yaml_dst = os.path.join(CLUSTERROLES_DIR, '{{cookiecutter.app_name}}.yaml')

    print('Moving RBAC yaml')
    _rename_file(rbac_yaml_src, rbac_yaml_dst)

def add_app_to_runaiconfig_spec():
    # opens the file "runaiconfig_types.go", searches for the go struct RunaiConfigSpec,
    # and adds the app struct as a field to RunaiConfigSpec.
    print('Adding {{cookiecutter.__app_name_camel_case_upper}} struct as a field in runaiconfig_types.go:RunaiConfigSpec')

    runaiconfig_types = os.path.join(RUNAI_API_DEST_DIR, 'runaiconfig_types.go')
    res = _insert_lines_to_file_after_line_with_str(runaiconfig_types, APP_STRUCT_FIELD_TO_ADD_IN_RUNAICONFIG_STRUCT,
                                              'ExternalWorkloadIntegrator *external_workload_integrator.ExternalWorkloadIntegrator')
    if not res:
        raise Exception('Failed to add app struct to runaiconfig_types.go:RunaiConfigSpec')

def add_app_to_runaiconfig_set_defaults_where_needed():
    # opens the file "runaiconfig_types.go", searches for the method SetDefaultsWhereNeeded,
    # and adds a call to the app's SetDefaultsWhereNeeded.
    print('Adding call to {{cookiecutter.__app_name_camel_case_upper}}.SetDefaultsWhereNeeded in runaiconfig_types.go:SetDefaultsWhereNeeded')

    runaiconfig_types = os.path.join(RUNAI_API_DEST_DIR, 'runaiconfig_types.go')

    res = _insert_lines_to_file_after_line_with_str(runaiconfig_types, APP_SET_POINTER_TO_ADD_IN_RUNAICONFIG_SET_DEFAULTS_WHERE_NEEDED,
                                                    'spec.ExternalWorkloadIntegrator = &external_workload_integrator.ExternalWorkloadIntegrator{}',
                                                    1)
    if not res:
        raise Exception('Failed to add pointer to {{cookiecutter.__app_name_camel_case_upper}} in runaiconfig_types.go:SetDefaultsWhereNeeded')

    res = _insert_lines_to_file_after_line_with_str(runaiconfig_types, APP_SET_DEFAULTS_WHERE_NEEDED_CALL_TO_ADD_IN_RUNAICONFIG_SET_DEFAULTS_WHERE_NEEDED,
                                              'spec.ExternalWorkloadIntegrator.SetDefaultsWhereNeeded(')
    if not res:
        raise Exception('Failed to add call to {{cookiecutter.__app_name_camel_case_upper}}.SetDefaultsWhereNeeded in runaiconfig_types.go:SetDefaultsWhereNeeded')

    print('Adding import in runaiconfig_types.go')
    res = _insert_lines_to_file_after_line_with_str(runaiconfig_types, APP_IMPORT_IN_RUNAICONFIG,
                                                    'external_workload_integrator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/external-workload-integrator"')
    if not res:
        raise Exception('Failed to add import in runaiconfig_types.go')

def add_app_to_go_operator_operands():
    print('Adding operand name const to go-operator/pkg/operands/common.go')
    res = _insert_lines_to_file_after_line_with_str(OPERANDS_COMMON_PATH, APP_OPERAND_NAME_CONST_TO_ADD_IN_COMMON,
                                                    'RunaiOperatorName                    = "runai-operator"')
    if not res:
        raise Exception('Failed to add operand name const to go-operator/pkg/operands/common.go')

    if DEPLOY_TO_REMOTE_CLUSTER:
        print('Adding app to managed operands (REMOTE cluster) in go-operator/controllers/run/managed_operands.go')
        res = _insert_lines_to_file_after_line_with_str(MANAGED_OPERANDS_PATH, APP_ADD_TO_MANAGED_OPERANDS_REMOTE,
                                                  'mo.addOldRemoteCluster(researcher_service.Operand(), *spec.ResearcherService.Enabled && !spec.Internal.Development.Headless)')
        if not res:
            raise Exception('Failed to add app to managed operands in go-operator/controllers/run/managed_operands.go')
    else:
        print('Adding app to managed operands (LOCAL cluster) in go-operator/controllers/run/managed_operands.go')
        res = _insert_lines_to_file_after_line_with_str(MANAGED_OPERANDS_PATH, APP_ADD_TO_MANAGED_OPERANDS_LOCAL,
                                                  'mo.addOldLocalCluster(external_workload_integrator.Operand(), *spec.ExternalWorkloadIntegrator.Enabled)')
        if not res:
            raise Exception('Failed to add app to managed operands in go-operator/controllers/run/managed_operands.go')

    print('Adding import in go-operator/controllers/run/managed_operands.go')
    res = _insert_lines_to_file_after_line_with_str(MANAGED_OPERANDS_PATH, APP_IMPORT_IN_MANAGED_OPERANDS,
                                                    'external_workload_integrator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/external-workload-integrator"')
    if not res:
        raise Exception('Failed to add import in go-operator/controllers/run/managed_operands.go')

def add_project_based_rbac_if_needed():
    if not {{cookiecutter.add_project_based_rolebinding}}:
        print('Skipping project-based RoleBinding in project_role_bindings.go')
        return
    print('Adding project-based RoleBinding in project_role_bindings.go')

    res = _insert_lines_to_file_after_line_with_str(PROJECT_ROLE_BINDINGS_PATH, APP_PROJECT_BASED_ROLEBINDING_TO_ADD_IN_PROJECT_ROLE_BINDINGS,
                                                    '{Name: "cluster-sync", ClusterRoleName: "cluster-sync-project", ServiceAccountName: "cluster-sync"},')
    if not res:
        raise Exception('Failed to add project-based RoleBinding in project_role_bindings.go')

def run_make_generate_manifests_copycrds():
    os.chdir(GO_OPERATOR_DIR)
    _execute('make generate manifests')

move_to_engine_if_needed()
go_mod_init()
add_app_to_runaiconfig_spec()
add_app_to_runaiconfig_set_defaults_where_needed()
add_app_to_go_operator_operands()
move_runai_api_dir()
move_runai_operand_dir()
move_rbac_yaml()
add_project_based_rbac_if_needed()
run_make_generate_manifests_copycrds()

if IS_ENGINE and ORIGINAL_PROJECT_DIRECTORY != PROJECT_DIRECTORY:
  os.rmdir(ORIGINAL_PROJECT_DIRECTORY)
