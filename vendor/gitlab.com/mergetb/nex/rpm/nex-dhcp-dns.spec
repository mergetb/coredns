Name:		  nex-dhcp-dns
Version:	0.2.1
Release:	1%{?dist}
Summary:	Software defined DHCP/DNS

License:	Apache-2.0
URL:		  https://gitlab.com/mergetb/nex

BuildRequires: golang dep systemd

%description
Nex is the spirit of dnsmasq meets modern scale out architecture. Nex is 
designed around a stateless service model for both DHCP and DNS - with an 
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


%install
cd src/gitlab.com/mergetb/nex
export prefix=%{buildroot}/usr
make install
curl -L https://github.com/mergetb/coredns/releases/download/v1.2.2-nex/coredns -o %{buildroot}/usr/bin/coredns
chmod a+x %{buildroot}/usr/bin/coredns

%files
/usr/bin/coredns
/usr/bin/loader
/usr/bin/nexd
/usr/bin/nex-dhcpd
/lib/systemd/system/coredns.service
/lib/systemd/system/nexd.service
/lib/systemd/system/nex-dhcpd.service

%post
%systemd_post nex-dhcpd.service
%systemd_post nexd.service
%systemd_post coredns.service

%preun
%systemd_preun nex-dhcpd.service
%systemd_preun nexd.service
%systemd_preun coredns.service

%postun
%systemd_postun_with_restart nex-dhcpd.service
%systemd_postun_with_restart nexd.service
%systemd_postun_with_restart coredns.service


%changelog
* Mon Oct 29 2018 Ryan Goodfellow <rgoodfel@isi.edu> - 0.2.1-0
- Pure-Go

* Fri Oct 19 2018 Ryan Goodfellow <rgoodfel@isi.edu> - 0.1.1-0
- Add nexd service

* Fri Aug 24 2018 Ryan Goodfellow <rgoodfel@isi.edu> - 0.1.0-0
- Initial
