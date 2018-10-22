Name:		  nex
Version:	0.1.1
Release:	1%{?dist}
Summary:	Software defined DHCP/DNS

License:	Apache-2.0
URL:		  https://gitlab.com/mergetb/nex

BuildRequires: gcc-c++ openssl cpprest-devel
BuildRequires: protobuf       >= 3.6.1
BuildRequires: grpc-devel     >= 2.0.0
BuildRequires: etcd_cpp_apiv3 >= 0.0.1
BuildRequires: kea-devel      >= 1.4.0_P1

Requires: golang protobuf     >= 3.6.1
Requires: grpc-devel          >= 2.0.0
Requires: etcd_cpp_apiv3      >= 0.0.1
Requires: kea-libs            >= 1.4.0_P1
Requires: kea                 >= 1.4.0_P1

%description
Nex is the spirit of dnsmasq meets modern scale out architecture. Nex is 
designed around a stateless service model for both Kea and CoreDNS - with an 
etcd deployment that brings them together into a cohesive system - ala 
dnsmasq.


%prep
go get github.com/golang/protobuf/protoc-gen-go
mkdir -p src/gitlab.com/mergetb
cp -r ../SOURCES/nex src/gitlab.com/mergetb/

%build
export GOPATH=`pwd`
cd src/gitlab.com/mergetb/nex
dep ensure
make clean && make %{?_smp_mflags}
cd kea
make clean && make %{?_smp_mflags}


%install
cd src/gitlab.com/mergetb/nex
export prefix=%{buildroot}/usr
make install
cd kea
export prefix=%{buildroot}/usr
make install
curl -L https://github.com/mergetb/coredns/releases/download/v1.2.2-nex/coredns -o %{buildroot}/usr/bin/coredns
chmod a+x %{buildroot}/usr/bin/coredns

%files
/usr/bin/coredns
/usr/bin/loader
/usr/bin/nexd
/usr/lib/kea/hooks/nex.so
/lib/systemd/system/coredns.service
/lib/systemd/system/nexd.service


%changelog
* Fri Oct 19 2018 Ryan Goodfellow <rgoodfel@isi.edu> - 0.1.1-0
- Add nexd service

* Fri Aug 24 2018 Ryan Goodfellow <rgoodfel@isi.edu> - 0.1.0-0
- Initial
