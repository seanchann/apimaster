#!/usr/bin/env bash
#
# check you installed the following tools:
#  GOPROXY=off go install k8s.io/code-generator/cmd/deepcopy-gen
#  GOPROXY=off go install k8s.io/code-generator/cmd/defaulter-gen
#  GOPROXY=off go install k8s.io/code-generator/cmd/conversion-gen
#  GOPROXY=off go install k8s.io/code-generator/cmd/client-gen
#  GOPROXY=off go install k8s.io/code-generator/cmd/lister-gen
#  GOPROXY=off go install k8s.io/code-generator/cmd/informer-gen
#  GOPROXY=off go install k8s.io/code-generator/cmd/openapi-gen



set -o errexit
set -o nounset
set -o pipefail

unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    CYGWIN*)    machine=Cygwin;;
    MINGW*)     machine=MinGw;;
    *)          machine="UNKNOWN:${unameOut}"
esac

READLINK=readlink
if [ "$machine" = "Mac" ];then
  READLINK=greadlink
fi
CURRENT_DIR=`dirname $($READLINK -f $0)`
SCRIPT_ROOT=${CURRENT_DIR%%/hack}


# This tool wants a different default than usual.
VERBOSE="${VERBOSE:-1}"
BOILERPLATE_FILENAME=${SCRIPT_ROOT}/hack/boilerplate.go.txt
GENERATED_FILE_PREFIX="${GENERATED_FILE_PREFIX:-zz_generated.}"
OUT_DIR="output"

CLIENT_OUTPUT_DIR=${SCRIPT_ROOT}/pkg/client/generated
CLIENT_OUTPUT_PKG="github.com/seanchann/apimaster/pkg/client/generated"

CLIENTSET_OUTPUT_DIR=${CLIENT_OUTPUT_DIR}
CLIENTSET_OUTPUT_PKG=${CLIENT_OUTPUT_PKG}
CLIENTSET_VERSIONED_OUTPUT_PKG=${CLIENTSET_OUTPUT_PKG}/clientset

LISTERS_OUTPUT_DIR=${CLIENT_OUTPUT_DIR}/listers
LISTERS_OUTPUT_PKG=${CLIENT_OUTPUT_PKG}/listers

INFORMERS_OUTPUT_DIR=${CLIENT_OUTPUT_DIR}/informers
INFORMERS_OUTPUT_PKG=${CLIENT_OUTPUT_PKG}/informers

if [ ! -d "${OUT_DIR}" ]; then
  mkdir -p "${OUT_DIR}"
fi


function set_tag_pkgs() {
  local apis_pkg=${SCRIPT_ROOT}/pkg/apis/

  gv_dirs=()
  gv_dirs+=("coreres/v1")
  gv_dirs+=("rbac/v1")

  tag_pkgs=()
  for pkg in "${gv_dirs[@]}"; do
    tag_pkgs+=("${apis_pkg}/${pkg}")
  done

  tag_pkgs_internal=()
  for pkg in "${gv_dirs[@]}"; do
    local internal_pkg=${pkg%%/*}
    tag_pkgs_internal+=("${apis_pkg}/${internal_pkg}")
  done

  apimachinery_pkgs=()
  apimachinery_pkgs+=("k8s.io/apimachinery/pkg/apis/meta/v1")
  apimachinery_pkgs+=("k8s.io/apimachinery/pkg/version")
  apimachinery_pkgs+=("k8s.io/apimachinery/pkg/runtime")
  apimachinery_pkgs+=("k8s.io/apimachinery/pkg/util/intstr")
}

function codegen_deepcopy() {
  set_tag_pkgs
  for pkg in "${tag_pkgs_internal[@]}"; do
    tag_pkgs+=("$pkg")
  done

  # The result file, in each pkg, of defaulter generation.
  local output_file="${GENERATED_FILE_PREFIX}deepcopy.go"


  deepcopy-gen \
    --go-header-file ${BOILERPLATE_FILENAME} \
    --output-file ${output_file} \
    "${tag_pkgs[@]}"
}

function codegen_defaults() {
  set_tag_pkgs
  
  # The result file, in each pkg, of defaulter generation.
  local output_file="${GENERATED_FILE_PREFIX}defaults.go"

  defaulter-gen \
    --go-header-file ${BOILERPLATE_FILENAME} \
    --output-file ${output_file} \
    "${tag_pkgs[@]}"
}

function codegen_conversions() {
  set_tag_pkgs

  # The result file, in each pkg, of conversion generation.
  local output_file="${GENERATED_FILE_PREFIX}conversion.go"

  local extra_peer_pkgs=(
  )

  # this for if you have extra peer packages
  # conversion-gen \
  #   -v "${VERBOSE}" \
  #   --go-header-file "${BOILERPLATE_FILENAME}" \
  #   --output-file "${output_file}" \
  #   $(printf -- " --extra-peer-dirs %s" "${extra_peer_pkgs[@]}") \
  #   "${tag_pkgs[@]}" \
  #   "$@"

  conversion-gen \
    -v "${VERBOSE}" \
    --go-header-file "${BOILERPLATE_FILENAME}" \
    --output-file "${output_file}" \
    "${tag_pkgs[@]}" \
    "$@"
}

function codegen_clients() {
  set_tag_pkgs

  client-gen \
    -v "${VERBOSE}" \
    --go-header-file "${BOILERPLATE_FILENAME}" \
    --output-dir "${CLIENTSET_OUTPUT_DIR}" \
    --output-pkg="${CLIENTSET_OUTPUT_PKG}" \
    --clientset-name="clientset" \
    --input-base="${SCRIPT_ROOT}/pkg/apis" \
    $(printf -- " --input %s" "${gv_dirs[@]}") \
    "$@"
}

function codegen_listers() {

  local ext_apis=()
  set_tag_pkgs
  for pkg in "${tag_pkgs[@]}"; do
    ext_apis+=("$pkg")
  done

  lister-gen \
      -v "${VERBOSE}" \
      --go-header-file "${BOILERPLATE_FILENAME}" \
      --output-dir "${LISTERS_OUTPUT_DIR}" \
      --output-pkg "${LISTERS_OUTPUT_PKG}" \
      "${ext_apis[@]}" \
      "$@"
}

function codegen_informers() {
  local ext_apis=()
  set_tag_pkgs
  for pkg in "${tag_pkgs[@]}"; do
    ext_apis+=("$pkg")
  done

  informer-gen \
    -v "${VERBOSE}" \
    --go-header-file "${BOILERPLATE_FILENAME}" \
    --output-dir "${INFORMERS_OUTPUT_DIR}" \
    --output-pkg "${INFORMERS_OUTPUT_PKG}" \
    --single-directory \
    --versioned-clientset-package "${CLIENTSET_VERSIONED_OUTPUT_PKG}" \
    --listers-package "${LISTERS_OUTPUT_PKG}" \
    "${ext_apis[@]}" \
    "$@"
}


function codegen_openapi() {
  set_tag_pkgs
  for pkg in "${apimachinery_pkgs[@]}"; do
    tag_pkgs+=("$pkg")
  done
  

  # The result file, in each pkg, of open-api generation.
  local output_file="${GENERATED_FILE_PREFIX}openapi.go"

  local output_dir="pkg/generated/openapi"
  local output_pkg="github.com/seanchann/apimaster/${output_dir}"
  local report_file="${OUT_DIR}/api_violations.report"

  openapi-gen \
    -v "${VERBOSE}" \
    --go-header-file "${BOILERPLATE_FILENAME}" \
    --output-file "${output_file}" \
    --output-dir "${output_dir}" \
    --output-pkg "${output_pkg}" \
    --report-filename "${report_file}" \
    "${tag_pkgs[@]}" \
    "$@"
}

codegen_deepcopy
codegen_defaults
codegen_conversions
codegen_openapi

codegen_clients
codegen_listers
codegen_informers
