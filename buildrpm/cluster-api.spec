

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global debug_package %{nil}
%endif

%global app_name                cluster-api
%global app_version             1.10.10
%global oracle_release_version  1
%global _buildhost              build-ol%{?oraclelinux}-%{?_arch}.oracle.com

Name:           %{app_name}
Version:        %{app_version}
Release:        %{oracle_release_version}%{?dist}
Summary:        Cluster API provides declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.
License:        Apache-2.0
Group:          System/Management
Url:            https://github.com/kubernetes-sigs/cluster-api.git
Source:         %{name}-%{version}.tar.bz2
BuildRequires:  golang >= 1.20.12
BuildRequires:	make

%package -n clusterctl
Summary:    A CLI tool that handles the lifecycle of a Cluster API management cluster.

%description
Cluster API is a Kubernetes subproject focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

%description -n clusterctl
A CLI tool that handles the lifecycle of a Cluster API management cluster.

%prep
%setup -q -n %{name}-%{version}

%build
git fetch --tags
go mod download
make manager-core
make manager-kubeadm-bootstrap
make manager-kubeadm-control-plane
make clusterctl

%install
install -m 755 -d %{buildroot}/%{app_name}
install -m 755 bin/* %{buildroot}/%{app_name}
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 bin/clusterctl %{buildroot}%{_bindir}/clusterctl

%files
%license LICENSE THIRD_PARTY_LICENSES.txt olm/SECURITY.md
/%{app_name}/

%files -n clusterctl
%license LICENSE THIRD_PARTY_LICENSES.txt olm/SECURITY.md
%{_bindir}/clusterctl

%changelog
* Thu Jan 08 2026 Oracle Cloud Native Environment Authors <noreply@oracle.com> - 1.10.10-1
- Added Oracle specific build files for cluster-api.
