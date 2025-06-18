Name:		tto		
Version:	0.5.2
Release:	1%{?dist}
Summary:	tto package	

Group:		System Environment/Base	
Packager:	Craig Tomkow
License:	MIT	
URL:		https://github.com/ctomkow/tto	
Source0:	tto-0.5.2.tar.gz

Requires:	mariadb

%description
 -- tto
Three backups, two copies on different storage, one located off-site.
An asynchronous client-server app for synchronizing a MySQL database between two systems. In addition, it keeps a ring buffer of X backups on the secondary system.
The main use-case for developing this was to help maintain a hybrid primary / [primary/secondary] application deployment where replication was not possible.

 -- Configuration
The application needs to be installed on the primary and secondary systems. Each will be configured for their respective roles (sender | receiver).
Edit conf.json in /etc/tto/

%prep
%setup -q # unpack tar.gz

%install
mkdir -p %{buildroot}/usr/local/bin/
cp -rfa * %{buildroot}/usr/local/bin/

%files
%attr(0744, root, root) /usr/local/bin/*

%post
# Only run on fresh install, not upgrade
if [ "$1" -eq 1 ]; then
    /usr/local/bin/tto install || true
fi

%preun
# Only run on full removal, not upgrade
if [ "$1" -eq 0 ]; then
    /usr/local/bin/tto remove || true
fi

%postun
# Do NOT delete config or opt directory on upgrade
if [ "$1" -eq 0 ]; then
    rm -rf /opt/tto/ /etc/tto/
fi

%doc
%changelog

