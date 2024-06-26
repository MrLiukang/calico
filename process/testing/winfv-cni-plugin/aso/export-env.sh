export SUFFIX="${SUFFIX:=${USER}}"

export AZURE_LOCATION="${AZURE_LOCATION:="eastus2"}"
export AZURE_RESOURCE_GROUP="${AZURE_RESOURCE_GROUP:=rg-winfv-${SUFFIX}}"

#export AZURE_WINDOWS_IMAGE_SKU="${AZURE_WINDOWS_IMAGE_SKU:="2022-datacenter-core-g2"}"
#export AZURE_WINDOWS_IMAGE_VERSION="${AZURE_WINDOWS_IMAGE_VERSION:="20348.2402.240405"}"

export AZURE_WINDOWS_IMAGE_SKU="${AZURE_WINDOWS_IMAGE_SKU:="2019-datacenter-core-g2"}"
export AZURE_WINDOWS_IMAGE_VERSION="${AZURE_WINDOWS_IMAGE_VERSION:="17763.5696.240406"}"

export LINUX_NODE_COUNT="${LINUX_NODE_COUNT:=1}"
export WINDOWS_NODE_COUNT="${WINDOWS_NODE_COUNT:=1}"

export KUBE_VERSION="${KUBE_VERSION:="1.28.7"}"
export CONTAINERD_VERSION="${CONTAINERD_VERSION:="1.6.6"}"

export SSH_KEY_FILE="$PWD/.sshkey"
