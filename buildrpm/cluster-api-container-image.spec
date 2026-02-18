

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global debug_package   %{nil}
%endif

%{!?registry: %global registry container-registry.oracle.com/olcne}
%global app_name               cluster-api
%global app_version            1.11.6
%global oracle_release_version 1
%global _buildhost             build-ol%{?oraclelinux}-%{?_arch}.oracle.com

Name:           %{app_name}-container-image
Version:        %{app_version}
Release:        %{oracle_release_version}%{?dist}
Summary:        Cluster API provides declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.
License:        Apache-2.0
Group:          System/Management
Url:            https://github.com/kubernetes-sigs/cluster-api.git
Source:         %{name}-%{version}.tar.bz2


%description
Cluster API is a Kubernetes subproject focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

%prep
%setup -q -n %{name}-%{version}

%build
%global rpm_name %{app_name}-%{version}-%{release}.%{_build_arch}
%global app_name_capi_controller cluster-api-controller
%global docker_tag_capi_controller %{registry}/%{app_name_capi_controller}:v%{version}
%global app_name_kubeadm_bootstrap kubeadm-bootstrap-controller
%global docker_tag_kubeadm_bootstrap %{registry}/%{app_name_kubeadm_bootstrap}:v%{version}
%global app_name_kubeadm_control_plane kubeadm-control-plane-controller
%global docker_tag_kubeadm_control_plane %{registry}/%{app_name_kubeadm_control_plane}:v%{version}
%global app_name_clusterctl clusterctl
%global docker_tag_clusterctl %{registry}/%{app_name_clusterctl}:v%{version}
%global rpm_name_clusterctl %{app_name_clusterctl}-%{version}-%{release}.%{_build_arch}

yum clean all
yumdownloader --destdir=${PWD}/rpms %{rpm_name}
yumdownloader --destdir=${PWD}/rpms %{rpm_name_clusterctl}

docker build --pull \
    --build-arg https_proxy=${https_proxy} \
    --build-arg RPM=%{rpm_name}.rpm \
    --build-arg managerImage=/cluster-api/manager \
    -t %{docker_tag_capi_controller} -f ./olm/builds/Dockerfile .
docker save -o %{app_name_capi_controller}.tar %{docker_tag_capi_controller}

docker build --pull \
    --build-arg https_proxy=${https_proxy} \
    --build-arg RPM=%{rpm_name}.rpm \
    --build-arg managerImage=/cluster-api/kubeadm-bootstrap-manager \
    -t %{docker_tag_kubeadm_bootstrap} -f ./olm/builds/Dockerfile .
docker save -o %{app_name_kubeadm_bootstrap}.tar %{docker_tag_kubeadm_bootstrap}

docker build --pull \
    --build-arg https_proxy=${https_proxy} \
    --build-arg RPM=%{rpm_name}.rpm \
    --build-arg managerImage=/cluster-api/kubeadm-control-plane-manager \
    -t %{docker_tag_kubeadm_control_plane} -f ./olm/builds/Dockerfile .
docker save -o %{app_name_kubeadm_control_plane}.tar %{docker_tag_kubeadm_control_plane}

docker build --pull \
    --build-arg https_proxy=${https_proxy} \
    --build-arg RPM=%{rpm_name_clusterctl}.rpm \
    -t %{docker_tag_clusterctl} -f ./olm/builds/DockerfileClusterctl .
docker save -o %{app_name_clusterctl}.tar %{docker_tag_clusterctl}

%install
%__install -D -m 644 %{app_name_capi_controller}.tar %{buildroot}/usr/local/share/olcne/%{app_name_capi_controller}.tar
%__install -D -m 644 %{app_name_kubeadm_bootstrap}.tar %{buildroot}/usr/local/share/olcne/%{app_name_kubeadm_bootstrap}.tar
%__install -D -m 644 %{app_name_kubeadm_control_plane}.tar %{buildroot}/usr/local/share/olcne/%{app_name_kubeadm_control_plane}.tar
%__install -D -m 644 %{app_name_clusterctl}.tar %{buildroot}/usr/local/share/olcne/%{app_name_clusterctl}.tar

%files
%license LICENSE THIRD_PARTY_LICENSES.txt olm/SECURITY.md
/usr/local/share/olcne/%{app_name_capi_controller}.tar
/usr/local/share/olcne/%{app_name_kubeadm_bootstrap}.tar
/usr/local/share/olcne/%{app_name_kubeadm_control_plane}.tar
/usr/local/share/olcne/%{app_name_clusterctl}.tar

%changelog
* Wed Feb 18 2026 Oracle Cloud Native Environment Authors <noreply@oracle.com> - 1.11.6-1
- Added Oracle specific build files for cluster-api.
