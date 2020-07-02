#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# corresponding to go mod init <module>
MODULE=example.com/foo-controller
# api package
APIS_PKG=apis
# generated output package
OUTPUT_PKG=generated/webapp
# group-version such as foo:v1alpha1
GROUP_VERSION=webapp:v1

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash "${CODEGEN_PKG}"/generate-groups.sh "client,lister,informer" \
  ${MODULE}/${OUTPUT_PKG} ${MODULE}/${APIS_PKG} \
  ${GROUP_VERSION} \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/.."
#  --output-base "$(dirname "${BASH_SOURCE[0]}")/../../.." \

# To use your own boilerplate text append:
#   --go-header-file "${SCRIPT_ROOT}"/hack/custom-boilerplate.go.txt