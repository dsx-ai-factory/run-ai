#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


sdk_version="{{ or .SdkVersion "${RUNAI_CLI_SDK_VERSION}" }}"
cp_version="{{ or .CpVersion "${RUNAI_CLI_CP_VERSION}" }}"
runai_home_dir="${HOME}/.runai"
default_cache_file_dir="/tmp/runai/cache"
cache_file_name="version.json"
cp_url="{{ or .CpURL "${RUNAI_CP_URL}" }}"
auth_url="{{ or .AuthURL "${RUNAI_AUTH_URL}" }}"
upgrade_mode=false
validate_old_version=true
curl_max_time=""
configured_shell=""
shell_config_file=""
shell_configuration_supported=false
shell_config_file_available=false
path_configured=false
autocomplete_configured=false
path_configuration_failed=false
autocomplete_configuration_failed=false

# Define colors
RED='\033[0;31m'
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

### Script Functions Section ###
logError() {
    echo -e "${RED} ▸ Error: ${NC}$1" >&2
}

logInfo() {
    echo -e "${GREEN} ▸ ${NC}$1"
}

logWarning() {
    echo -e "${ORANGE} ▸ Warning: ${NC}$1"
}

logQuestion() {
    printf "${PURPLE} ▸ ${NC}%s" "$1"
}
# Function to display usage
usage() {
  echo "Usage: $0 [-u|--upgrade] [-s|--skip-old-version-validation] [-m|--max-time <seconds>] [-c|--cache-directory <directory>]"
  exit 1
}

should_override_home_path() {
      if [ -z "${RUNAI_CLI_INSTALLATION_PATH}" ]; then
          echo "${runai_home_dir}"
      else
          echo "${RUNAI_CLI_INSTALLATION_PATH}"
      fi
}

check_dependencies() {
    # Check if the curl package is installed
    if ! [ -x "$(command -v curl)" ]; then
        logError "Error: curl is not installed. Please install it first"
        exit 1
    fi
}

check_old_version() {
  # Check if the old version v1 of the CLI is installed
  binary_installation_path=$(which runai) || {
    logInfo "Old version of the CLI is not installed"
    return
  }

  if [[ "${binary_installation_path}" != "${soft_link_path}"  ]]; then
    logError "Old version of the CLI is installed in ${binary_installation_path}"
    logError "Please remove the old version of the CLI by running 'sudo rm -f ${binary_installation_path}' and re-run the installer"
    exit 1
  fi
}

# Function to determine the operating system
check_os() {
    OS_TYPE="Unknown operating system"
    # Check if the operating system is Linux
    if [ -f /etc/os-release ]; then
        OS_TYPE="linux"
    # Check if the operating system is macOS
    elif [ -f /System/Library/CoreServices/SystemVersion.plist ]; then
        if [[ $(uname) == "Darwin" ]]; then
          OS_TYPE="darwin"
        fi
    fi
    # Return the operating system type
    echo "$OS_TYPE"
}

validate_arch() {
  arch=$(uname -m)
  case $arch in
  arm64)
    ;;
  aarch64)
    arch="arm64"
    ;;
  x86_64)
    arch="amd64"
    ;;
  *)
    logError "Unsupported architecture: ${arch}"
    exit 1
    ;;
  esac
  echo "$arch"
}

configRunaiServers() {
   logQuestion "Enable auto-update? (Y/n): "
   read -r autoupdate_input

   autoupdate_input_lower=$(echo "$autoupdate_input" | tr '[:upper:]' '[:lower:]')

   case "$autoupdate_input_lower" in
     n|no)
       autoupdate_answer="disabled"
       ;;
     *)
       autoupdate_answer="enabled"
       ;;
   esac

   export RUNAI_CLI_CONFIG_PATH=${RUNAI_CLI_CONFIG_PATH:-${runai_home_dir}}
   ${soft_link_path} config set --cp-url="${cp_url}" -q || {
     logError "Failed to configure control plane URL"
   }
   ${soft_link_path} config set --auto-update="${autoupdate_answer}" -q || {
     logError "Failed to configure auto update"
   }
   if [ -n "${cache_dir_input}" ]; then
       ${soft_link_path} config set --cache-directory="${cache_dir_input}" -q || {
           logError "Failed to configure the cache directory"
         }
   fi
   echo "Configuration completed."
 }

detectShell() {
  configured_shell="${SHELL:-}"
  configured_shell="${configured_shell##*/}"

  case "${configured_shell}" in
    bash)
      shell_config_file="${HOME}/.bashrc"
      shell_configuration_supported=true
      ;;
    zsh)
      shell_config_file="${HOME}/.zshrc"
      shell_configuration_supported=true
      ;;
    *)
      shell_config_file=""
      shell_configuration_supported=false
      ;;
  esac
}

autocomplete() {
  logQuestion "Configure auto-completion? (Y/n): "
  read -r answer

  # Default to "Y" if the answer is empty
  answer="${answer:-Y}"

  # Convert answer to lowercase for easier comparison
  answer_lowercase=$(echo "$answer" | tr '[:upper:]' '[:lower:]')

  if [[ "$answer_lowercase" != "yes" && "$answer_lowercase" != "y" ]]; then
    return
  fi

  completion_command="source <(${soft_link_path} completion ${configured_shell})"
  if grep -q "${soft_link_path} completion ${configured_shell}" "${shell_config_file}"; then
    autocomplete_configured=true
  elif echo "${completion_command}" >> "${shell_config_file}"; then
    autocomplete_configured=true
  else
    autocomplete_configuration_failed=true
  fi
}



addCliToPath() {
  logQuestion "Configure PATH variable? (Y/n): "
  read -r answer

  # Default to "Y" if the answer is empty
  answer="${answer:-Y}"

  # Convert answer to lowercase for easier comparison
  answer_lowercase=$(echo "$answer" | tr '[:upper:]' '[:lower:]')

  if [[ "$answer_lowercase" != "yes" && "$answer_lowercase" != "y" ]]; then
    return
  fi

  path_export="export PATH=\$PATH:${install_path}"
  if grep -q "${install_path}" "${shell_config_file}"; then
    path_configured=true
  elif echo "${path_export}" >> "${shell_config_file}"; then
    path_configured=true
  else
    path_configuration_failed=true
  fi
}

configureShell() {
  detectShell

  if [ "${shell_configuration_supported}" = false ]; then
    logWarning "Does not support adding the CLI to the PATH variable for this shell; supported shells are bash and zsh"
    return
  fi

  if [ ! -f "${shell_config_file}" ]; then
    return
  fi

  shell_config_file_available=true
  addCliToPath
  autocomplete
}

printShellConfigurationSummary() {
  if [ "${shell_configuration_supported}" = false ] || [ "${shell_config_file_available}" = false ]; then
    echo "Add '${install_path}' to PATH in your shell's startup file."
    echo "Auto-completion was not configured."
  else
    if [ "${path_configured}" = true ] || [ "${autocomplete_configured}" = true ]; then
      echo "To apply shell changes, run 'source ${shell_config_file}'."
    fi
    if [ "${path_configuration_failed}" = true ]; then
      echo "Add '${install_path}' to PATH in your shell's startup file."
    fi
    if [ "${autocomplete_configuration_failed}" = true ]; then
      echo "Auto-completion was not configured."
    fi
    if [ "${path_configured}" = false ] && [ "${autocomplete_configured}" = false ] &&
       [ "${path_configuration_failed}" = false ] && [ "${autocomplete_configuration_failed}" = false ]; then
      echo "Shell configuration was skipped."
    fi
  fi
}

installUnixCliBinary() {
  arch=$(validate_arch)
  echo "## Installing..."

  url_download="${cp_url}/api/v1/cli-exposer/dist/${os}/${arch}/runai"
  bin_dir="${install_path}/${sdk_version}"
  bin_path="${bin_dir}/runai"

  logInfo "Preparing directory      (${bin_dir})"
  if [ ! -d "${bin_dir}" ]
  then
       mkdir -p "${bin_dir}" || {
         logError "Failed to create directory: ${bin_dir}"
         exit 1
       }
  # else
  #      echo "Directory exists ${bin_dir}"
  fi

  logInfo "Downloading binary       (${url_download})"

  if [ -n "${curl_max_time}" ]; then
    curl_max_time_flag=" --max-time ${curl_max_time}"
  fi

  curl -o "${bin_path}""${curl_max_time_flag}" -H "Authorization: Bearer {{ or .Token "${TOKEN}" }}" -fsSL "${url_download}" || {
    logError "Failed to download binary from ${url_download}"
    exit 1
  }

  chmod u+x ${bin_path} || {
    logError "Failed to set execute permission on ${bin_path}"
    exit 1
  }

  chmod 755 ${bin_path} || {
    logError "Failed to set execute permission on ${bin_path}"
    exit 1
  }

  ln -nsf ${bin_path} ${soft_link_path} || {
    logError "Failed to create soft link: ${soft_link_path}"
    exit 1
  }

  logInfo "Linking executable       (${soft_link_path})"
}

validate_number() {
  re='^[0-9]+([.][0-9]+)?$'
  if ! [[ $1 =~ $re ]] ; then
    logError "max-time: Not a number" >&2
    usage
  fi
}

validate_directory() {
  local dir_path="$1"

  # Check if input is empty
  if [ -z "$dir_path" ]; then
    logError "Error: Directory path cannot be empty."
    usage
  fi

  # Check if path exists
  if [ ! -e "$dir_path" ]; then
    logError "Error: '$dir_path' does not exist."
    usage
  fi

  # Check if path is a directory
  if [ ! -d "$dir_path" ]; then
    logError "Error: '$dir_path' is not a directory."
    usage
  fi

  # Check if directory is writable
  if [ ! -w "$dir_path" ]; then
    logError "Error: Directory '$dir_path' is not writable."
    usage
  fi
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -s|--skip-old-version-validation)
      echo
      logInfo "Skipping old version validation"
      validate_old_version=false
      shift # past argument
      ;;
    -m|--max-time)
      echo
      curl_max_time="${2}"
      validate_number "${curl_max_time}"
      logInfo "Set curl maximum time, ${curl_max_time} seconds to spend before curl exits with a timeout error code"
      shift # past argument
      shift # past value
      ;;
    -c|--cache-directory)
      echo
      cache_dir_input="${2}"
      validate_directory "${cache_dir_input}"
      logInfo "Set cache directory to ${cache_dir_input}"
      shift # past argument
      shift # past value
      ;;
    -u|--upgrade|upgrade)
      echo
      logInfo "Upgrade mode enabled"
      upgrade_mode=true
      shift # past argument
      ;;
    *)    # unknown option
      logError "Invalid argument: ${1}"
      usage
  esac
done

#main starts here
runai_home_dir=$(should_override_home_path)
install_path="${runai_home_dir}/bin"
soft_link_path="${install_path}/runai"

check_dependencies
if [ "${validate_old_version}" = true ]; then
  check_old_version
fi


# Call the function to check the operating system
os=$(check_os)

echo
echo "# Run:ai CLI Installer"
echo "Version:       ${cp_version}"
echo "SDK version:   ${sdk_version}"
echo "OS:            ${os}"
echo "Target:        ${install_path}/${sdk_version}"
echo

case $os in
linux)
  installUnixCliBinary
  ;;
darwin)
  installUnixCliBinary
  ;;

  *)
    logError "OS \"${os}\" installation is not supported by this script"
    exit 1
    ;;
esac

echo "Installation completed."
echo

if [ "${upgrade_mode}" = false ]; then
  echo "## Configuring..."
  configureShell
  configRunaiServers
  echo
  echo "## Finished!"
  echo
  echo "Run:ai CLI installed successfully."
  printShellConfigurationSummary
  echo
else
  echo "Upgrade mode enabled; skipped configurations steps"
fi

echo "Next steps:"
echo "  'runai login'    - Connect to your Run:ai cluster"
echo "  'runai version'  - Verify installation"
echo "  'runai help'     - See available commands"

exit 0
