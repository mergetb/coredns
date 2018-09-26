Name:		  nex
Version:	0.1.0
Release:	1%{?dist}
Summary:	Software defined DHCP/DNS

License:	Apache-2.0
URL:		  https://gitlab.com/mergetb/nex

BuildRequires: gcc-c++ openssl cpprest-devel protobuf-devel 
BuildRequires: protobuf-c-devel golang 
# BuildRequires: godep
# BuildRequires: https://mirror.deterlab.net/merge/kea-libs-1.4.0_P1-1.fc28.x86_64.rpm
# BuildRequires: https://mirror.deterlab.net/merge/kea-devel-1.4.0_P1-1.fc28.x86_64.rpm
# BuildRequires: https://mirror.deterlab.net/merge/grpc-devel-2.0.0-1.fc28.x86_64.rpm
# BuildRequires: https://mirror.deterlab.net/merge/etcd_cpp_apiv3-0.1.1-Linux.rpm

# Requires:	https://mirror.deterlab.net/merge/kea-1.4.0_P1-1.fc28.x86_64.rpm

%description
Nex is the spirit of dnsmasq meets modern scale out architecture. Nex is 
designed around a stateless service model for both Kea and CoreDNS - with an 
etcd deployment that brings them together into a cohesive system - ala 
dnsmasq.


%prep
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

%files
/usr/bin/loader
/usr/lib/kea/hooks/nex.so


%changelog
* Fri Aug 24 2018 Ryan Goodfellow <rgoodfel@isi.edu> - 0.1.0-0
- Initial
