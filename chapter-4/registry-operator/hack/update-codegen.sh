#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# 根目录
SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

go mod vendor

MODULE="setting.zhihu.com"  # go mod 需要对应
OUTPUT_PACKAGE="generated"  # 输出目录
APIS_PACKAGE="api"          # api目录
GROUP="repo"                # group信息
VERSION="v1beta1"           # crd version
CODEGEN_PKG="./vendor/k8s.io/code-generator"


trap EXIT SIGINT SIGTERM

# 创建临时目录
GENERATED_TMP_DIR=$(mktemp -d)

# 判断目录是否存在，并删除原本目录
if [ --e ${APIS_PACKAGE}/${VERSION} ]; then
    mkdir -p "${APIS_PACKAGE}/${GROUP}" && cp -r "${APIS_PACKAGE}/${VERSION}" "${APIS_PACKAGE}/${GROUP}"
    rm -rf ${APIS_PACKAGE}/${VERSION}
fi

chmod +x ${CODEGEN_PKG}/generate-groups.sh

# 生成client,informer,lister，输出到临时目录
${CODEGEN_PKG}/generate-groups.sh "client,informer,lister" \
  ${MODULE}/${OUTPUT_PACKAGE} ${MODULE}/${APIS_PACKAGE} \
  "${GROUP}:${VERSION}" \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt \
  --output-base "${GENERATED_TMP_DIR}"

# 复制文件到generated中
cp -rf "${GENERATED_TMP_DIR}"/${MODULE}/* ${OUTPUT_PACKAGE}

# 删除临时目录
rm -r "${GENERATED_TMP_DIR}"

