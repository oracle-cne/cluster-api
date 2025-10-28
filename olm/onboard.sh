#!/bin/bash -x

# This script is run when the branch for a new release are initially onboarded.

if [ "$(uname)" == "Linux" ]; then
  echo "Installing the build dependencies required for generating the helm charts"
  dnf config-manager --add-repo https://yum.oracle.com/repo/OracleLinux/OL8/olcne18/x86_64
  dnf config-manager --enable olcne_incubator
  dnf install -y yq helmify clusterctl
fi

echo "Generating the helm charts"
mkdir charts
./bin/clusterctl generate provider --core cluster-api | helmify -crd-dir charts/core-capi
./bin/clusterctl generate provider --bootstrap kubeadm | helmify -crd-dir charts/bootstrap-capi
./bin/clusterctl generate provider --control-plane kubeadm | helmify -crd-dir charts/control-plane-capi
./bin/clusterctl generate provider --infrastructure oci | helmify -crd-dir charts/oci-capi

echo "Customizing values for core-capi"
yq -i '.controllerManager.manager.image.repository = "olcne/cluster-api-controller"' charts/core-capi/values.yaml

yq -i '.description = "Helm chart for Cluster API core provider"' charts/core-capi/Chart.yaml

# Set the chart versions to match the image tag version
coreCapiVersion=$(yq '.controllerManager.manager.image.tag' charts/core-capi/values.yaml)
coreCapiVersion=${coreCapiVersion:1}

for chart in core-capi bootstrap-capi control-plane-capi; do
	echo "fullnameOverride: capi" >> charts/$chart/values.yaml

	echo "icon: icons/capi.svg" >> charts/$chart/Chart.yaml
	yq -i ".appVersion = \"$coreCapiVersion\"" charts/$chart/Chart.yaml
	yq -i ".version = \"$coreCapiVersion\"" charts/$chart/Chart.yaml
done


echo "Customizing values for oci-capi"
yq -i '.description = "Helm chart for Cluster API Provider OCI"' charts/oci-capi/Chart.yaml
echo "icon: icons/capi.svg" >> charts/oci-capi/Chart.yaml


# Set the chart versions to match the image tag version
ociCapiVersion=$(yq '.controllerManager.manager.image.tag' charts/oci-capi/values.yaml)
yq -i ".appVersion = \"$ociCapiVersion\"" charts/oci-capi/Chart.yaml
ociCapiSemVer=${ociCapiVersion:1}
yq -i ".version = \"$ociCapiSemVer\"" charts/oci-capi/Chart.yaml
